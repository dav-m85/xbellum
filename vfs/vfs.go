package vfs

import (
	"context"
	"errors"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/webdav"
)

type Store interface {
	Get() ([]byte, error)
	Set([]byte) error
}

var _ webdav.FileSystem = &VFS{}

type VFS struct {
	mu   sync.Mutex
	xbel *memFSNode
	lock *memFSNode
	set  func([]byte) error
}

// NewMemFS returns a new in-memory FileSystem implementation.
func NewVFS(r Store) webdav.FileSystem {
	vfs := &VFS{set: r.Set}
	res, err := r.Get()
	if err == nil {
		vfs.xbel = &memFSNode{
			mode: 0,
			data: res,
		}
	}

	return vfs
}

type memFSNode struct {
	mu      sync.Mutex
	data    []byte
	mode    os.FileMode
	modTime time.Time

	// children is protected by memFS.mu.
	children map[string]*memFSNode
}

func (n *memFSNode) stat(name string) *memFileInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	return &memFileInfo{
		name:    name,
		size:    int64(len(n.data)),
		mode:    n.mode,
		modTime: n.modTime,
	}
}

func (fs *VFS) root() *memFSNode {
	cn := make(map[string]*memFSNode)
	if fs.xbel != nil {
		cn["bookmarks.xbel"] = fs.xbel
	}
	if fs.lock != nil {
		cn["bookmarks.xbel.lock"] = fs.lock
	}

	return &memFSNode{
		children: cn,
		mode:     0660 | os.ModeDir,
		modTime:  time.Now(),
	}
}

// create xbel: 1538 -rw-rw-rw- O_RDWR O_CREATE O_TRUNC
// download xbel: 0 ---------- O_RDONLY

// Algo
// If read from file, always get latest version
// If write to file, wait until wrote then compare to latest
//   If latest different, record changes and create an html recap
//   If latest same, replace it (dont recap for shuffling favs)

func (fs *VFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	log.Print("OpenFile ", name)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	var frag string
	var n *memFSNode
	if name == "" {
		// We're opening the root.
		if runtime.GOOS == "zos" {
			if flag&os.O_WRONLY != 0 {
				return nil, os.ErrPermission
			}
		} else {
			if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
				return nil, os.ErrPermission
			}
		}

		n, frag = fs.root(), "/"

	} else {
		switch name {
		case "/bookmarks.xbel":
			n = fs.xbel
		case "/bookmarks.xbel.lock":
			n = fs.lock
		default:
			return nil, os.ErrInvalid
		}

		if flag&(os.O_SYNC|os.O_APPEND) != 0 {
			// memFile doesn't support these flags yet.
			return nil, os.ErrInvalid
		}
		if flag&os.O_CREATE != 0 {
			if flag&os.O_EXCL != 0 && n != nil {
				return nil, os.ErrExist
			}
			if n == nil {
				n = &memFSNode{
					mode: perm.Perm(),
				}
				switch name {
				case "/bookmarks.xbel":
					fs.xbel = n
				case "/bookmarks.xbel.lock":
					fs.lock = n
				}
			}
		}
		if n == nil {
			return nil, os.ErrNotExist
		}
		if flag&(os.O_WRONLY|os.O_RDWR) != 0 && flag&os.O_TRUNC != 0 {
			n.mu.Lock()
			n.data = nil
			n.mu.Unlock()
		}
	}

	children := make([]os.FileInfo, 0, len(n.children))
	for cName, c := range n.children {
		children = append(children, c.stat(cName))
	}
	onClose := func(*memFile) error { return nil }

	if name == "/bookmarks.xbel" {
		onClose = func(f *memFile) error {
			if f.written {
				return fs.set(f.n.data)
			}
			return nil
		}
	}

	return &memFile{
		n:                n,
		nameSnapshot:     frag,
		childrenSnapshot: children,
		onClose:          onClose,
	}, nil
}

func (fs *VFS) RemoveAll(ctx context.Context, name string) error {
	log.Print("RemoveAll ", name)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if name == "" {
		// We can't remove the root.
		return os.ErrInvalid
	}
	switch name {
	case "/bookmarks.xbel":
		fs.xbel = nil
	case "/bookmarks.xbel.lock":
		fs.lock = nil
	default:
		return os.ErrInvalid
	}
	return nil
}

func (fs *VFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Print("Stat ", name)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if name == "" {
		// We're stat'ting the root.
		return fs.root().stat("/"), nil
	}
	switch name {
	case "/bookmarks.xbel":
		if fs.xbel == nil {
			return nil, os.ErrNotExist
		}
		return fs.xbel.stat(path.Base(name)), nil
	case "/bookmarks.xbel.lock":
		if fs.lock == nil {
			return nil, os.ErrNotExist
		}
		return fs.lock.stat(path.Base(name)), nil
	}

	return nil, os.ErrNotExist
}

var ErrNotImplemented = errors.New("not implemented")

func (v *VFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Print("Mkdir ", name)
	return ErrNotImplemented
}

func (v *VFS) Rename(ctx context.Context, oldName, newName string) error {
	log.Print("Rename ", oldName)
	return ErrNotImplemented
}

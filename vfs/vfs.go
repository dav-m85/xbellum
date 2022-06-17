package vfs

import (
	"context"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = VFS{}

type VFS struct{}

func resolve(name string) string {
	// This implementation is based on Dir.Open's code in the standard net/http package.
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}

	if name != "/bookmarks.xbel" && name != "/bookmarks.xbel.lock" {
		return ""
	}

	return filepath.Join("./data", filepath.FromSlash(slashClean(name)))
}

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

func (v VFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	log.Print("OpenFile ", name)

	if name = resolve(name); name == "" {
		return nil, os.ErrNotExist
	}
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (v VFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Print("Stat ", name)
	if name = resolve(name); name == "" {
		return nil, os.ErrNotExist
	}
	return os.Stat(name)
}

var ErrNotImplemented = errors.New("not implemented")

func (v VFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Print("Mkdir ", name)
	return ErrNotImplemented
}

func (v VFS) RemoveAll(ctx context.Context, name string) error {
	log.Print("RemoveAll ", name)
	if name = resolve(name); name == "" {
		return os.ErrNotExist
	}
	// if name == filepath.Clean(string(d)) {
	// 	// Prohibit removing the virtual root directory.
	// 	return os.ErrInvalid
	// }
	return os.RemoveAll(name)
}

func (v VFS) Rename(ctx context.Context, oldName, newName string) error {
	log.Print("Rename ", oldName)
	return ErrNotImplemented
}

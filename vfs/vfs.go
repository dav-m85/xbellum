package vfs

import (
	"context"
	"errors"
	"log"
	"os"

	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = VFS{}

type VFS struct{}

func (v VFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	log.Print("OpenFile", name)
	return NewFile(), nil
}

func (v VFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Print("Stat", name)
	return FileInfo{name}, nil
}

var ErrNotImplemented = errors.New("not implemented")

func (v VFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Print("Mkdir", name)
	return ErrNotImplemented
}

func (v VFS) RemoveAll(ctx context.Context, name string) error {
	log.Print("RemoveAll", name)
	return ErrNotImplemented
}
func (v VFS) Rename(ctx context.Context, oldName, newName string) error {
	log.Print("Rename", oldName)
	return ErrNotImplemented
}

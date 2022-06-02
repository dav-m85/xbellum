package vfs

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = VFS{}

type VFS struct{}

var ErrNotImplemented = errors.New("not implemented")

func (v VFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	log.Print("Mkdir", name)
	return ErrNotImplemented
}
func (v VFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	log.Print("OpenFile", name)
	return NewFile(), nil
}
func (v VFS) RemoveAll(ctx context.Context, name string) error {
	log.Print("RemoveAll", name)
	return ErrNotImplemented
}
func (v VFS) Rename(ctx context.Context, oldName, newName string) error {
	log.Print("Rename", oldName)
	return ErrNotImplemented
}

func (v VFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Print("Stat", name)

	return FileInfo{name}, nil
}

var _ webdav.File = &File{}

func NewFile() *File {
	return &File{bytes.NewBuffer([]byte{})}
}

type File struct {
	*bytes.Buffer
}

func (f *File) Close() error {
	return nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *File) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}
func (f *File) Stat() (fs.FileInfo, error) {
	return FileInfo{"/"}, nil
}

var _ os.FileInfo = FileInfo{}

type FileInfo struct {
	name string
}

func (fi FileInfo) Name() string {
	return fi.name
}
func (fi FileInfo) Size() int64 {
	return 0
}
func (fi FileInfo) Mode() os.FileMode {
	return 0
}
func (fi FileInfo) ModTime() time.Time {
	return time.Now()
}
func (fi FileInfo) IsDir() bool {
	return true
}
func (fi FileInfo) Sys() interface{} {
	return nil
}

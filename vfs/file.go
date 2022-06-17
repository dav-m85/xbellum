package vfs

import (
	"bytes"
	"io/fs"

	"golang.org/x/net/webdav"
)

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

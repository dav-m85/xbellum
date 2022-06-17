package vfs

import (
	"os"
	"time"
)

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

package ui

import (
	"os"
	"time"
)

type myFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi myFileInfo) Name() string {
	return fi.name
}

func (fi myFileInfo) Size() int64 {
	return fi.size
}

func (fi myFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi myFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi myFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi myFileInfo) Sys() interface{} {
	return nil
}

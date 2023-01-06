package reposaur

import (
	"github.com/go-git/go-billy/v5"
	"io/fs"
)

// Custom implementation of fs.FS for billy.Filesystem
type billyFS struct {
	billy.Filesystem
}

func (bfs billyFS) Open(name string) (fs.File, error) {
	f, err := bfs.Filesystem.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := bfs.Stat(name)
	if err != nil {
		return nil, err
	}

	return &billyFile{File: f, stat: stat}, nil
}

func (bfs billyFS) ReadDir(path string) ([]fs.DirEntry, error) {
	files, err := bfs.Filesystem.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dirEntries := make([]fs.DirEntry, len(files))
	for i, f := range files {
		dirEntries[i] = billyDirEntry{FileInfo: f}
	}

	return dirEntries, nil
}

// Custom implementation of fs.File for billy.File
type billyFile struct {
	billy.File
	stat fs.FileInfo
}

func (bf billyFile) Stat() (fs.FileInfo, error) {
	return bf.stat, nil
}

// Custom implementation of fs.DirEntry for fs.FileInfo
type billyDirEntry struct {
	fs.FileInfo
}

func (bde billyDirEntry) Type() fs.FileMode {
	return bde.FileInfo.Mode()
}

func (bde billyDirEntry) Info() (fs.FileInfo, error) {
	return bde.FileInfo, nil
}

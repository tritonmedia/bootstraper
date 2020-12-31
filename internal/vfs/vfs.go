package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5"
)

// Compile time assertion
var _ billy.Filesystem = &MergedFS{}

type MergedFS struct {
	filesystems []billy.Filesystem
}

// NewMergedFilesystem creates a "layered" file-system where
// higher indexed file systems are searched first, and then down, returning
// an error if not found
func NewMergedFilesystem(filesystems ...billy.Filesystem) *MergedFS {
	return &MergedFS{
		filesystems: filesystems,
	}
}

// findFile finds a file across all filesystems, and returns a os.ErrNotExist
// if it's not found, and returns the filesystem it belongs to if it's found
func (m *MergedFS) findFile(path string) (billy.Filesystem, error) {
	for _, fs := range m.filesystems {
		if _, err := fs.Stat(path); err == nil {
			return fs, nil
		}
	}

	return nil, os.ErrNotExist
}

func (m *MergedFS) Create(path string) (billy.File, error) {
	return nil, fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Open(path string) (billy.File, error) {
	fs, err := m.findFile(path)
	if err != nil {
		return nil, err
	}

	return fs.Open(path)
}

func (m *MergedFS) OpenFile(path string, flag int, perm os.FileMode) (billy.File, error) {
	return nil, fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Stat(path string) (os.FileInfo, error) {
	fs, err := m.findFile(path)
	if err != nil {
		return nil, err
	}

	return fs.Stat(path)
}

func (m *MergedFS) Rename(oldPath, newPath string) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Remove(path string) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (m *MergedFS) TempFile(dir, prefix string) (billy.File, error) {
	return nil, fmt.Errorf("unsupported on merged filesystem")
}

// ReadDir will read a directory across all available filesystems, and deduplicate
func (m *MergedFS) ReadDir(dir string) ([]os.FileInfo, error) {
	files := make(map[string]os.FileInfo)

	foundDir := true
	for _, fs := range m.filesystems {
		filelist, err := fs.ReadDir(dir)
		if err != nil {
			// TODO(jaredallard): better error handling here
			continue
		}

		// if we had no error at one point, we set foundDir
		// to make sure we return the correct error type across
		// all of the filesystems
		foundDir = true
		for _, f := range filelist {
			files[f.Name()] = f
		}
	}

	// if we didn't find it at, assume ErrNotExist was the reason
	if !foundDir {
		return nil, os.ErrNotExist
	}

	// convert our de-duplication hash map back into a list
	list := make([]os.FileInfo, len(files))
	i := 0
	for _, f := range files {
		list[i] = f
		i++
	}

	return list, nil
}

func (m *MergedFS) MkdirAll(dir string, perm os.FileMode) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

///
// Symlink
///

func (m *MergedFS) Lstat(path string) (os.FileInfo, error) {
	fs, err := m.findFile(path)
	if err != nil {
		return nil, err
	}

	return fs.Lstat(path)
}

func (m *MergedFS) Symlink(target, link string) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Readlink(path string) (string, error) {
	fs, err := m.findFile(path)
	if err != nil {
		return "", err
	}

	return fs.Readlink(path)
}

///
// Chmod
///

func (m *MergedFS) Chmod(name string, mode os.FileMode) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Lchown(name string, uid, gid int) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Chown(name string, uid, gid int) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

func (m *MergedFS) Chtimes(name string, atime, mtime time.Time) error {
	return fmt.Errorf("unsupported on merged filesystem")
}

///
// Chroot
///

func (m *MergedFS) Chroot(path string) (billy.Filesystem, error) {
	fs, err := m.findFile(path)
	if err != nil {
		return nil, err
	}

	return fs.Chroot(path)
}

func (m *MergedFS) Root() string {
	// TODO(jaredallard): figure this one out
	return ""
}

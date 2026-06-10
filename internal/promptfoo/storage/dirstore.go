package storage

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DirStore is a Store backed by a directory on the local filesystem. Reads go
// through os.DirFS; writes create the directory (and any parents within it) as
// needed. It is the default backing for pe's cache and run-result stores.
type DirStore struct {
	root string
	fsys fs.FS
}

// NewDirStore returns a Store rooted at dir. The directory is created lazily on
// the first write, so constructing a DirStore never touches the filesystem.
func NewDirStore(dir string) *DirStore {
	return &DirStore{root: dir, fsys: os.DirFS(dir)}
}

// Open implements fs.FS.
func (d *DirStore) Open(name string) (fs.File, error) {
	return d.fsys.Open(name)
}

// ReadDir implements fs.ReadDirFS so listing works even before any write has
// created the directory (a missing root reads as empty).
func (d *DirStore) ReadDir(name string) ([]fs.DirEntry, error) {
	full := d.root
	if name != "." {
		full = filepath.Join(d.root, filepath.FromSlash(name))
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return entries, nil
}

// Put writes data to <root>/<name>, creating parent directories as needed.
func (d *DirStore) Put(name string, data []byte) error {
	path := filepath.Join(d.root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Remove deletes <root>/<name>. Removing a missing name is not an error.
func (d *DirStore) Remove(name string) error {
	path := filepath.Join(d.root, filepath.FromSlash(name))
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

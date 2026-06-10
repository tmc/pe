package storage

import (
	"bytes"
	"io"
	"io/fs"
	"sort"
	"sync"
	"time"
)

// MemStore is an in-memory Store. It is safe for concurrent use and is handy
// for tests and ephemeral runs, and demonstrates that the Store interface is
// not tied to the filesystem.
type MemStore struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore {
	return &MemStore{files: make(map[string][]byte)}
}

// Open implements fs.FS.
func (m *MemStore) Open(name string) (fs.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{name: name, r: bytes.NewReader(data), size: int64(len(data))}, nil
}

// ReadDir implements fs.ReadDirFS for the flat top-level namespace MemStore
// uses (cache and run stores store flat "<key>.json" names).
func (m *MemStore) ReadDir(name string) ([]fs.DirEntry, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	entries := make([]fs.DirEntry, 0, len(m.files))
	for n, data := range m.files {
		entries = append(entries, memDirEntry{name: n, size: int64(len(data))})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

// Put stores data under name.
func (m *MemStore) Put(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	clone := make([]byte, len(data))
	copy(clone, data)
	m.files[name] = clone
	return nil
}

// Remove deletes name. Removing a missing name is not an error.
func (m *MemStore) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, name)
	return nil
}

type memFile struct {
	name string
	r    *bytes.Reader
	size int64
}

func (f *memFile) Stat() (fs.FileInfo, error) { return memDirEntry{name: f.name, size: f.size}, nil }
func (f *memFile) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f *memFile) Close() error               { return nil }

var _ io.Reader = (*memFile)(nil)

type memDirEntry struct {
	name string
	size int64
}

func (e memDirEntry) Name() string               { return e.name }
func (e memDirEntry) IsDir() bool                { return false }
func (e memDirEntry) Type() fs.FileMode          { return 0 }
func (e memDirEntry) Info() (fs.FileInfo, error) { return e, nil }
func (e memDirEntry) Size() int64                { return e.size }
func (e memDirEntry) Mode() fs.FileMode          { return 0 }
func (e memDirEntry) ModTime() time.Time         { return time.Time{} }
func (e memDirEntry) Sys() any                   { return nil }

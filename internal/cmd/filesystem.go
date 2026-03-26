package cmd

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// FileSystem abstracts file operations for testing.
// This interface allows commands to be tested without touching the real filesystem.
type FileSystem interface {
	// ReadFile reads a file and returns its contents.
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file with the given permissions.
	WriteFile(path string, data []byte, perm fs.FileMode) error

	// Stat returns file info.
	Stat(path string) (fs.FileInfo, error)

	// MkdirAll creates a directory and all parent directories.
	MkdirAll(path string, perm fs.FileMode) error

	// Remove removes a file or empty directory.
	Remove(path string) error

	// RemoveAll removes a path and any children it contains.
	RemoveAll(path string) error

	// Open opens a file for reading.
	Open(path string) (io.ReadCloser, error)

	// Create creates or truncates a file for writing.
	Create(path string) (io.WriteCloser, error)

	// Exists checks if a path exists.
	Exists(path string) bool

	// IsDir checks if a path is a directory.
	IsDir(path string) bool

	// Glob returns file paths matching a pattern.
	Glob(pattern string) ([]string, error)

	// Walk walks a directory tree.
	Walk(root string, fn filepath.WalkFunc) error
}

// OSFileSystem implements FileSystem using the real OS filesystem.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS filesystem.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads a file.
func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file.
func (f *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Stat returns file info.
func (f *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// MkdirAll creates a directory and parents.
func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Remove removes a file.
func (f *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a path and children.
func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Open opens a file for reading.
func (f *OSFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Create creates a file for writing.
func (f *OSFileSystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

// Exists checks if a path exists.
func (f *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory.
func (f *OSFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Glob returns matching paths.
func (f *OSFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Walk walks a directory tree.
func (f *OSFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// MemoryFileSystem implements FileSystem using an in-memory map.
// Use this for testing to avoid touching the real filesystem.
type MemoryFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

// NewMemoryFileSystem creates a new in-memory filesystem.
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// ReadFile reads from memory.
func (m *MemoryFileSystem) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// WriteFile writes to memory.
func (m *MemoryFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	m.files[path] = data
	return nil
}

// Stat returns fake file info.
func (m *MemoryFileSystem) Stat(path string) (fs.FileInfo, error) {
	if m.dirs[path] {
		return &memFileInfo{name: filepath.Base(path), isDir: true}, nil
	}
	if _, ok := m.files[path]; ok {
		return &memFileInfo{name: filepath.Base(path), size: int64(len(m.files[path]))}, nil
	}
	return nil, os.ErrNotExist
}

// MkdirAll marks directories as existing.
func (m *MemoryFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	m.dirs[path] = true
	return nil
}

// Remove removes from memory.
func (m *MemoryFileSystem) Remove(path string) error {
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

// RemoveAll removes path and children.
func (m *MemoryFileSystem) RemoveAll(path string) error {
	for k := range m.files {
		if k == path || hasPrefix(k, path+"/") {
			delete(m.files, k)
		}
	}
	for k := range m.dirs {
		if k == path || hasPrefix(k, path+"/") {
			delete(m.dirs, k)
		}
	}
	return nil
}

// Open opens for reading.
func (m *MemoryFileSystem) Open(path string) (io.ReadCloser, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &memReadCloser{data: data}, nil
}

// Create creates for writing.
func (m *MemoryFileSystem) Create(path string) (io.WriteCloser, error) {
	return &memWriteCloser{fs: m, path: path}, nil
}

// Exists checks if path exists.
func (m *MemoryFileSystem) Exists(path string) bool {
	if m.dirs[path] {
		return true
	}
	_, ok := m.files[path]
	return ok
}

// IsDir checks if path is a directory.
func (m *MemoryFileSystem) IsDir(path string) bool {
	return m.dirs[path]
}

// Glob is not fully implemented for memory filesystem.
func (m *MemoryFileSystem) Glob(pattern string) ([]string, error) {
	var matches []string
	for path := range m.files {
		if matched, _ := filepath.Match(pattern, path); matched {
			matches = append(matches, path)
		}
	}
	return matches, nil
}

// Walk walks the in-memory tree.
func (m *MemoryFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	// Walk directories first
	for path := range m.dirs {
		if path == root || hasPrefix(path, root+"/") {
			info := &memFileInfo{name: filepath.Base(path), isDir: true}
			if err := fn(path, info, nil); err != nil {
				return err
			}
		}
	}
	// Then files
	for path := range m.files {
		if hasPrefix(path, root+"/") || path == root {
			info := &memFileInfo{name: filepath.Base(path), size: int64(len(m.files[path]))}
			if err := fn(path, info, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddFile adds a file to the memory filesystem (test helper).
func (m *MemoryFileSystem) AddFile(path string, content string) {
	m.files[path] = []byte(content)
}

// AddDir adds a directory to the memory filesystem (test helper).
func (m *MemoryFileSystem) AddDir(path string) {
	m.dirs[path] = true
}

// memFileInfo implements fs.FileInfo for memory files.
type memFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i *memFileInfo) Name() string       { return i.name }
func (i *memFileInfo) Size() int64        { return i.size }
func (i *memFileInfo) Mode() fs.FileMode  { return 0644 }
func (i *memFileInfo) ModTime() time.Time { return time.Time{} }
func (i *memFileInfo) IsDir() bool        { return i.isDir }
func (i *memFileInfo) Sys() interface{}   { return nil }

// memReadCloser implements io.ReadCloser for memory files.
type memReadCloser struct {
	data []byte
	pos  int
}

func (r *memReadCloser) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *memReadCloser) Close() error { return nil }

// memWriteCloser implements io.WriteCloser for memory files.
type memWriteCloser struct {
	fs   *MemoryFileSystem
	path string
	data []byte
}

func (w *memWriteCloser) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *memWriteCloser) Close() error {
	w.fs.files[w.path] = w.data
	return nil
}

// hasPrefix checks if s starts with prefix.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

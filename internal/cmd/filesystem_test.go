package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileSystemInterface validates that our implementations satisfy FileSystem.
// REQ-GO-062: Go CLI shall provide FileSystem interface for command testing
func TestFileSystemInterface(t *testing.T) {
	// OSFileSystem must satisfy FileSystem
	var _ FileSystem = &OSFileSystem{}

	// MemoryFileSystem must satisfy FileSystem
	var _ FileSystem = &MemoryFileSystem{}
}

func TestOSFileSystem(t *testing.T) {
	fs := NewOSFileSystem()

	// Create temp dir
	dir, err := os.MkdirTemp("", "fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	testPath := filepath.Join(dir, "test.txt")

	// Test WriteFile and ReadFile
	content := []byte("test content")
	if err := fs.WriteFile(testPath, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := fs.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", data, content)
	}

	// Test Exists
	if !fs.Exists(testPath) {
		t.Error("Exists returned false for existing file")
	}

	// Test Stat
	info, err := fs.Stat(testPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("Size mismatch: got %d, want %d", info.Size(), len(content))
	}

	// Test IsDir
	if fs.IsDir(testPath) {
		t.Error("IsDir returned true for file")
	}
	if !fs.IsDir(dir) {
		t.Error("IsDir returned false for directory")
	}

	// Test Remove
	if err := fs.Remove(testPath); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if fs.Exists(testPath) {
		t.Error("Exists returned true after Remove")
	}
}

func TestOSFileSystemMkdirAll(t *testing.T) {
	fs := NewOSFileSystem()

	dir, err := os.MkdirTemp("", "fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	nested := filepath.Join(dir, "a", "b", "c")
	if err := fs.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if !fs.IsDir(nested) {
		t.Error("Nested directory was not created")
	}
}

func TestOSFileSystemOpenCreate(t *testing.T) {
	fs := NewOSFileSystem()

	dir, err := os.MkdirTemp("", "fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	testPath := filepath.Join(dir, "stream.txt")

	// Test Create
	writer, err := fs.Create(testPath)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_, _ = writer.Write([]byte("streamed content"))
	writer.Close()

	// Test Open
	reader, err := fs.Open(testPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer reader.Close()

	buf := make([]byte, 100)
	n, _ := reader.Read(buf)
	if string(buf[:n]) != "streamed content" {
		t.Errorf("Content mismatch: got %q", buf[:n])
	}
}

func TestMemoryFileSystem(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Test WriteFile and ReadFile
	path := "/test/file.txt"
	content := []byte("memory content")

	if err := fs.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Content mismatch: got %q, want %q", data, content)
	}

	// Test Exists
	if !fs.Exists(path) {
		t.Error("Exists returned false for existing file")
	}

	// Test Stat
	info, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("Size mismatch: got %d, want %d", info.Size(), len(content))
	}

	// Test Remove
	if err := fs.Remove(path); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if fs.Exists(path) {
		t.Error("Exists returned true after Remove")
	}
}

func TestMemoryFileSystemMkdirAll(t *testing.T) {
	fs := NewMemoryFileSystem()

	path := "/a/b/c"
	if err := fs.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if !fs.IsDir(path) {
		t.Error("IsDir returned false for created directory")
	}
}

func TestMemoryFileSystemOpenCreate(t *testing.T) {
	fs := NewMemoryFileSystem()

	path := "/stream.txt"

	// Test Create
	writer, err := fs.Create(path)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_, _ = writer.Write([]byte("part1"))
	_, _ = writer.Write([]byte("part2"))
	writer.Close()

	// Test Open
	reader, err := fs.Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer reader.Close()

	buf := make([]byte, 100)
	n, _ := reader.Read(buf)
	if string(buf[:n]) != "part1part2" {
		t.Errorf("Content mismatch: got %q", buf[:n])
	}
}

func TestMemoryFileSystemRemoveAll(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Create nested structure
	fs.AddDir("/root")
	fs.AddDir("/root/a")
	fs.AddFile("/root/file1.txt", "content1")
	fs.AddFile("/root/a/file2.txt", "content2")

	// Remove all
	if err := fs.RemoveAll("/root"); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	if fs.Exists("/root") {
		t.Error("Root directory still exists")
	}
	if fs.Exists("/root/a") {
		t.Error("Nested directory still exists")
	}
	if fs.Exists("/root/file1.txt") {
		t.Error("Root file still exists")
	}
	if fs.Exists("/root/a/file2.txt") {
		t.Error("Nested file still exists")
	}
}

func TestMemoryFileSystemGlob(t *testing.T) {
	fs := NewMemoryFileSystem()

	fs.AddFile("/test/a.txt", "a")
	fs.AddFile("/test/b.txt", "b")
	fs.AddFile("/test/c.go", "c")

	matches, err := fs.Glob("/test/*.txt")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d: %v", len(matches), matches)
	}
}

func TestMemoryFileSystemAddHelpers(t *testing.T) {
	fs := NewMemoryFileSystem()

	fs.AddFile("/test.txt", "hello")
	fs.AddDir("/mydir")

	if !fs.Exists("/test.txt") {
		t.Error("AddFile did not create file")
	}

	if !fs.IsDir("/mydir") {
		t.Error("AddDir did not create directory")
	}

	data, _ := fs.ReadFile("/test.txt")
	if string(data) != "hello" {
		t.Errorf("Unexpected content: %s", data)
	}
}

func TestMemoryFileSystemReadNonExistent(t *testing.T) {
	fs := NewMemoryFileSystem()

	_, err := fs.ReadFile("/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMemoryFileSystemStatNonExistent(t *testing.T) {
	fs := NewMemoryFileSystem()

	_, err := fs.Stat("/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMemoryFileSystemOpenNonExistent(t *testing.T) {
	fs := NewMemoryFileSystem()

	_, err := fs.Open("/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

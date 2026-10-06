// Package utils provides content-addressed, zlib-compressed object storage and
// shared filesystem helpers used across the MicroGit commands.
package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

const (
	DEFAULT_PATH = ".microgit"
)

type SavePoint struct {
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Parent    string            `json:"parent"`
	Files     map[string]string `json:"files"`
}

// HashContent returns the SHA-256 hash of the file content
func HashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// ObjectPath returns the on-disk path for a stored object (blob or commit).
// Objects are sharded by the first two hex characters of their hash
// (objects/ab/cdef...) so a single directory never holds every object, the
// same layout git uses.
func ObjectPath(hash string) string {
	if len(hash) < 2 {
		return filepath.Join(DEFAULT_PATH, "objects", hash)
	}
	return filepath.Join(DEFAULT_PATH, "objects", hash[:2], hash[2:])
}

// IsInitialized reports whether the current directory is a MicroGit repository.
func IsInitialized() bool {
	_, err := os.Stat(DEFAULT_PATH)
	return err == nil
}

// WriteObject stores zlib-compressed content at the sharded object path.
// Objects are content-addressed and immutable, so an existing object is left
// untouched.
func WriteObject(hash string, content []byte) error {
	path := ObjectPath(hash)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil // already stored
	}

	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(content); err != nil {
		zw.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	return WriteFileAtomic(path, buf.Bytes())
}

// ReadObject reads and decompresses the object stored at the given hash.
func ReadObject(hash string) ([]byte, error) {
	data, err := os.ReadFile(ObjectPath(hash))
	if err != nil {
		return nil, err
	}

	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	return io.ReadAll(zr)
}

// WriteFileAtomic writes data to path via a temporary file in the same
// directory followed by a rename, so a crash mid-write can never leave a
// partially-written (corrupt) file at path.
func WriteFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".microgit-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, path)
}

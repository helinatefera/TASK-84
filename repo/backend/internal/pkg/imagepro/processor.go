package imagepro

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var allowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// Magic bytes for format sniffing
var magicBytes = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8, 0xFF},
	"image/png":  {0x89, 0x50, 0x4E, 0x47},
	"image/webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header
}

type ProcessResult struct {
	SHA256Hash  string
	MimeType    string
	FileSize    int64
	StoragePath string
	IsSuspicious bool
	SuspiciousReason string
}

func ProcessUpload(data []byte, imagesDir string, maxSize int64) (*ProcessResult, error) {
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file exceeds maximum size of %d bytes", maxSize)
	}

	// Detect MIME type via magic bytes (format sniffing)
	detectedMIME := http.DetectContentType(data)
	if !allowedMIME[detectedMIME] {
		// Try manual magic byte check
		detectedMIME = sniffMagicBytes(data)
		if detectedMIME == "" {
			return nil, fmt.Errorf("unsupported image format: %s", http.DetectContentType(data))
		}
	}

	// Verify magic bytes match claimed type
	if expected, ok := magicBytes[detectedMIME]; ok {
		if len(data) < len(expected) {
			return nil, fmt.Errorf("file too small to be a valid image")
		}
		if !bytes.HasPrefix(data, expected) {
			return nil, fmt.Errorf("file content does not match declared format")
		}
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Build storage path with 2-level hash prefix
	dir1 := hashStr[:2]
	dir2 := hashStr[2:4]
	storagePath := filepath.Join(dir1, dir2, hashStr)
	fullPath := filepath.Join(imagesDir, storagePath)

	// Check for suspicious content
	suspicious := false
	suspiciousReason := ""
	if containsScript(data) {
		suspicious = true
		suspiciousReason = "File contains script-like content"
	}

	// Create directory structure
	dirPath := filepath.Dir(fullPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write image file: %w", err)
	}

	return &ProcessResult{
		SHA256Hash:       hashStr,
		MimeType:         detectedMIME,
		FileSize:         int64(len(data)),
		StoragePath:      storagePath,
		IsSuspicious:     suspicious,
		SuspiciousReason: suspiciousReason,
	}, nil
}

func MoveToQuarantine(imagesDir, quarantineDir, storagePath string) error {
	src := filepath.Join(imagesDir, storagePath)
	dst := filepath.Join(quarantineDir, storagePath)

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create quarantine directory: %w", err)
	}

	return os.Rename(src, dst)
}

func MoveFromQuarantine(quarantineDir, imagesDir, storagePath string) error {
	src := filepath.Join(quarantineDir, storagePath)
	dst := filepath.Join(imagesDir, storagePath)

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create images directory: %w", err)
	}

	return os.Rename(src, dst)
}

func DeleteFile(baseDir, storagePath string) error {
	return os.Remove(filepath.Join(baseDir, storagePath))
}

func ComputeHash(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sniffMagicBytes(data []byte) string {
	for mime, magic := range magicBytes {
		if len(data) >= len(magic) && bytes.HasPrefix(data, magic) {
			return mime
		}
	}
	return ""
}

func containsScript(data []byte) bool {
	lower := bytes.ToLower(data)
	patterns := [][]byte{
		[]byte("<script"),
		[]byte("javascript:"),
		[]byte("onerror="),
		[]byte("onload="),
	}
	for _, p := range patterns {
		if bytes.Contains(lower, p) {
			return true
		}
	}
	return false
}

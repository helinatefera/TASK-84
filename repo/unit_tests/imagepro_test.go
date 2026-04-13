package unit_tests_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/localinsights/portal/internal/pkg/imagepro"
)

func TestProcessValidJPEG(t *testing.T) {
	tmpDir := t.TempDir()

	// Minimal JPEG: FF D8 FF E0 header + enough padding to be detected
	jpegData := make([]byte, 256)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	jpegData[3] = 0xE0
	// Fill the rest with benign bytes
	for i := 4; i < len(jpegData); i++ {
		jpegData[i] = 0x00
	}

	result, err := imagepro.ProcessUpload(jpegData, tmpDir, 1024*1024)
	if err != nil {
		t.Fatalf("ProcessUpload returned error: %v", err)
	}
	if result.SHA256Hash == "" {
		t.Fatal("expected non-empty SHA256Hash")
	}
	if result.MimeType != "image/jpeg" {
		t.Errorf("expected MimeType image/jpeg, got %s", result.MimeType)
	}
	if result.StoragePath == "" {
		t.Fatal("expected non-empty StoragePath")
	}

	// Verify the file was actually written
	fullPath := filepath.Join(tmpDir, result.StoragePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("expected file to exist at %s", fullPath)
	}
}

func TestProcessValidPNG(t *testing.T) {
	tmpDir := t.TempDir()

	// Minimal PNG header: 89 50 4E 47 0D 0A 1A 0A + minimal IHDR chunk
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length (13 bytes)
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08,                   // bit depth: 8
		0x02,                   // color type: RGB
		0x00, 0x00, 0x00,       // compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
	}
	// Pad to ensure net/http can detect the content type
	padding := make([]byte, 256)
	pngData = append(pngData, padding...)

	result, err := imagepro.ProcessUpload(pngData, tmpDir, 1024*1024)
	if err != nil {
		t.Fatalf("ProcessUpload returned error: %v", err)
	}
	if result.SHA256Hash == "" {
		t.Fatal("expected non-empty SHA256Hash")
	}
	if result.MimeType != "image/png" {
		t.Errorf("expected MimeType image/png, got %s", result.MimeType)
	}
}

func TestRejectInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()

	textData := []byte("This is a plain text file, not an image.")
	_, err := imagepro.ProcessUpload(textData, tmpDir, 1024*1024)
	if err == nil {
		t.Fatal("ProcessUpload should reject non-image data")
	}
}

func TestRejectOversizedFile(t *testing.T) {
	tmpDir := t.TempDir()
	maxSize := int64(1024) // 1KB limit

	// Create data larger than maxSize with JPEG header
	oversized := make([]byte, 2048)
	oversized[0] = 0xFF
	oversized[1] = 0xD8
	oversized[2] = 0xFF
	oversized[3] = 0xE0

	_, err := imagepro.ProcessUpload(oversized, tmpDir, maxSize)
	if err == nil {
		t.Fatal("ProcessUpload should reject files exceeding maxSize")
	}
}

func TestDetectSuspiciousContent(t *testing.T) {
	tmpDir := t.TempDir()

	// JPEG header with embedded script tag
	jpegData := make([]byte, 512)
	jpegData[0] = 0xFF
	jpegData[1] = 0xD8
	jpegData[2] = 0xFF
	jpegData[3] = 0xE0
	// Embed a script tag in the data
	scriptTag := []byte("<script>alert('xss')</script>")
	copy(jpegData[100:], scriptTag)

	result, err := imagepro.ProcessUpload(jpegData, tmpDir, 1024*1024)
	if err != nil {
		t.Fatalf("ProcessUpload returned error: %v", err)
	}
	if !result.IsSuspicious {
		t.Fatal("expected IsSuspicious to be true for image containing <script")
	}
	if result.SuspiciousReason == "" {
		t.Fatal("expected non-empty SuspiciousReason")
	}
}

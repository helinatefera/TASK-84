package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/pkg/hash"
)

func TestHashPassword(t *testing.T) {
	t.Run("hash is not empty", func(t *testing.T) {
		hashed, err := hash.HashPassword("mysecretpassword")
		if err != nil {
			t.Fatalf("HashPassword returned error: %v", err)
		}
		if hashed == "" {
			t.Fatal("HashPassword returned empty string")
		}
	})

	t.Run("hash is not equal to plaintext", func(t *testing.T) {
		password := "mysecretpassword"
		hashed, err := hash.HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword returned error: %v", err)
		}
		if hashed == password {
			t.Fatal("HashPassword returned the plaintext password")
		}
	})
}

func TestVerifyPassword(t *testing.T) {
	password := "correctpassword"
	hashed, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	ok, err := hash.VerifyPassword(password, hashed)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !ok {
		t.Fatal("VerifyPassword should return true for the correct password")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	hashed, err := hash.HashPassword("correct")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	ok, err := hash.VerifyPassword("wrong", hashed)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if ok {
		t.Fatal("VerifyPassword should return false for an incorrect password")
	}
}

func TestHashUniqueness(t *testing.T) {
	password := "samepassword"
	hash1, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword returned error: %v", err)
	}

	hash2, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("two hashes of the same password should differ due to different salts")
	}
}

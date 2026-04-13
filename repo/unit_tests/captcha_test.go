package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/pkg/captcha"
)

func TestGenerateCaptcha(t *testing.T) {
	store := captcha.NewStore()
	challenge, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if challenge.ID == "" {
		t.Fatal("Generate returned empty ID")
	}
	if challenge.Image == "" {
		t.Fatal("Generate returned empty Image")
	}
}

func TestVerifyCorrectAnswer(t *testing.T) {
	store := captcha.NewStore()
	challenge, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	ok := store.Verify(challenge.ID, challenge.Answer)
	if !ok {
		t.Fatal("Verify should return true for the correct answer")
	}
}

func TestVerifyWrongAnswer(t *testing.T) {
	store := captcha.NewStore()
	challenge, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	ok := store.Verify(challenge.ID, "99999")
	if ok {
		t.Fatal("Verify should return false for a wrong answer")
	}
}

func TestVerifyConsumesChallenge(t *testing.T) {
	store := captcha.NewStore()
	challenge, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	// First verification should succeed
	ok := store.Verify(challenge.ID, challenge.Answer)
	if !ok {
		t.Fatal("first Verify should return true for the correct answer")
	}

	// Second verification with the same ID should fail (single-use)
	ok = store.Verify(challenge.ID, challenge.Answer)
	if ok {
		t.Fatal("second Verify should return false because the challenge was consumed")
	}
}

func TestVerifyNonexistentID(t *testing.T) {
	store := captcha.NewStore()
	ok := store.Verify("nonexistent-id-abc123", "42")
	if ok {
		t.Fatal("Verify should return false for a nonexistent challenge ID")
	}
}

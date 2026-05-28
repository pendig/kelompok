package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUpdateProfileValidationUsesSentinelErrors(t *testing.T) {
	repository := NewRepository(nil)

	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: "   "}); !errors.Is(err, ErrProfileNameRequired) {
		t.Fatalf("expected ErrProfileNameRequired, got %v", err)
	}

	longName := strings.Repeat("a", 121)
	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: longName}); !errors.Is(err, ErrProfileNameTooLong) {
		t.Fatalf("expected ErrProfileNameTooLong, got %v", err)
	}
}

func TestUpdateProfileNameLimitCountsUTF8Runes(t *testing.T) {
	repository := NewRepository(nil)
	name := strings.Repeat("界", 121)

	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: name}); !errors.Is(err, ErrProfileNameTooLong) {
		t.Fatalf("expected ErrProfileNameTooLong for 121 runes, got %v", err)
	}
}

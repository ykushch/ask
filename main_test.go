package main

import (
	"os"
	"testing"
)

func TestGetEnvDefault(t *testing.T) {
	// Test env set
	os.Setenv("TEST_ASK_VAR", "custom_value")
	defer os.Unsetenv("TEST_ASK_VAR")
	got := getEnvDefault("TEST_ASK_VAR", "default")
	if got != "custom_value" {
		t.Errorf("getEnvDefault() with env set = %q, want %q", got, "custom_value")
	}

	// Test env unset
	os.Unsetenv("TEST_ASK_VAR_UNSET")
	got = getEnvDefault("TEST_ASK_VAR_UNSET", "fallback")
	if got != "fallback" {
		t.Errorf("getEnvDefault() with env unset = %q, want %q", got, "fallback")
	}
}

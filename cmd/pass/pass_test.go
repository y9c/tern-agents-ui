package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindRealPathSkipsSelf pins that the shim never execs itself: it must
// pick a real tool on PATH even when the shim's own directory comes first.
func TestFindRealPathSkipsSelf(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	shim := filepath.Join(first, "mytool")
	real := filepath.Join(second, "mytool")
	for _, p := range []string{shim, real} {
		if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", first+":"+second)

	if got := findRealPath("mytool", realPath(shim)); got != real {
		t.Fatalf("findRealPath = %q, want real %q", got, real)
	}
}

// TestFindRealPathEnvOverride pins that TOOL_REAL wins over PATH.
func TestFindRealPathEnvOverride(t *testing.T) {
	t.Setenv("MYTOOL_REAL", "/opt/real/mytool")
	if got := findRealPath("mytool", ""); got != "/opt/real/mytool" {
		t.Fatalf("env override not honored: %q", got)
	}
}

// TestFindRealPathNoMatch reports empty rather than execing itself.
func TestFindRealPathNoMatch(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if got := findRealPath("does-not-exist", ""); got != "" {
		t.Fatalf("want empty for no match, got %q", got)
	}
}

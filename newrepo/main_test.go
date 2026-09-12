// SPDX-FileCopyrightText: 2026 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetLogLevelFromEnv(t *testing.T) {
	tests := []struct {
		env      string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("LOG_LEVEL=%q", tt.env), func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tt.env)
			got := getLogLevelFromEnv()
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReplaceStringInFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("replaces matching string", func(t *testing.T) {
		path := filepath.Join(dir, "replace.txt")
		if err := os.WriteFile(path, []byte("hello tpl world"), 0644); err != nil {
			t.Fatal(err)
		}

		err := replaceStringInFile(path, "tpl", "myrepo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, _ := os.ReadFile(path)
		if string(got) != "hello myrepo world" {
			t.Errorf("got %q, want %q", string(got), "hello myrepo world")
		}
	})

	t.Run("no match leaves file unchanged", func(t *testing.T) {
		path := filepath.Join(dir, "noop.txt")
		content := "nothing to replace here"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		err := replaceStringInFile(path, "missing", "something")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, _ := os.ReadFile(path)
		if string(got) != content {
			t.Errorf("file was modified when it should not have been")
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		err := replaceStringInFile(filepath.Join(dir, "nonexistent.txt"), "a", "b")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})

	t.Run("replaces all occurrences", func(t *testing.T) {
		// Matches how the tpl README has "tpl" in multiple places
		path := filepath.Join(dir, "multi.txt")
		if err := os.WriteFile(path, []byte("tpl is tpl and tpl"), 0644); err != nil {
			t.Fatal(err)
		}

		err := replaceStringInFile(path, "tpl", "myrepo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, _ := os.ReadFile(path)
		if string(got) != "myrepo is myrepo and myrepo" {
			t.Errorf("got %q, want %q", string(got), "myrepo is myrepo and myrepo")
		}
	})
}

func TestReuseRegistration(t *testing.T) {
	formHTML := `<form><input id="csrf_token" name="csrf_token" type="hidden" value="test_token_123"></form>`
	successHTML := `<h1>Registration successful</h1>`

	t.Run("successful registration", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				// Set cookie like the real server does
				http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc"})
				_, _ = fmt.Fprint(w, formHTML)
				return
			}
			// Verify the form data was sent correctly
			if err := r.ParseForm(); err != nil {
				t.Errorf("ParseForm failed: %v", err)
			}
			if r.FormValue("csrf_token") != "test_token_123" {
				t.Errorf("wrong csrf_token: %q", r.FormValue("csrf_token"))
			}
			if r.FormValue("name") != "Test Name" {
				t.Errorf("wrong name: %q", r.FormValue("name"))
			}
			if r.FormValue("confirm") != "test@example.com" {
				t.Errorf("wrong confirm: %q", r.FormValue("confirm"))
			}
			if r.FormValue("project") != "github.com/user/repo" {
				t.Errorf("wrong project: %q", r.FormValue("project"))
			}
			_, _ = fmt.Fprint(w, successHTML)
		}))
		defer server.Close()

		err := reuseRegistration(server.URL, "Test Name", "test@example.com", "github.com/user/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing csrf token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "<form>no token here</form>")
		}))
		defer server.Close()

		err := reuseRegistration(server.URL, "Test Name", "test@example.com", "github.com/user/repo")
		if err == nil {
			t.Error("expected error for missing CSRF token, got nil")
		}
	})

	t.Run("registration rejected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				_, _ = fmt.Fprint(w, formHTML)
				return
			}
			// Server returns the form again on failure
			_, _ = fmt.Fprint(w, formHTML)
		}))
		defer server.Close()

		err := reuseRegistration(server.URL, "Test Name", "test@example.com", "github.com/user/repo")
		if err == nil {
			t.Error("expected error for failed registration, got nil")
		}
	})
}

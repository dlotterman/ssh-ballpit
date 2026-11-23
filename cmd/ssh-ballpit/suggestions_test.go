package main

import (
	"os"
	"sync"
	"testing"
)

func TestConcurrentWhitelistSuggestions(t *testing.T) {
	s, err := newSuggestions(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, command := range []string{"printf 'a.b'", "printf 'a.b'", "hostname", "hostname | cat"} {
		wg.Go(func() {
			if err := s.record(command); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	info, err := os.Stat(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Fatalf("suggestion file permissions = %o, want 0600", mode)
	}
	patterns, err := readCommandPatterns(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) != 2 {
		t.Fatalf("got %d suggested patterns, want 2", len(patterns))
	}
	for command, want := range map[string]bool{
		"printf 'a.b'":   true,
		"hostname":       true,
		"printf 'axb'":   false,
		"hostname | cat": false,
	} {
		if got := matchesCommand(patterns, command); got != want {
			t.Errorf("suggestions matched %q = %t, want %t", command, got, want)
		}
	}
}

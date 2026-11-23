package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
	"mvdan.cc/sh/v3/syntax"
)

// A fresh, private file per daemon run keeps collection separate from active rules.
type suggestions struct {
	mu       sync.Mutex
	path     string
	seen     map[string]bool
	document []byte
}

func newSuggestions(dir string) (*suggestions, error) {
	path, err := writePrivateTemp(dir, "suggested-whitelist-*.yaml", []byte("patterns: []\n"))
	if err != nil {
		return nil, err
	}
	return &suggestions{
		path:     path,
		seen:     make(map[string]bool),
		document: []byte("patterns:\n"),
	}, nil
}

func (s *suggestions) record(command string) error {
	if !utf8.ValidString(command) {
		return nil
	}
	tree, err := syntax.NewParser(syntax.KeepComments(true)).Parse(strings.NewReader(command), "")
	if err != nil || !singleWhitelistCommand(tree) {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen[command] {
		return nil
	}
	// A YAML sequence can be extended with another single-item sequence at the
	// same indentation. Encode only the new item, even in a long-running daemon.
	item, err := yaml.Marshal([]string{regexp.QuoteMeta(command)})
	if err != nil {
		return err
	}
	if len(s.document)+len(item) > maxYAMLFileBytes {
		return fmt.Errorf("suggested rules exceed 1 MiB")
	}
	data := append(s.document, item...)
	// Readers always see a complete rules document, including during a failed write.
	tmpPath, err := writePrivateTemp(filepath.Dir(s.path), ".suggestions-*", data)
	if err != nil {
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	s.seen[command] = true
	s.document = data
	return nil
}

// Return the name only after the entire private file has been written and closed.
func writePrivateTemp(dir, pattern string, data []byte) (string, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

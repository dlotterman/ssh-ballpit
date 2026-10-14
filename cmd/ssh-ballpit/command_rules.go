package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
	"mvdan.cc/sh/v3/syntax"
)

// Rules are compiled once at startup and shared read-only by all sessions.
type commandRules struct {
	whitelist []*regexp.Regexp
	blacklist []*regexp.Regexp
}

// localCommandDecision is shared by SSH and the check CLI. A nonempty source
// is a final local decision; an empty source requires cache/provider review.
func localCommandDecision(config Config, command string) (string, bool, error) {
	if err := policy(command); err != nil {
		return "static_policy", false, err
	}
	source, safe := config.commandRules.decision(command)
	if source == "blacklist" {
		return source, false, fmt.Errorf("command matches blacklist")
	}
	return source, safe, nil
}

func readCommandRules(dir string) (commandRules, error) {
	whitelist, err := readCommandPatterns(filepath.Join(dir, "whitelist.yaml"))
	if err != nil {
		return commandRules{}, err
	}
	blacklist, err := readCommandPatterns(filepath.Join(dir, "blacklist.yaml"))
	if err != nil {
		return commandRules{}, err
	}
	return commandRules{whitelist: whitelist, blacklist: blacklist}, nil
}

func readCommandPatterns(path string) ([]*regexp.Regexp, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, nil // Optional files preserve existing deployments.
	}
	if err != nil {
		return nil, err
	}
	// A dangling symlink is a broken configured list, not an absent list.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config struct {
		Patterns []yaml.Node `yaml:"patterns"`
	}
	if decodeYAML(data, &config) != nil || config.Patterns == nil {
		return nil, fmt.Errorf("invalid command rules YAML: %s (expected patterns list)", path)
	}

	patterns := make([]*regexp.Regexp, 0, len(config.Patterns))
	for index, node := range config.Patterns {
		if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
			return nil, fmt.Errorf("invalid command regex in %s at pattern %d (expected string)", path, index+1)
		}
		pattern := node.Value
		// Absolute anchors cannot be weakened by multiline mode or alternation.
		compiled, err := regexp.Compile(`\A(?:` + pattern + `)\z`)
		if err != nil || strings.TrimSpace(pattern) == "" {
			return nil, fmt.Errorf("invalid command regex in %s at pattern %d", path, index+1)
		}
		patterns = append(patterns, compiled)
	}
	return patterns, nil
}

func matchesCommand(patterns []*regexp.Regexp, command string) bool {
	for _, re := range patterns {
		if re.MatchString(command) {
			return true
		}
	}
	return false
}

// decision runs after the static policy and before the cache or reviewers.
// Empty source means normal review is required. Blacklist always wins.
func (r commandRules) decision(command string) (source string, safe bool) {
	if matchesCommand(r.blacklist, command) {
		return "blacklist", false
	}
	if len(r.blacklist) == 0 && len(r.whitelist) == 0 {
		return "", false
	}
	tree, err := syntax.NewParser(syntax.KeepComments(true)).Parse(strings.NewReader(command), "")
	if err != nil {
		return "", false // The caller's static policy rejects invalid syntax.
	}
	// Also inspect each simple command, so lists, pipelines, assignments and
	// surrounding whitespace cannot hide a blacklisted invocation.
	blocked := false
	syntax.Walk(tree, func(n syntax.Node) bool {
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 || blocked {
			return !blocked
		}
		first, last := call.Args[0], call.Args[len(call.Args)-1]
		invocation := command[first.Pos().Offset():last.End().Offset()]
		blocked = matchesCommand(r.blacklist, invocation)
		// Normalize inter-argument whitespace, but preserve each word's spelling.
		words := make([]string, len(call.Args))
		for i, word := range call.Args {
			words[i] = command[word.Pos().Offset():word.End().Offset()]
		}
		blocked = blocked || matchesCommand(r.blacklist, strings.Join(words, " "))
		return !blocked
	})
	if blocked {
		return "blacklist", false
	}
	if matchesCommand(r.whitelist, command) && singleWhitelistCommand(tree) {
		return "whitelist", true
	}
	return "", false
}

// A wildcard must not approve additional shell operations. Restrict bypass to
// one foreground simple command with literal/quoted words (globs are allowed).
// Other syntax still goes through normal review, unless static policy denies it.
func singleWhitelistCommand(tree *syntax.File) bool {
	if len(tree.Stmts) != 1 || len(tree.Last) != 0 {
		return false
	}
	stmt := tree.Stmts[0]
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 || len(call.Assigns) != 0 || len(stmt.Redirs) != 0 ||
		stmt.Background || stmt.Negated || stmt.Coprocess || stmt.Disown || len(stmt.Comments) != 0 {
		return false
	}
	literal := true
	for _, word := range call.Args {
		syntax.Walk(word, func(n syntax.Node) bool {
			switch v := n.(type) {
			case nil, *syntax.Word, *syntax.Lit:
			case *syntax.SglQuoted:
				literal = literal && !v.Dollar
			case *syntax.DblQuoted:
				literal = literal && !v.Dollar
			default:
				literal = false
			}
			return literal
		})
	}
	return literal
}

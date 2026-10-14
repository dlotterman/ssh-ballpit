package main

import (
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// policy is the non-overridable first check, before command rules or models.
func policy(command string) error {
	if strings.TrimSpace(command) == "" || len(command) > maxCommandBytes || strings.ContainsRune(command, 0) {
		return fmt.Errorf("empty or oversized command")
	}
	tree, err := syntax.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return fmt.Errorf("unsupported shell syntax")
	}
	syntax.Walk(tree, func(n syntax.Node) bool {
		if err != nil {
			return false
		}
		switch v := n.(type) {
		case *syntax.CallExpr:
			if len(v.Args) > 0 {
				name := v.Args[0].Lit()
				if name == "" || strings.Contains(name, "/") || disallowedEntryPoint(name) {
					err = fmt.Errorf("command entry point is disallowed by initial policy")
				}
			}
		case *syntax.CmdSubst, *syntax.ProcSubst:
			err = fmt.Errorf("command substitutions are disallowed by initial policy")
		}
		return err == nil
	})
	return err
}

// These entry points accept hidden programs or provide alternate sessions.
func disallowedEntryPoint(name string) bool {
	switch name {
	case "sh", "bash", "dash", "zsh", "fish", "ksh": // Shells.
		return true
	case "python", "python3", "perl", "ruby", "node", "lua": // Interpreters.
		return true
	case "eval", "exec", "source", ".": // Shell builtins that execute more code.
		return true
	case "ssh", "scp", "sftp": // Alternate remote access.
		return true
	case "sudo", "su", "doas": // Account or privilege changes.
		return true
	case "env", "xargs", "nohup", "setsid", "screen", "tmux": // Command wrappers.
		return true
	default:
		return false
	}
}

// Package examples contains the starter configuration embedded in ssh-ballpit.
package examples

import "embed"

// Files keeps configure independent of the working directory or source tree.
//
//go:embed config.example.json server.yaml providers/ollama.yaml whitelist.yaml blacklist.yaml
var Files embed.FS

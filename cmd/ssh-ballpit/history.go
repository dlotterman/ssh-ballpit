package main

import (
	"encoding/json"
	"strings"
)

const recentHistoryLimit = 5
const maxCommandBytes = 16384

type historyEntry struct {
	Command   string `json:"command"`
	Source    string `json:"source"`
	Truncated bool   `json:"truncated,omitempty"`
}

type recentHistory struct {
	Approved []historyEntry `json:"approved"`
	Rejected []historyEntry `json:"rejected"`
}

type reviewPayload struct {
	User          string        `json:"user"`
	Command       string        `json:"command"`
	RecentHistory recentHistory `json:"recent_history"`
}

func marshalReviewPayload(user, command string, history recentHistory) []byte {
	if history.Approved == nil {
		history.Approved = []historyEntry{}
	}
	if history.Rejected == nil {
		history.Rejected = []historyEntry{}
	}
	b, _ := json.Marshal(reviewPayload{User: user, Command: command, RecentHistory: history})
	return b
}

func (r *approvalReviewer) historySnapshot(user, backendUser string) recentHistory {
	r.mu.Lock()
	defer r.mu.Unlock()
	h := r.histories[[2]string{user, backendUser}]
	return recentHistory{Approved: append([]historyEntry{}, h.Approved...), Rejected: append([]historyEntry{}, h.Rejected...)}
}

func (r *approvalReviewer) remember(user, backendUser, command string, approved bool, source string) {
	entry := historyEntry{Command: command, Source: source}
	if len(command) > maxCommandBytes {
		entry.Command = strings.Clone(command[:maxCommandBytes])
		entry.Truncated = true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := [2]string{user, backendUser}
	h := r.histories[key]
	entries := &h.Rejected
	if approved {
		entries = &h.Approved
	}
	if len(*entries) == recentHistoryLimit {
		*entries = (*entries)[1:]
	}
	*entries = append(*entries, entry)
	r.histories[key] = h
}

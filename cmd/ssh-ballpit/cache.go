package main

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

const defaultApprovalCacheEntries = 10000

// One reviewer belongs to one daemon and its immutable startup configuration.
// The approval cache stores only hashes; recent decision history stores up to
// five approved and five rejected command texts per frontend/backend user.
type approvalReviewer struct {
	histories map[[2]string]recentHistory
	config    Config
	mu        sync.Mutex
	entries   map[[32]byte]*list.Element
	recent    list.List
}

func newApprovalReviewer(config Config) *approvalReviewer {
	if config.ApprovalCacheEntries == 0 {
		config.ApprovalCacheEntries = defaultApprovalCacheEntries
	}
	return &approvalReviewer{config: config, entries: make(map[[32]byte]*list.Element), histories: make(map[[2]string]recentHistory)}
}

func (r *approvalReviewer) check(ctx context.Context, user, backendUser, command string) (results []Result, safe, hit bool) {
	if ctx.Err() != nil {
		r.remember(user, backendUser, command, false, "review_error")
		return nil, false, false
	}
	// Length prefixes preserve boundaries and all bytes, including invalid UTF-8.
	var b []byte
	for _, value := range []string{user, backendUser, command} {
		b = binary.BigEndian.AppendUint64(b, uint64(len(value)))
		b = append(b, value...)
	}
	key := sha256.Sum256(b)
	if r.config.ApprovalCacheEntries > 0 {
		r.mu.Lock()
		entry := r.entries[key]
		if entry != nil {
			r.recent.MoveToFront(entry)
		}
		r.mu.Unlock()
		if entry != nil {
			if r.config.metrics != nil {
				r.config.metrics.cacheLookups.WithLabelValues("hit").Inc()
			}
			r.remember(user, backendUser, command, true, "approval_cache")
			return nil, true, true
		}
	}
	if r.config.metrics != nil {
		outcome := "miss"
		if r.config.ApprovalCacheEntries < 0 {
			outcome = "disabled"
		}
		r.config.metrics.cacheLookups.WithLabelValues(outcome).Inc()
	}
	results, safe = reviewWithHistory(ctx, r.config, backendUser, command, r.historySnapshot(user, backendUser))
	if safe && ctx.Err() == nil && r.config.ApprovalCacheEntries > 0 {
		r.mu.Lock()
		if entry := r.entries[key]; entry != nil {
			r.recent.MoveToFront(entry)
		} else {
			r.entries[key] = r.recent.PushFront(key)
			if r.recent.Len() > r.config.ApprovalCacheEntries {
				old := r.recent.Back()
				delete(r.entries, old.Value.([32]byte))
				r.recent.Remove(old)
			}
		}
		if r.config.metrics != nil {
			r.config.metrics.cacheEntries.Set(float64(r.recent.Len()))
		}
		r.mu.Unlock()
	}
	source := "review_error"
	if safe {
		source = "reviewers"
	} else {
		for _, result := range results {
			if result.Verdict != nil && !result.Safe && result.Error == "" {
				source = "reviewers"
				break
			}
		}
	}
	r.remember(user, backendUser, command, safe, source)
	return results, safe, false
}

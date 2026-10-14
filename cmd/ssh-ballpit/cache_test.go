package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestApprovalCache(t *testing.T) {
	var calls atomic.Int32
	var fail atomic.Bool
	var deny atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if fail.Load() && r.URL.Path == "/second/api/chat" {
			w.WriteHeader(503)
			return
		}
		answer := safeJSON
		if deny.Load() && r.URL.Path == "/second/api/chat" {
			answer = `{"offensive":false,"danger_score":90,"reasons":[],"destinations":[]}`
		}
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": answer}})
	}))
	defer server.Close()
	c := Config{ReviewSeconds: 2, ApprovalCacheEntries: 2, Providers: []Provider{
		{Name: "first", Kind: "ollama", Endpoint: server.URL, Model: "test", Threshold: 30},
		{Name: "second", Kind: "ollama", Endpoint: server.URL + "/second", Model: "test", Threshold: 30},
	}}
	cache := newApprovalReviewer(c)
	check := func(user, backend, command string, wantSafe, wantHit bool) {
		t.Helper()
		_, safe, hit := cache.check(context.Background(), user, backend, command)
		if safe != wantSafe || hit != wantHit {
			t.Fatalf("safe=%v hit=%v", safe, hit)
		}
	}
	check("alice", "local", "pwd", true, false)
	fail.Store(true)
	check("alice", "local", "pwd", true, true)
	if calls.Load() != 2 {
		t.Fatal("cache hit called a reviewer")
	}
	for i := 0; i < 2; i++ {
		check("alice", "local", "ls", false, false)
	}
	if calls.Load() != 6 {
		t.Fatal("partial approval/error was cached")
	}
	fail.Store(false)
	deny.Store(true)
	for i := 0; i < 2; i++ {
		check("alice", "local", "ls", false, false)
	}
	if calls.Load() != 10 {
		t.Fatal("denial was cached")
	}
	deny.Store(false)
	check("bob", "local", "pwd", true, false)
	check("alice", "local", "pwd", true, true) // refresh LRU, evict bob next
	check("alice", "other", "pwd", true, false)
	check("alice", "local", "pwd", true, true)
	check("bob", "local", "pwd", true, false)
	check("bob", "local", "pwd ", true, false)
	check("bob", "local", "pwd\xff", true, false)
	check("bob", "local", "pwd\xfe", true, false) // distinct non-UTF8 bytes
	fresh := newApprovalReviewer(c)
	if _, _, hit := fresh.check(context.Background(), "bob", "local", "pwd\xfe"); hit {
		t.Fatal("cache survived new daemon configuration")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, safe, _ := cache.check(ctx, "bob", "local", "pwd\xfe"); safe {
		t.Fatal("approved cancelled request")
	}
	c.ApprovalCacheEntries = -1
	disabled := newApprovalReviewer(c)
	for i := 0; i < 2; i++ {
		if _, safe, hit := disabled.check(context.Background(), "alice", "local", "pwd"); !safe || hit {
			t.Fatal("disabled cache")
		}
	}
}

func TestApprovalCacheConcurrentHits(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": safeJSON}})
	}))
	defer server.Close()
	cache := newApprovalReviewer(Config{ReviewSeconds: 2, Providers: []Provider{{Kind: "ollama", Endpoint: server.URL, Threshold: 30}}})
	if _, safe, _ := cache.check(context.Background(), "alice", "local", "pwd"); !safe {
		t.Fatal("initial approval failed")
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, safe, hit := cache.check(context.Background(), "alice", "local", "pwd"); !safe || !hit {
				t.Error("missing concurrent hit")
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal("concurrent cache hits called provider")
	}
}

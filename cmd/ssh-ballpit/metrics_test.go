package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsListenerLifecycle(t *testing.T) {
	m := newDaemonMetrics(Config{})
	stop, err := startMetrics("", m, func() { t.Error("disabled metrics failed") })
	if err != nil {
		t.Fatal(err)
	}
	stop()
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := occupied.Addr().String()
	if stop, err := startMetrics(address, m, func() {}); err == nil {
		stop()
		t.Fatal("ignored occupied metrics port")
	}
	occupied.Close()
	stop, err = startMetrics(address, m, func() { t.Error("normal close reported as failure") })
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get("http://" + address + "/metrics")
	if err != nil {
		stop()
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status %d", resp.StatusCode)
	}
	stop()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal("metrics listener leaked after close:", err)
	}
	listener.Close()
	// Each daemon has its own registry; no global collector collisions.
	other := newDaemonMetrics(Config{})
	recorder := httptest.NewRecorder()
	other.handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/not-metrics", nil))
	if recorder.Code != 404 {
		t.Fatal("unexpected metrics endpoint")
	}
}

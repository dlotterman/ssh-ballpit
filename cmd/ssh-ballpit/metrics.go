package main

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type daemonMetrics struct {
	registry *prometheus.Registry

	connections          prometheus.Counter
	connectionRejections prometheus.Counter
	handshakeFailures    prometheus.Counter
	keyBypasses          prometheus.Counter
	hostKeyBypasses      prometheus.Counter
	activeConnections    prometheus.Gauge

	authentication     *prometheus.CounterVec
	keyRejections      *prometheus.CounterVec
	protocolRejections *prometheus.CounterVec
	commands           *prometheus.CounterVec

	cacheEntries      prometheus.Gauge
	cacheLookups      *prometheus.CounterVec
	reviews           *prometheus.CounterVec
	reviewDuration    *prometheus.HistogramVec
	executionDuration prometheus.Histogram
}

func newDaemonMetrics(config Config) *daemonMetrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	factory := promauto.With(registry)
	counter := func(name, help string) prometheus.Counter {
		return factory.NewCounter(prometheus.CounterOpts{Namespace: "ssh_ballpit", Name: name, Help: help})
	}
	vector := func(name, help string, labels ...string) *prometheus.CounterVec {
		return factory.NewCounterVec(prometheus.CounterOpts{Namespace: "ssh_ballpit", Name: name, Help: help}, labels)
	}
	gauge := func(name, help string) prometheus.Gauge {
		return factory.NewGauge(prometheus.GaugeOpts{Namespace: "ssh_ballpit", Name: name, Help: help})
	}
	m := &daemonMetrics{
		registry:             registry,
		connections:          counter("connections_total", "Accepted TCP connections, including those rejected at capacity."),
		connectionRejections: counter("connection_limit_rejections_total", "TCP connections rejected at the configured connection limit."),
		handshakeFailures:    counter("handshake_failures_total", "SSH handshakes that failed outside daemon shutdown."),
		hostKeyBypasses:      counter("backend_host_key_bypasses_total", "Backend SSH host-key trust failures bypassed in accept-all mode after signature verification."),
		keyBypasses:          counter("key_authorization_bypasses_total", "Authenticated SSH connections using an otherwise unauthorized public key in accept-all mode."),
		activeConnections:    gauge("connections_active", "Currently handled TCP connections, including SSH handshakes."),
		cacheEntries:         gauge("approval_cache_entries", "Approvals currently stored in the in-memory cache."),
		authentication:       vector("authentication_attempts_total", "SSH authentication results, excluding normal none-method discovery.", "result"),
		keyRejections:        vector("public_key_rejections_total", "Rejected public-key offers; offers do not prove possession of the private key.", "reason"),
		protocolRejections:   vector("protocol_rejections_total", "Unsupported or malformed SSH channels and requests.", "kind"),
		commands:             vector("commands_total", "Completed command decisions and execution outcomes.", "outcome"),
		cacheLookups:         vector("approval_cache_lookups_total", "Approval cache lookups, after static policy and command rules.", "result"),
		reviews:              vector("reviews_total", "Logical provider reviews; retries count as one review.", "reviewer", "outcome"),
		reviewDuration:       factory.NewHistogramVec(prometheus.HistogramOpts{Namespace: "ssh_ballpit", Name: "review_duration_seconds", Help: "Provider review latency including retries and subscription queue time.", Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60, 120}}, []string{"reviewer"}),
		executionDuration:    factory.NewHistogram(prometheus.HistogramOpts{Namespace: "ssh_ballpit", Name: "execution_duration_seconds", Help: "Backend connection and execution latency, including failures.", Buckets: []float64{.01, .1, .5, 1, 5, 10, 30, 60, 300}}),
	}
	lab := gauge("accept_all_keys_enabled", "Whether client-key authorization and backend host-key trust checks may be bypassed (1 enabled, 0 disabled).")
	if config.AcceptAllKeys {
		lab.Set(1)
	}
	for _, label := range []string{"success", "failure"} {
		m.authentication.WithLabelValues(label)
	}
	for _, label := range []string{"unknown_user", "missing_backend_user", "key_not_in_authorized_keys", "key_not_allowed_for_user"} {
		m.keyRejections.WithLabelValues(label)
	}
	for _, label := range []string{"channel", "request", "malformed_exec"} {
		m.protocolRejections.WithLabelValues(label)
	}
	for _, label := range []string{"static_denied", "blacklist_denied", "review_denied", "success", "nonzero_exit", "backend_error"} {
		m.commands.WithLabelValues(label)
	}
	for _, label := range []string{"hit", "miss", "disabled"} {
		m.cacheLookups.WithLabelValues(label)
	}
	for _, provider := range config.Providers {
		for _, label := range []string{"approved", "denied", "error"} {
			m.reviews.WithLabelValues(provider.Name, label)
		}
		m.reviewDuration.WithLabelValues(provider.Name)
	}
	return m
}

func (m *daemonMetrics) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Timeout: 8 * time.Second, MaxRequestsInFlight: 5}))
	return mux
}

func startMetrics(address string, m *daemonMetrics, onFailure func()) (func(), error) {
	if address == "" {
		return func() {}, nil
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("metrics listen failed", "listen_addr", address, "reason", connectionReason(err))
		return nil, err
	}
	server := &http.Server{
		Handler:           m.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8192,
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", "reason", connectionReason(err))
			onFailure()
		}
	}()
	slog.Info("metrics listening", "address", listener.Addr().String(), "path", "/metrics")
	return func() {
		server.Close()
		<-done
	}, nil
}

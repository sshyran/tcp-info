// The metrics package defines prometheus metric types and provides
// convenience methods to add accounting to various parts of the pipeline.
//
// When defining new operations or metrics, these are helpful values to track:
//  - things coming into or go out of the system: requests, files, tests, api calls.
//  - the success or error status of any of the above.
//  - the distribution of processing latency.
package metrics

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/rtx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupPrometheus configures prometheus metrics on the specified port.
// If promPort is zero, it will ask the system to choose a port (e.g. for testing).
// Also registers pprof handlers on the same port.
func SetupPrometheus(promPort int) *http.Server {
	if promPort < 0 {
		log.Println("Not exporting prometheus metrics")
		return nil
	}

	// Define a custom serve mux for prometheus to listen on a separate port.
	// We listen on a separate port so we can forward this port on the host VM.
	// We cannot forward port 8080 because it is used by AppEngine.
	mux := http.NewServeMux()
	// Assign the default prometheus handler to the standard exporter path.
	mux.Handle("/metrics", promhttp.Handler())
	// Assign the pprof handling paths to the external port to access individual
	// instances.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	prometheus.MustRegister(SyscallTimeHistogram)

	prometheus.MustRegister(ConnectionCountHistogram)
	prometheus.MustRegister(CacheSizeHistogram)

	prometheus.MustRegister(NewFileCount)
	prometheus.MustRegister(ErrorCount)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", promPort),
		Handler: mux,
	}
	rtx.Must(httpx.ListenAndServeAsync(server), "Could not start metrics server")

	log.Println("Exporting prometheus metrics on", server.Addr)
	return server
}

var (
	// SyscallTime tracks the latency in the syscall.  It does NOT include
	// the time to process the netlink messages.
	SyscallTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tcpinfo_syscall_time_histogram",
			Help: "netlink syscall latency distribution",
			Buckets: []float64{
				0.001, 0.00125, 0.0016, 0.002, 0.0025, 0.0032, 0.004, 0.005, 0.0063, 0.0079,
				0.01, 0.0125, 0.016, 0.02, 0.025, 0.032, 0.04, 0.05, 0.063, 0.079,
				0.1, 0.125, 0.16, 0.2,
			},
		},
		[]string{"af"})

	// ConnectionCountHistogram tracks the number of connections returned by
	// each syscall.  This ??? includes local connections that are NOT recorded
	// in the cache or output.
	// TODO - convert this to integer bins.
	ConnectionCountHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tcpinfo_connection_count_histogram",
			Help: "connection count histogram",
			Buckets: []float64{
				1, 2, 3, 4, 5, 6, 8,
				10, 12.5, 16, 20, 25, 32, 40, 50, 63, 79,
				100, 125, 160, 200, 250, 320, 400, 500, 630, 790,
				1000, 1250, 1600, 2000, 2500, 3200, 4000, 5000, 6300, 7900,
				10000, 12500, 16000, 20000, 25000, 32000, 40000, 50000, 63000, 79000,
				10000000,
			},
		},
		[]string{"af"})

	// CacheSizeHistogram tracks the number of entries in connection cache.
	// TODO - convert this to integer bins.
	CacheSizeHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "tcpinfo_cache_count_histogram",
			Help: "cache connection count histogram",
			Buckets: []float64{
				1, 2, 3, 4, 5, 6, 8,
				10, 12.5, 16, 20, 25, 32, 40, 50, 63, 79,
				100, 125, 160, 200, 250, 320, 400, 500, 630, 790,
				1000, 1250, 1600, 2000, 2500, 3200, 4000, 5000, 6300, 7900,
				10000, 12500, 16000, 20000, 25000, 32000, 40000, 50000, 63000, 79000,
				10000000,
			},
		})

	// ErrorCount measures the number of errors
	// Provides metrics:
	//    tcpinfo_Error_Count
	// Example usage:
	//    metrics.ErrorCount.With(prometheus.Labels{"type", "foobar"}).Inc()
	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tcpinfo_error_total",
			Help: "The total number of errors encountered.",
		}, []string{"type"})

	// NewFileCount counts the number of connection files written.
	//
	// Provides metrics:
	//   tcpinfo_new_file_count
	// Example usage:
	//   metrics.FileCount.Inc()
	NewFileCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tcpinfo_new_file_total",
			Help: "Number of files created.",
		},
	)
)

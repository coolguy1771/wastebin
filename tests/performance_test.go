package tests_test

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/pkg/testutil"
	"github.com/coolguy1771/wastebin/routes"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger, err := log.New(os.Stdout, "ERROR")
	if err != nil {
		panic(err)
	}
	log.ResetDefault(logger)

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// BenchmarkCreatePaste benchmarks paste creation performance.
func BenchmarkCreatePaste(b *testing.B) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(&testing.T{}, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	formData := map[string]string{
		"text":      "Benchmark test content for performance testing",
		"extension": "txt",
		"expires":   "60",
		"burn":      "false",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server.MakeRequest(testutil.HTTPRequest{
				Method:   "POST",
				Path:     "/api/v1/paste",
				FormData: formData,
			})
		}
	})
}

// BenchmarkGetPaste benchmarks paste retrieval performance.
func BenchmarkGetPaste(b *testing.B) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(&testing.T{}, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// Create a test paste
	paste := server.CreateTestPaste("Benchmark retrieval content", "txt", 60, false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + paste.UUID.String(),
			})
		}
	})
}

// BenchmarkGetRawPaste benchmarks raw paste retrieval performance.
func BenchmarkGetRawPaste(b *testing.B) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(&testing.T{}, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// Create a test paste
	paste := server.CreateTestPaste("Benchmark raw retrieval content", "txt", 60, false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/paste/" + paste.UUID.String() + "/raw",
			})
		}
	})
}

// BenchmarkLargePasteCreation benchmarks creation of large pastes.
func BenchmarkLargePasteCreation(b *testing.B) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(&testing.T{}, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// 1MB content
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + (i % 26))
	}

	formData := map[string]string{
		"text":      string(largeContent),
		"extension": "txt",
		"expires":   "60",
		"burn":      "false",
	}

	b.ResetTimer()

	for range b.N {
		server.MakeRequest(testutil.HTTPRequest{
			Method:   "POST",
			Path:     "/api/v1/paste",
			FormData: formData,
		})
	}
}

// TestConcurrentOperations tests concurrent read/write performance.
func TestConcurrentOperations(t *testing.T) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	const (
		numGoroutines = 50
		numOperations = 20
	)

	var wg sync.WaitGroup

	results := make(chan result, numGoroutines*numOperations)

	start := time.Now()

	// Start concurrent workers
	for i := range numGoroutines {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for j := range numOperations {
				opStart := time.Now()

				// Create paste
				resp := server.MakeRequest(testutil.HTTPRequest{
					Method: "POST",
					Path:   "/api/v1/paste",
					FormData: map[string]string{
						"text":      fmt.Sprintf("Concurrent test %d-%d", workerID, j),
						"extension": "txt",
						"expires":   "60",
						"burn":      "false",
					},
				})

				opDuration := time.Since(opStart)
				results <- result{
					operation: "create",
					success:   resp.StatusCode == http.StatusCreated,
					duration:  opDuration,
				}

				if resp.StatusCode == http.StatusCreated {
					pasteUUID := resp.JSON["uuid"].(string)

					// Read paste
					opStart = time.Now()
					getResp := server.MakeRequest(testutil.HTTPRequest{
						Method: "GET",
						Path:   "/api/v1/paste/" + pasteUUID,
					})
					opDuration = time.Since(opStart)

					results <- result{
						operation: "read",
						success:   getResp.StatusCode == http.StatusOK,
						duration:  opDuration,
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(start)

	// Analyze results
	var (
		createOps, readOps, successfulOps int
		totalCreateTime, totalReadTime    time.Duration
	)

	for res := range results {
		if res.operation == "create" {
			createOps++
			totalCreateTime += res.duration
		} else {
			readOps++
			totalReadTime += res.duration
		}

		if res.success {
			successfulOps++
		}
	}

	totalOps := createOps + readOps
	successRate := float64(successfulOps) / float64(totalOps) * 100

	t.Logf("Performance Test Results:")
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Total Operations: %d", totalOps)
	t.Logf("  Success Rate: %.2f%%", successRate)
	t.Logf("  Throughput: %.2f ops/sec", float64(totalOps)/totalDuration.Seconds())
	if createOps > 0 {
		t.Logf("  Average Create Time: %v", totalCreateTime/time.Duration(createOps))
	} else {
		t.Logf("  Average Create Time: N/A (no successful creates)")
	}
	if readOps > 0 {
		t.Logf("  Average Read Time: %v", totalReadTime/time.Duration(readOps))
	} else {
		t.Logf("  Average Read Time: N/A (no successful reads)")
	}

	// Assertions
	assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
	if createOps > 0 {
		assert.Less(
			t,
			totalCreateTime/time.Duration(createOps),
			100*time.Millisecond,
			"Average create time should be < 100ms",
		)
	}
	if readOps > 0 {
		assert.Less(t, totalReadTime/time.Duration(readOps), 50*time.Millisecond, "Average read time should be < 50ms")
	}
}

// TestLoadTesting performs basic load testing.
func TestLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// Load test parameters
	const (
		duration = 30 * time.Second
		workers  = 10
	)

	var wg sync.WaitGroup

	results := make(chan result, 1000)
	stop := make(chan struct{})

	start := time.Now()

	// Start workers
	for i := range workers {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			counter := 0

			for {
				select {
				case <-stop:
					return
				default:
					counter++
					opStart := time.Now()

					resp := server.MakeRequest(testutil.HTTPRequest{
						Method: "POST",
						Path:   "/api/v1/paste",
						FormData: map[string]string{
							"text":      fmt.Sprintf("Load test %d-%d", workerID, counter),
							"extension": "txt",
							"expires":   "60",
							"burn":      "false",
						},
					})

					results <- result{
						operation: "create",
						success:   resp.StatusCode == http.StatusCreated,
						duration:  time.Since(opStart),
					}

					// Small delay to prevent overwhelming
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// Stop after duration
	time.AfterFunc(duration, func() {
		close(stop)
	})

	wg.Wait()
	close(results)

	// Analyze results
	var (
		totalOps, successfulOps  int
		totalDuration            time.Duration
		minDuration, maxDuration time.Duration = time.Hour, 0
	)

	for res := range results {
		totalOps++
		totalDuration += res.duration

		if res.success {
			successfulOps++
		}

		if res.duration < minDuration {
			minDuration = res.duration
		}

		if res.duration > maxDuration {
			maxDuration = res.duration
		}
	}

	actualDuration := time.Since(start)
	successRate := float64(successfulOps) / float64(totalOps) * 100
	avgDuration := totalDuration / time.Duration(totalOps)
	throughput := float64(totalOps) / actualDuration.Seconds()

	t.Logf("Load Test Results:")
	t.Logf("  Test Duration: %v", actualDuration)
	t.Logf("  Total Requests: %d", totalOps)
	t.Logf("  Successful Requests: %d", successfulOps)
	t.Logf("  Success Rate: %.2f%%", successRate)
	t.Logf("  Throughput: %.2f req/sec", throughput)
	t.Logf("  Average Response Time: %v", avgDuration)
	t.Logf("  Min Response Time: %v", minDuration)
	t.Logf("  Max Response Time: %v", maxDuration)

	// Performance assertions
	assert.Greater(t, successRate, 90.0, "Success rate should be > 90%")
	assert.Greater(t, throughput, 10.0, "Throughput should be > 10 req/sec")
	assert.Less(t, avgDuration, 500*time.Millisecond, "Average response time should be < 500ms")
	assert.Less(t, maxDuration, 2*time.Second, "Max response time should be < 2s")
}

// TestMemoryUsage tests memory usage under load.
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// Create many pastes to test memory usage
	const numPastes = 1000

	for i := range numPastes {
		content := fmt.Sprintf("Memory test paste %d with some content to test memory usage", i)

		resp := server.MakeRequest(testutil.HTTPRequest{
			Method: "POST",
			Path:   "/api/v1/paste",
			FormData: map[string]string{
				"text":      content,
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
		})

		require.Equal(t, 201, resp.StatusCode, "Should create paste successfully")

		// Every 100 pastes, test retrieval
		if (i+1)%100 == 0 {
			pasteUUID := resp.JSON["uuid"].(string)
			getResp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + pasteUUID,
			})
			assert.Equal(t, 200, getResp.StatusCode, "Should retrieve paste successfully")
		}
	}

	// Verify final count
	count := server.CountPastesInDB()
	assert.Equal(t, int64(numPastes), count, "Should have created all pastes")

	t.Logf("Successfully created and managed %d pastes", numPastes)
}

// TestDatabasePerformance tests database-specific performance.
func TestDatabasePerformance(t *testing.T) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	// Test batch creation performance
	const batchSize = 100

	start := time.Now()

	for i := range batchSize {
		server.CreateTestPaste(
			fmt.Sprintf("Batch test paste %d", i),
			"txt",
			60,
			false,
		)
	}

	batchDuration := time.Since(start)
	avgPerPaste := batchDuration / batchSize

	t.Logf("Database Performance:")
	t.Logf("  Batch Size: %d", batchSize)
	t.Logf("  Total Time: %v", batchDuration)
	t.Logf("  Time per Paste: %v", avgPerPaste)
	t.Logf("  Rate: %.2f pastes/sec", float64(batchSize)/batchDuration.Seconds())

	// Performance assertions
	assert.Less(t, avgPerPaste, 10*time.Millisecond, "Average paste creation should be < 10ms")
	assert.Greater(t, float64(batchSize)/batchDuration.Seconds(), 50.0, "Should create > 50 pastes/sec")
}

// result represents an operation result for performance testing.
type result struct {
	operation string
	success   bool
	duration  time.Duration
}

// TestResponseTimeDistribution tests the distribution of response times.
func TestResponseTimeDistribution(t *testing.T) {
	router := routes.AddRoutes(nil)

	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: true,
		EnableLogging: false,
	})
	defer server.Close()

	const numRequests = 200

	durations := make([]time.Duration, numRequests)

	// Make requests and collect durations
	for i := range numRequests {
		start := time.Now()

		server.MakeRequest(testutil.HTTPRequest{
			Method: "POST",
			Path:   "/api/v1/paste",
			FormData: map[string]string{
				"text":      "Response time test " + strconv.Itoa(i),
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
		})

		durations[i] = time.Since(start)
	}

	// Calculate percentiles
	percentiles := calculatePercentiles(durations)

	t.Logf("Response Time Distribution:")
	t.Logf("  P50: %v", percentiles[50])
	t.Logf("  P90: %v", percentiles[90])
	t.Logf("  P95: %v", percentiles[95])
	t.Logf("  P99: %v", percentiles[99])

	// Performance assertions
	assert.Less(t, percentiles[50], 50*time.Millisecond, "P50 should be < 50ms")
	assert.Less(t, percentiles[90], 100*time.Millisecond, "P90 should be < 100ms")
	assert.Less(t, percentiles[95], 200*time.Millisecond, "P95 should be < 200ms")
	assert.Less(t, percentiles[99], 500*time.Millisecond, "P99 should be < 500ms")
}

// calculatePercentiles calculates percentiles from duration slice.
func calculatePercentiles(durations []time.Duration) map[int]time.Duration {
	// Make a copy to avoid modifying the original slice
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := range len(sorted) - 1 - i {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	percentiles := make(map[int]time.Duration)

	for _, p := range []int{50, 90, 95, 99} {
		idx := (p * len(sorted)) / 100
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		percentiles[p] = sorted[idx]
	}

	return percentiles
}

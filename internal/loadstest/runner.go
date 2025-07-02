package loadtest

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Duration     time.Duration
	SuccessCount int
	TotalCount   int
}

func RunLoadTest(url string, count, concurrency int) (Result, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	var wg sync.WaitGroup
	requests := make(chan struct{}, concurrency)
	successCount := 0
	var mu sync.Mutex

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			for range requests {
				wg.Add(1)
				resp, err := client.Get(url)
				if err == nil && resp.StatusCode == http.StatusOK {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				if resp != nil {
					resp.Body.Close()
				}
				wg.Done()
			}
		}()
	}

	for i := 0; i < count; i++ {
		requests <- struct{}{}
	}
	close(requests)

	wg.Wait()
	duration := time.Since(start)

	return Result{
		Duration:     duration,
		SuccessCount: successCount,
		TotalCount:   count,
	}, nil
}

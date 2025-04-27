package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL           = "https://blogapi.project-test.info/blog/related?blogId=5ce5d19b-e010-4eae-8d58-63462f500c94&categories=rent-a-car,activities&language=en"
	ratePerSec        = 100
	durationSec       = 60
	reportIntervalSec = 2
)

func main() {
	var wg sync.WaitGroup
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var totalRequests int64
	var success int64
	var fail int64

	ticker := time.NewTicker(time.Second / time.Duration(ratePerSec))
	defer ticker.Stop()

	reportTicker := time.NewTicker(time.Duration(reportIntervalSec) * time.Second)
	defer reportTicker.Stop()

	stop := time.After(time.Duration(durationSec) * time.Second)
	startTime := time.Now()

	var prevTotal, prevSuccess, prevFail int64

	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-stop:
			wg.Wait()
			fmt.Printf("\nTest tamamlandı. Toplam istek: %d, Başarılı: %d, Başarısız: %d\n",
				atomic.LoadInt64(&totalRequests),
				atomic.LoadInt64(&success),
				atomic.LoadInt64(&fail),
			)
			return

		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				// URL'ye random bir query param ekle
				u, err := url.Parse(baseURL)
				if err != nil {
					atomic.AddInt64(&fail, 1)
					return
				}
				q := u.Query()
				q.Set("nocache", fmt.Sprintf("%d", rand.Int63()))
				u.RawQuery = q.Encode()

				resp, err := client.Get(u.String())
				atomic.AddInt64(&totalRequests, 1)
				if err != nil {
					atomic.AddInt64(&fail, 1)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&success, 1)
				} else {
					atomic.AddInt64(&fail, 1)
				}
			}()

		case <-reportTicker.C:
			elapsed := int(time.Since(startTime).Seconds())
			currTotal := atomic.LoadInt64(&totalRequests)
			currSuccess := atomic.LoadInt64(&success)
			currFail := atomic.LoadInt64(&fail)

			fmt.Printf(
				"[%2ds] Toplam: %d (+%d), Başarılı: %d (+%d), Başarısız: %d (+%d)\n",
				elapsed,
				currTotal, currTotal-prevTotal,
				currSuccess, currSuccess-prevSuccess,
				currFail, currFail-prevFail,
			)

			prevTotal = currTotal
			prevSuccess = currSuccess
			prevFail = currFail
		}
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"runtime/trace"
	"sync"
	"time"
	"unsafe"
)

var (
	parallelFlag   = flag.Int("p", 10, "The number goroutines making requests")
	requestsFlag   = flag.Int("r", 1000, "The total number of requests to make")
	idleFlag       = flag.Int("i", 10000, "The number of idle connections")
	timeoutFlag    = flag.Duration("t", 3*time.Second, "The timeout for each request")
	traceFlag      = flag.Bool("trace", false, "")
	reqFlag        = flag.Bool("req", false, "")
	sleepFlag      = flag.Bool("sleep", false, "")
	sleepAllocFlag = flag.Bool("sleepAlloc", false, "")
	chanFlag       = flag.Bool("chan", false, "")
	chanAllocFlag  = flag.Bool("chanAlloc", false, "")
	allocFlag      = flag.Bool("alloc", false, "")
	nogcFlag       = flag.Bool("nogc", false, "")
)

func main() {
	flag.Parse()

	if *nogcFlag {
		debug.SetGCPercent(-1)
	}

	httpClient := &http.Client{
		Timeout: time.Second,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          *idleFlag,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	wg := &sync.WaitGroup{}
	requestsPer := (*requestsFlag) / (*parallelFlag)
	extras := (*requestsFlag) % (*parallelFlag)

	if *traceFlag {
		startTrace()
		defer stopTrace()
	}
	for i := 0; i < *parallelFlag; i++ {
		requests := requestsPer
		if extras > 0 {
			requests++
			extras--
		}
		wg.Add(1)
		go func() {
			switch {
			case *reqFlag:
				makeRequests(requests, httpClient)
			case *sleepFlag:
				makeSleeps(requests)
			case *sleepAllocFlag:
				makeSleepsAlloc(requests)
			case *chanFlag:
				makeWaitChan(requests)
			case *chanAllocFlag:
				makeWaitChanAlloc(requests)
			case *allocFlag:
				makeAlloc(requests)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func makeRequests(requests int, client *http.Client) {
	for i := 0; i < requests; i++ {
		req, cancelFunc := makeRequest()
		start := time.Now()
		if resp, err := client.Do(req); err != nil {
			// TODO
			// Might be a timeout - should panic if not timeout
			//panic(err)
		} else {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}
		cancelFunc()
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
}

func makeSleeps(sleeps int) {
	for i := 0; i < sleeps; i++ {
		start := time.Now()
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
}

func makeSleepsAlloc(sleeps int) uintptr {
	var sum uintptr
	for i := 0; i < sleeps; i++ {
		req, cancelFunc := makeRequest()
		start := time.Now()
		time.Sleep(*timeoutFlag)
		cancelFunc()
		sum += uintptr(unsafe.Pointer(req))
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
	return sum
}

func makeWaitChan(sleeps int) uintptr {
	var sum uintptr
	for i := 0; i < sleeps; i++ {
		cancel := make(chan struct{}, 1)
		start := time.Now()
		go func() {
			time.Sleep(*timeoutFlag)
			close(cancel)
		}()
		<-cancel
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
	return sum
}

func makeWaitChanAlloc(sleeps int) uintptr {
	var sum uintptr
	for i := 0; i < sleeps; i++ {
		cancel := make(chan struct{}, 1)
		req, cancelFunc := makeRequest()
		start := time.Now()
		go func() {
			time.Sleep(*timeoutFlag)
			close(cancel)
		}()
		<-cancel
		cancelFunc()
		sum += uintptr(unsafe.Pointer(req))
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
	return sum
}

func makeAlloc(sleeps int) uintptr {
	var sum uintptr
	for i := 0; i < sleeps; i++ {
		start := time.Now()
		req, cancelFunc := makeRequest()
		cancelFunc()
		sum += uintptr(unsafe.Pointer(req))
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
	return sum
}

func makeRequest() (*http.Request, func()) {
	req, err := http.NewRequest("GET", "http://localhost:9001", nil)
	if err != nil {
		panic(err)
	}
	ctx, cancelFunc := context.WithTimeout(req.Context(), *timeoutFlag)
	req = req.WithContext(ctx)
	return req, cancelFunc
}

func startTrace() {
	f, err := os.OpenFile("trace", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	return
}

func stopTrace() {
	trace.Stop()
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	if nanoSamples, err := readNanoSamples(os.Stdin); err != nil {
		fmt.Printf("%s\n", err)
	} else {
		q := newQuartiles(nanoSamples)
		fmt.Printf("N: %d \nP100: %s \nP99 : %s \nP90 : %s \nP50 : %s \nP10 : %s \nP01 : %s \nP00 : %s \n", len(q.samples), time.Duration(q.p100), time.Duration(q.p99), time.Duration(q.p90), time.Duration(q.p50), time.Duration(q.p10), time.Duration(q.p01), time.Duration(q.p00))
	}
}

func readNanoSamples(r io.Reader) ([]int, error) {
	nanosSamples := []int{}
	scanner := bufio.NewScanner(os.Stdin)
	prefix := "Duration: "
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			nanoStr := strings.TrimPrefix(line, prefix)
			if nanos, err := strconv.Atoi(nanoStr); err != nil {
				return nil, err
			} else {
				nanosSamples = append(nanosSamples, nanos)
			}
		} else {
			fmt.Println(line)
		}
	}
	return nanosSamples, scanner.Err()
}

type quartiles struct {
	samples []int
	p100    int
	p99     int
	p90     int
	p50     int
	p10     int
	p01     int
	p00     int
}

func (q *quartiles) String() string {
	return fmt.Sprintf("P100: %d \nP99: %d \nP90: %d \nP50: %d \nP10: %d \nP1: %d \nP0%d \n", q.p100, q.p99, q.p90, q.p50, q.p10, q.p01, q.p00)
}

func newQuartiles(samples []int) *quartiles {
	if len(samples) == 0 {
		return &quartiles{}
	}

	sort.Ints(samples)

	return &quartiles{
		samples: samples,
		p100:    getPercentile(samples, 100),
		p99:     getPercentile(samples, 99),
		p90:     getPercentile(samples, 90),
		p50:     getPercentile(samples, 50),
		p10:     getPercentile(samples, 10),
		p01:     getPercentile(samples, 1),
		p00:     getPercentile(samples, 0),
	}
}

func getPercentile(samples []int, percentile float64) int {
	floatLen := float64(len(samples) - 1)
	idx := int(floatLen * (percentile / 100))
	return samples[idx]
}

package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkScanRange(b *testing.B) {
	wg := sync.WaitGroup{}
	count := atomic.Int64{}

	for n := 0; n < b.N; n++ {
		wg.Add(1)
		scanRange(1, int64(n), int64(n+1), &count, &wg)
	}
}

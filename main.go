package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

var start = flag.Int64("start", 0, "start timestamp")
var end = flag.Int64("end", time.Now().Unix(), "end timestamp")
var debug = flag.Bool("debug", false, "print all timestamps processed")
var chunks = flag.Int("chunks", runtime.NumCPU(), "how many parts to process at once")

func main() {
	flag.Parse()

	// TODO: Investigate the suspiciously round numbers
	beginRange := *start
	endRange := *end
	span := endRange - beginRange

	chunkSize := int64(float64(span) / float64(*chunks))
	log.Printf("Dividing %d (%d–%d) timestamps into %d chunks of size %d", span, beginRange, endRange, *chunks, chunkSize)

	if profile, _ := os.LookupEnv("PROFILE"); profile != "" {
		filename := fmt.Sprintf("profile_%d-%d.pprof", beginRange, endRange)
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	count := atomic.Int64{}

	for i := 0; i < *chunks; i += 1 {
		wg.Add(1)
		start := beginRange + (int64(i) * chunkSize)
		end := start + chunkSize - 1
		log.Printf("Spawning goroutine %d from %d to %d (chunk size %d)", i, start, end, chunkSize)
		go scanRange(i, start, end, &count, &wg)
	}

	wg.Wait()
	total := count.Load()
	log.Printf("Total: %d/%d (%f)", total, span, float64(total)/float64(span))
}

func scanRange(id int, start int64, end int64, count *atomic.Int64, wg *sync.WaitGroup) {
	timestamp := uint32(start)

	numericalCount := 0
	passes := 0

	var b [4]byte
	for {
		if timestamp >= uint32(end) {
			log.Printf("Goroutine %d reached the end (%d). Found numerical hexes in %d/%d (%f) passes.", id, timestamp, numericalCount, passes, float64(numericalCount)/float64(passes))
			count.Add(int64(numericalCount))
			wg.Done()
			return
		}

		binary.BigEndian.PutUint32(b[0:4], timestamp)
		encoded := make([]byte, 8)
		hex.Encode(encoded, b[:])

		if containsLetter(encoded) {
			// Found an alphabetical character, skip this ID altogether
			passes += 1
			timestamp += 1
			continue
		}

		// No alphabetical characters encountered
		if *debug {
			fmt.Printf("%s\t%d\t%08b\n", string(encoded), timestamp, timestamp)
		}
		numericalCount += 1
		passes += 1
		timestamp += 1
	}
}

func containsLetter(id []byte) bool {
	for _, ch := range id {
		if ch >= 97 && ch <= 122 {
			return true
		}
	}

	return false
}

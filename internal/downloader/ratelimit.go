package downloader

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

type rateLimitedReader struct {
	reader      io.Reader
	rateLimit   int64 // bytes per second
	chunkBytes  int64
	chunkStart  time.Time
	sleepAdjust time.Duration
}

func newRateLimitedReader(r io.Reader, limit int64) *rateLimitedReader {
	return &rateLimitedReader{reader: r, rateLimit: limit}
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
	if r.rateLimit <= 0 {
		return 0, errors.New("invalid rate limit value")
	}

	// edge case if the requested rateLimit is smaller than the buffer
	if r.rateLimit < int64(len(p)) {
		p = p[:r.rateLimit]
	}

	// total bytes downloaded in a time window
	n, err := r.reader.Read(p)
	if n > 0 {
		r.Limit(n)
	}
	return n, err
}

func (r *rateLimitedReader) Limit(bytesRead int) {
	// total bytes downloaded in this time window
	r.chunkBytes += int64(bytesRead)
	deltaT := time.Since(r.chunkStart)

	expected := time.Duration(float64(r.chunkBytes) / float64(r.rateLimit) * float64(time.Second))

	if expected > deltaT {
		sleep_time := expected - deltaT + r.sleepAdjust

		t0 := time.Now()
		time.Sleep(sleep_time)
		actualSleep := time.Since(t0)

		r.sleepAdjust = sleep_time - actualSleep

		// normalize adjustement
		if r.sleepAdjust > 500*time.Millisecond {
			r.sleepAdjust = 500 * time.Millisecond
		} else if r.sleepAdjust < -500*time.Millisecond {
			r.sleepAdjust = -500 * time.Millisecond
		}
	}
	r.chunkBytes = 0
	r.chunkStart = time.Now()
}

func parseRateLimit(s string) int64 {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasSuffix(s, "k") {
		n, _ := strconv.ParseInt(s[:len(s)-1], 10, 64)
		return n * 1024
	}
	if strings.HasSuffix(s, "m") {
		n, _ := strconv.ParseInt(s[:len(s)-1], 10, 64)
		return n * 1024 * 1024
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

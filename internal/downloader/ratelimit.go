package downloader

import (
	"io"
	"strconv"
	"strings"
	"time"
)

type rateLimitedReader struct {
	reader    io.Reader
	rateLimit int64 // bytes per second
}

func newRateLimitedReader(r io.Reader, limit int64) *rateLimitedReader {
	return &rateLimitedReader{reader: r, rateLimit: limit}
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
    
}

func parseRateLimit(s string) int64 {
    s = strings.TrimSpace(s)
    if strings.HasSuffix(s, "k") {
        n, _ := strconv.ParseInt(s[:len(s)-1], 10, 64)
        return n * 1024
    }
    if strings.HasSuffix(s, "M") {
        n, _ := strconv.ParseInt(s[:len(s)-1], 10, 64)
        return n * 1024 * 1024
    }
    n, _ := strconv.ParseInt(s, 10, 64)
    return n
}

package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"
)

func Download(cfg *Config, url, basePath string) {
	fmt.Printf("start at %v\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Print("sending request, waiting for response...")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%v \n", err)
	}
	fmt.Printf(" status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	defer resp.Body.Close()

	var output_file_name string
	if cfg.OutputName == "" {
		parts := strings.Split(url, "/")
		output_file_name = parts[len(parts)-1]
	} else {
		output_file_name = cfg.OutputName
	}

	fileSavePath := filepath.Join(basePath, output_file_name)

	dir := filepath.Dir(fileSavePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Printf("failed to create directory %s: %v\n", dir, err)
		return
	}

	file, err := os.Create(fileSavePath)
	if err != nil {
		fmt.Printf("%v \n", err)
	}

	var reader io.Reader = resp.Body
	if cfg.RateLimit != "" {
		reader = newRateLimitedReader(reader, parseRateLimit(cfg.RateLimit))
	}

	if err := copyFile(url, cfg.OutputPath, file, reader, resp.ContentLength); err != nil {
		fmt.Printf("%v \n", err)
	}
}

func copyFile(url, savePath string, dst io.Writer, src io.Reader, total int64) error {
	buf := make([]byte, 64*1024)
	startTime := time.Now()
	var downloaded int64

	fmt.Printf("content size: %d [~%.2fMB]\n", total, float64(total)/1024/1024)
	fmt.Printf("saving file to: ./%s\n", savePath)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			_, err := dst.Write(buf[:n])
			if err != nil {
				return err
			}
			downloaded += int64(n)
			printProgress(downloaded, total, startTime)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n\nDownloaded [%s]\n", url)
	fmt.Printf("finished at %v", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func printProgress(downloaded, total int64, start time.Time) {
	percent := float64(downloaded) / float64(total) * 100
	elapsed_time := time.Since(start).Seconds()
	speed := float64(downloaded) / elapsed_time
	remaining_time := time.Duration(float64(total-downloaded)/speed) * time.Second

	barWidth := getBarWidth()
	filled := int(float64(barWidth) * float64(downloaded) / float64(total))
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)

	dl_str := formatSize(downloaded)
	tot_str := formatSize(total)
	speed_str := formatSize(int64(speed)) + "/s"

	var time_str string
	if remaining_time < time.Second {
		time_str = "0s"
	} else {
		time_str = formatDuration(remaining_time)
	}

	fmt.Printf("\r %s / %s [%s] %.2f%% %s %s",
		dl_str, tot_str, bar, percent, speed_str, time_str)
}

func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())

	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7

	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)

	case minutes < 60:
		return fmt.Sprintf("%dm", minutes)

	case hours < 24:
		return fmt.Sprintf("%dh %dm", hours, minutes-(hours*60))

	case days < 7:
		return fmt.Sprintf("%dd %dh", days, hours-(days*24))

	default:
		return fmt.Sprintf("%dw", weeks)
	}
}

func formatSize(b int64) string {
	if b >= 1024*1024 {
		return fmt.Sprintf("%d MiB", b/1024/1024)
	}
	return fmt.Sprintf("%d KiB", b/1024)
}

func getBarWidth() int {
	// Get the terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 40
	}

	reservedSpace := 60
	barWidth := width - reservedSpace

	if barWidth < 10 {
		return 10
	}
	return barWidth
}

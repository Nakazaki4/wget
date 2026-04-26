package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"
)

func Download(cfg *Config) {
	for _, url := range cfg.URLs {
		func() {
			fmt.Printf("start at %v\n", time.Now().Format("2006-01-02 15:04:05"))
			fmt.Print("sending request, waiting for response...")
			resp, err := http.Get(url)
			fmt.Printf(" status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
			if err != nil {
				fmt.Printf("%v \n", err)
			}
			defer resp.Body.Close()

			if cfg.OutputName == "" {
				parts := strings.Split(url, "/")
				cfg.OutputName = parts[len(parts)-1]
			}

			savePath := filepath.Join(cfg.OutputPath, cfg.OutputName)
			file, err := os.Create(savePath)
			if err != nil {
				fmt.Printf("%v \n", err)
			}

			if err := copyFile(url, savePath, file, resp.Body, resp.ContentLength); err != nil {
				fmt.Printf("%v \n", err)
			}
		}()
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
		time_str = fmt.Sprintf("%.0fs", remaining_time.Seconds())
	}

	fmt.Printf("\r %s / %s [%s] %.2f%% %s %s",
		dl_str, tot_str, bar, percent, speed_str, time_str)
}

func formatSize(b int64) string {
	if b >= 1024*1024 {
		return fmt.Sprintf("%.2f MiB", float64(b)/1024/1024)
	}
	return fmt.Sprintf("%.2f KiB", float64(b)/1024)
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

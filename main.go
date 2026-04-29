package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"wget/internal/downloader"
)

func main() {
	cfg := parseArgs()

	if len(cfg.URLs) == 0 && cfg.InputFile == "" {
		fmt.Println("At least one URL is required")
		os.Exit(1)
	}

	if cfg.Background {
		// need to fork the process
		downloader.HandleBackground()
		os.Exit(0)
	} else {
		commandExecute(&cfg)
	}
}

func commandExecute(cfg *downloader.Config) {
	if cfg.Mirror {
		// should route to the mirroring section
	} else {
		// route to the normal download section
		base_path := expandHomeDir(cfg.OutputPath)
		if cfg.InputFile != "" {
			file, err := os.Open(cfg.InputFile)
			if err != nil {
			}
			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				url := scanner.Text()
				// fmt.Println(url)
				downloader.Download(cfg, url, base_path)
			}
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading file:", err)
			}
		} else {
			for _, url := range cfg.URLs {
				downloader.Download(cfg, url, base_path)
			}
		}
	}
}

func expandHomeDir(path string) string {
	if len(path) > 1 && path[0] == '~' {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[1:])
	}
	return path
}

func parseArgs() downloader.Config {
	cfg := downloader.Config{}

	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			cfg.URLs = append(cfg.URLs, arg)
			continue
		}
		key, value, hasValue := strings.Cut(arg, "=")

		switch key {
		case "-B":
			cfg.Background = true
		case "-O":
			cfg.OutputName = value
		case "-P":
			cfg.OutputPath = value
		case "--rate-limit":
			cfg.RateLimit = value
		case "-i":
			cfg.InputFile = value
		case "--mirror":
			cfg.Mirror = true
		case "--reject":
			cfg.Reject = value
		case "--exclude":
			cfg.Exclude = value
		case "--convert-links":
			cfg.ConvertLinks = true
		default:
			if hasValue {
				fmt.Printf("Unknown option: %s\n", key)
			} else {
				fmt.Printf("Unknown flag: %s\n", arg)
			}
			os.Exit(1)
		}

		if strings.TrimSpace(value) == "" {
			continue
		}
	}
	return cfg
}

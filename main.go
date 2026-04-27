package main

import (
	"fmt"
	"os"
	"strings"

	"wget/internal/downloader"
)

func main() {
	cfg := parseArgs()

	if len(cfg.URLs) == 0 {
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
		downloader.Download(cfg)
	}
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
	}
	return cfg
}

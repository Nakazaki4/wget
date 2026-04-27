package downloader

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func HandleBackground() {
	logFile, err := os.Create("wget-log")
	if err != nil {
		log.Fatal(`can't create the log file "wget-log"`)
	}
	fmt.Println(`Output will be written to "wget-log"`)

	var newArgs []string
	for _, arg := range os.Args[1:] {
		if arg != "-B" {
			newArgs = append(newArgs, arg)
		}
	}

	cmd := exec.Command(os.Args[0], newArgs...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		log.Fatal("can't start background process")
	}
}

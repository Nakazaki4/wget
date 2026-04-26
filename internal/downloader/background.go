package downloader

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func HandleBackground(cfg *Config) {
	log_file, err := os.Create("wget-log")
	if err != nil {
		log.Fatal(`can't create the log file "wget-log"`)
	}
	fmt.Println(`Output will be written to "wget-log"`)

	os.Stdout = log_file
	defer log_file.Close()

	executable, err := os.Executable()
	if err != nil {
		log.Fatal("can't get the executable path")
	}

	newArgs := []string{}
	cmd := exec.Command(executable, )
	err = cmd.Start()
	if err != nil {
		log.Fatal("can't start the background process")
	}

}

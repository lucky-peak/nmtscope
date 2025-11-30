package nmt

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func GenerateNMTReport(jcmd string, pid int, reportDir string) {
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		os.MkdirAll(reportDir, 0755)
	}

	ts := time.Now().Unix()
	reportPath := fmt.Sprintf("%s/nmt_%d_%d.txt", reportDir, pid, ts)
	file, err := os.Create(reportPath)
	if err != nil {
		log.Fatalf("Failed to create report file: %v", err)
	}
	defer file.Close()

	// 执行 jcmd 命令
	cmd := exec.Command(jcmd, fmt.Sprintf("%d", pid), "VM.native_memory", "summary")
	cmd.Stdout = file
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run jcmd: %v", err)
	}
	log.Printf("Generated NMT report for pid %d at %d to %s", pid, ts, reportPath)
}

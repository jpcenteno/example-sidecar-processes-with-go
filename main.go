package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type PdfPreviewProcess struct {
	cmd *exec.Cmd
}

func OpenPdfPreview(filePath string) (*PdfPreviewProcess, error) {
	// Preliminary checks:

	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return nil, fmt.Errorf("Provided file is not a PDF: %s\n", filePath)
	}

	if _, err := exec.LookPath("zathura"); err != nil {
		return nil, fmt.Errorf("Command Zathura not found")
	}

	// Open Zathura:

	cmd := exec.Command("zathura", filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start zathura: %v", err)
	}

	return &PdfPreviewProcess{cmd: cmd}, nil
}

func (ppp *PdfPreviewProcess) Close() error {
	return syscall.Kill(-ppp.cmd.Process.Pid, syscall.SIGKILL) // Negative sign sends signal to whole process group.
}

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Printf("Usage: go run main.go <file-path>")
		os.Exit(1)
	}

	ppp, err := OpenPdfPreview(os.Args[1])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer ppp.Close()

	// Set up signal handling to clean up on SIGINT (Ctrl-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		ppp.Close()
		os.Exit(1)
	}()

	fmt.Println("Press enter to close the PDF previewer")
	fmt.Scanln()
}

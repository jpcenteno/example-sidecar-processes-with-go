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

// PdfPreviewProcess holds the command for the zathura process
type PdfPreviewProcess struct {
	cmd *exec.Cmd
}

// OpenPdfPreview starts a zathura process with the given file path
func OpenPdfPreview(filePath string) (*PdfPreviewProcess, error) {
	// Preliminary checks
	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return nil, fmt.Errorf("provided file is not a PDF: %s", filePath)
	}

	if _, err := exec.LookPath("zathura"); err != nil {
		return nil, fmt.Errorf("Command Zathura not found")
	}

	// Open Zathura
	cmd := exec.Command("zathura", filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start zathura: %v", err)
	}

	return &PdfPreviewProcess{cmd: cmd}, nil
}

// Close kills the zathura process
func (ppp *PdfPreviewProcess) Close() error {
	return syscall.Kill(-ppp.cmd.Process.Pid, syscall.SIGKILL) // Negative sign sends signal to whole process group.
}

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file-path>\n", os.Args[0])
		os.Exit(1)
	}

	ppp, err := OpenPdfPreview(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	fmt.Println("Press Enter to close the PDF previewer")
	fmt.Scanln()
}

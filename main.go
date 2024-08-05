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

// Close kills the zathura process
func (ppp *PdfPreviewProcess) Close() error {
	// Do nothing when no process is running. This avoids a nil pointer
	// dereference.
	if ppp.cmd == nil {
		return nil
	}

	// The negative sign broadcasts the signal to the whole process group
	err := syscall.Kill(-ppp.cmd.Process.Pid, syscall.SIGKILL)
	if err == nil {
		ppp.cmd = nil // Keep the `cmd` on error.
	}
	return err
}

func (ppp *PdfPreviewProcess) withPreview(filePath string, action func(string) error) error {
	// Preliminary checks
	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return fmt.Errorf("provided file is not a PDF: %s", filePath)
	}

	if _, err := exec.LookPath("zathura"); err != nil {
		return fmt.Errorf("Command Zathura not found")
	}

	// Open Zathura
	ppp.cmd = exec.Command("zathura", filePath)
	ppp.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := ppp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start zathura: %v", err)
	}

	defer ppp.Close()

	return action(filePath)
}

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [FILE]...\n", os.Args[0])
		os.Exit(1)
	}

	ppp := PdfPreviewProcess{}

	// Set up a signal handler to clean up child processes on SIGINT (Ctrl-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		err := ppp.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close preview process: %v\n", err)
		}
		os.Exit(1)
	}()

	files := os.Args[1:]
	for _, file := range files {
		err := ppp.withPreview(file, func(file string) error {
			fmt.Println("Press Enter to close the PDF previewer for", file)
			fmt.Scanln()
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

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

// PreviewProcess holds the command for the zathura process
type PreviewProcess struct {
	cmd *exec.Cmd
}

// Close kills the zathura process
func (pp *PreviewProcess) Close() error {
	// Do nothing when no process is running. This avoids a nil pointer
	// dereference.
	if pp.cmd == nil {
		return nil
	}

	// The negative sign broadcasts the signal to the whole process group
	err := syscall.Kill(-pp.cmd.Process.Pid, syscall.SIGKILL)
	if err == nil {
		pp.cmd = nil // Keep the `cmd` on error.
	}
	return err
}

func (pp *PreviewProcess) withPreview(filePath string, action func(string) error) error {
	// Preliminary checks
	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return fmt.Errorf("provided file is not a PDF: %s", filePath)
	}

	if _, err := exec.LookPath("zathura"); err != nil {
		return fmt.Errorf("Command Zathura not found")
	}

	// Open Zathura
	pp.cmd = exec.Command("zathura", filePath)
	pp.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := pp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start zathura: %v", err)
	}

	defer pp.Close()

	return action(filePath)
}

func main() {
	// Check if the file path is provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [FILE]...\n", os.Args[0])
		os.Exit(1)
	}

	pp := PreviewProcess{}

	// Set up a signal handler to clean up child processes on SIGINT (Ctrl-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		err := pp.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close preview process: %v\n", err)
		}
		os.Exit(1)
	}()

	files := os.Args[1:]
	for _, file := range files {
		err := pp.withPreview(file, func(file string) error {
			fmt.Println("Press Enter to close the PDF previewer for", file)
			fmt.Scanln()
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

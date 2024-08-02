package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// FIXME: Open Zathura.

	return nil, fmt.Errorf("OpenPdfPreview: Not implemented")
}

func (ppp *PdfPreviewProcess) Close() error {
	return fmt.Errorf("Close: Not implemented")
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

	fmt.Println("Press enter to close the PDF previewer")
	fmt.Scanln()
}

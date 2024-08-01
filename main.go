package main

import (
	"fmt"
	"os"
	"os/exec"
)

type PdfPreviewProcess struct {
	cmd *exec.Cmd
}

func OpenPdfPreview(_ string) (*PdfPreviewProcess, error) {
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

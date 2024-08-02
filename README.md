# Experiment: opening a PDF preview in GO

## Development log

### Created a Nix flake with a devshell for Go

```nix
devShells.default = pkgs.mkShell {
  buildInputs = with pkgs; [ go gopls gotools go-tools ];
};
```

### Initialized a Go module

I learned that I don't need a URL to initialize a Go module.

```sh
go mod init experiment-open-pdf-preview-in-golang
```

### Created a _Hello World_ to check that the nix development shell works

```go
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}
```

```sh
go run main.go
hello world
```

Now, I'm sure that the development shell works.

### Downloaded some sample PDFs

I will need a couple of PDFs to test the program. Luckily, I found some at the
[University of Waterloo sample PDF documents](https://uwaterloo.ca/onbase/help/sample-pdf-documents)
page (Cool resource by the way). I downloaded two of them into a new `pdfs/`
directory:

```sh
ls -1 pdfs
sample-certified-pdf-by-the-waterloo-university.pdf
sample-unsecure-pdf-by-the-waterloo-university.pdf
```

### Implement the process struct skeleton

I could have done a simpler thing for the purposes of this experiment, but I
decided to build a struct to wrap the Open and close logic.

```go
type PdfPreviewProcess struct {
	cmd *exec.Cmd
}

func OpenPdfPreview(_ string) (*PdfPreviewProcess, error) {
	return nil, fmt.Errorf("OpenPdfPreview: Not implemented")
}

func (ppp *PdfPreviewProcess) Close() error {
	return fmt.Errorf("Close: Not implemented")
}
```

Which will be used like this:

```go
func main() {
	// ...

	ppp, err := OpenPdfPreview(os.Args[1])
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer ppp.Close()

	fmt.Println("Press enter to close the PDF previewer")
	fmt.Scanln()
}
```

### Preliminary checks before opening Zathura

To preview a PDF, we need to make sure that:
- The file is indeed a PDF
- That Zathura is present in the `$PATH`.

```go
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
```

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

### Opening zathura

I added the following code to `OpenPdfPreview` after making sure that the
preliminary checks pass:

```go
// Open Zathura:

cmd := exec.Command("zathura", filePath)
if err := cmd.Start(); err != nil {
    return nil, fmt.Errorf("failed to start zathura: %v", err)
}

return &PdfPreviewProcess{cmd: cmd}, nil
```

This succeeds in opening the PDF preview.

If I kill the Go program using a keyboard interrupt `C-c`, the child Zathura
process dies with it, but it's not the same case when I press enter. What's
currently missing is to:

What's currently missing is to:
- Add a signal trap that kills the child process if it still lives when the
  program exits.
- Kill the previewer process "gracefully" when the user presses enter.

### Implementing `PdfPreviewProcess.Close`

```go
func (ppp *PdfPreviewProcess) Close() error {
	return syscall.Kill(ppp.cmd.Process.Pid, syscall.SIGKILL)
}
```

### Research: What can I learn from the PID of the program?

On a terminal, I ran:

```sh
$ go run main.go pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
```

In a separate terminal, I ran:
```sh
$ ps aux | grep 'go run main.go' | grep -v grep
<username>+   28975  0.0  0.1 1240760 16056 pts/2   Sl+  21:03   0:00 go run main.go pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
# USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
```

The pid is `28975`.

What's in it's group?

```sh
$ ps -o pid,pgid,comm 28975
PID    PGID  COMMAND
28975  28975 go
```

So, this process is the leader of it's own group.

```sh
$ pstree -- -g -p 28975
-+= 28975 <username> go run main.go pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
 \-+- 29134 <username> /run/user/1000/tmp/go-build1028005249/b001/exe/main pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
   \--- 29139 <username> /nix/store/i1i0a085whvl8dfl8h615lpp44h5myvl-zathura-0.5.2-bin/bin/zathura pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
```

#### What happens if we give the new process it's own group?

While Zathura seems to spawn a unique process, other programs might create more.
It's  a good idea to wrap the auxiliar process in a Group to have better control
when it's time to send it a kill signal.

These are the changes that I did to make the process have it's own group.

```diff
--- a/main.go
+++ b/main.go
@@ -27,6 +27,7 @@ func OpenPdfPreview(filePath string) (*PdfPreviewProcess, error) {
        // Open Zathura:

        cmd := exec.Command("zathura", filePath)
+       cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
        if err := cmd.Start(); err != nil {
                return nil, fmt.Errorf("failed to start zathura: %v", err)
        }
@@ -35,7 +36,7 @@ func OpenPdfPreview(filePath string) (*PdfPreviewProcess, error) {
 }

 func (ppp *PdfPreviewProcess) Close() error {
-       return syscall.Kill(ppp.cmd.Process.Pid, syscall.SIGKILL)
+       return syscall.Kill(-ppp.cmd.Process.Pid, syscall.SIGKILL) // Negative sign sends signal to whole process group.
 }

 func main() {
```

```sh
$ ps aux | grep 'go run main.go' | grep -v grep
<username>+   47588  0.0  0.0 1240504 14568 pts/2   Sl+  00:22   0:00 go run main.go pdfs/sample-certified-pdf-by-the-waterloo-university.pd
```

```sh
$ nix run nixpkgs#pstree -- -g -p 47588
-+= 47588 <username> go run main.go pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
 \-+- 47745 <username> /run/user/1000/tmp/go-build1733395772/b001/exe/main pdfs/sample-certified-pdf-by-the-waterloo-university.pdf
   \--= 47750 <username> /nix/store/i1i0a085whvl8dfl8h615lpp44h5myvl-zathura-0.5.2-bin/bin/zathura pdfs/sample-certified-pdf-by-the-waterloo-uni
```

So far, everything looks the same.

```sh
$ ps -o pid,pgid,comm 47588
    PID    PGID COMMAND
  47588   47588 go

$ ps -o pid,pgid,comm 47745
    PID    PGID COMMAND
  47745   47588 main

$ ps -o pid,pgid,comm 47750
    PID    PGID COMMAND
  47750   47750 .zathura-wrappe
```

This is the intended result! `go run` and it's child program `main ...`
(`main.go`) belong to the same group, with `go run` being it's leader process,
while `zathura` belongs to a different group.

The change that I noticed is that the preview window survives the go program
after killing it with `C-c`, so that is something to fix.

### Killing the child process on SIGINT

I added goroutine that traps SIGINT and SIGTERM and kills the child process.
This ensures that the program does not leak the preview process on ungraceful
exit.

```go
// Open the preview process.

// Set up signal handling to clean up on SIGINT (Ctrl-C)
c := make(chan os.Signal, 1)
signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
go func() {
	<-c
	ppp.Close()
	os.Exit(1) // Exit so that the program stops expecting the user to press enter.
}()

// Wait for the user to press the enter key.
```

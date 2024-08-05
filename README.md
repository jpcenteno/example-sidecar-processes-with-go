# Sidecar Processes with Go Example

This repository contains a standalone example of a Go program that opens a
preview process while doing some other work. This functionality is part of a
larger, private project I am working on. As I am new to Go, I created this as a
self-contained example to share, document and keep for future reference.

## Example program

This example program loops over a list of input files while opening a
user-specified preview program while it runs a function (which in my use case,
waits for user input). Finally, it kills the preview process group when work is
done.

## Usage

```
go run main.go <preview-command> <file1> <file2> ...
```

I added a directory called `sample-files/` which you can use to test the
program.

Here is an example using [Zathura](https://pwmt.org/projects/zathura/),
a document viewer for Linux to preview PDFs.

```sh
go run main.go zathura sample-files/*.pdf
```

Here is another example that uses a preview script that can handle either jpg or
pdf files:

```sh
go run main.go ./preview.sh sample-files/*
```

## Features

- No dependencies other than the Go standard library.
- Accepts an arbitrary preview command and a list of files to preview
  sequentially.
- Runs a sample job as parameter to run while the previewer is open.
- Kills the preview process group when the work is done.
- Provides an example interruption handler that cleans up the child process
  group.

## Pending work

- [ ] Handle restarts when the user accidentally closes the preview process.
  This is the case where the exit code is 0.
- [ ] Handle the case where the preview process fails and exits with non-zero
  exit code.
- [ ] Try it on Mac OS.
    - I don't have a Mac OS computer at hand, but I guess that you could try it
      with the `open` command. I don't know whether killing the process group
      will work or not. Please tell me if you try it.

## Credits for the sample files

- `sample-files/cat.pdf` was authored by Wikipedia user
  [Von.grzanka](https://en.wikipedia.org/wiki/User:Von.grzanka) and licensed
  under the
  [CC BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/deed.en)
  license. Here is a [link to the image](https://en.wikipedia.org/wiki/File:Felis_catus-cat_on_snow.jpg).
- The PDF files were sourced from the
  [University of Waterloo sample PDF documents](https://uwaterloo.ca/onbase/help/sample-pdf-documents).

A big thanks to them for sharing those files with the world!

## Development log

FIXME clean up this draft. Sorry for the typos.

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

### Opening a sequence of files in a loop

One of the use cases from the original project requires previewing a sequence of
files in a loop. Changing the program was very simple:

```go
files := os.Args[1:]
for _, file := range files {
	ppp, err := OpenPdfPreview(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

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
	ppp.Close()
}
```

```sh
go run main.go pdfs/*
```

I learned that `defer` works at a function level, so I had to change that line
and put the call to `ppp.Close()` at the end of the loop.

#### The program structure might become messy

One thing that I notice now is that the program has the following structure:

```
for _, file := range files {
    // Open preview
    // Start interrupt handler

    // Do something with the preview on

    // Close preview
//
}
```

Comming from a FP background, I'm tempted to refactor the loops body into a
higher order function to ensure that the "do something" part does not mess up
with the rest of the setup and teardown code. I will keep it this way for now,
but might do something about it in the future.

#### Preview process corner cases:

- [ ] How to handle the case where the user accidentally closes the preview
  window? What should the program do when the preview window closes due to a
  failure vs when the user closes the window?
- [ ] How do we know that we are killing the right process? (Race condition
  where the process dies and the OS creates an unrelated process group with the
  same PGID) How probable is this scenario?

### Refactoring the preview function to separate process logic from application logic

After noticing that the main loop was mixing process and application logic I
gave it a thought and decided to abstract it away. My solution was to replace
the `openPdfPreview` function with a higher order function called
`withPdfPreview`. The new function extends the previous code with a new
parameter of the type `func(string) error`. The closure receives the
file path as parameter and is expected to implement the application logic.

```go
func withPreview(filePath string, action func(string) error) error {
	// Ensure that fipePath is a PDF.
	// Ensure that the Zathura command is available.

	// Open Zathura

	ppp := &PdfPreviewProcess{cmd: cmd}
	defer ppp.Close() // This will run after return.

	// Set up signal handling to clean up on SIGINT (Ctrl-C)

	return action(filePath) // Runs application logic.
}
```

This change hides all the preview process logic from the main loop:

```go
files := os.Args[1:]
for _, file := range files {
	err := withPreview(file, func(file string) error {
		fmt.Println("Press Enter to close the PDF previewer for", file)
		fmt.Scanln()
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
```

### Fixing some problems with the `Close` method

During manual testing, I've noticed that the `Close()` method raised a
segmentation fault due to a `nil` pointer dereference if called before creating
the preview process. This is the case where `cmd * exec.Cmd` equals `nil`.

I fixed this with a short-circuiting nil check.

Another thing that I forgot about was to clean up `cmd` after broadcasting the
`SIGKILL`. This comes with the added feature of allowing the reuse of the
`PdfPreviewProcess` structure.

```go
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
```

### Interrupt handler goroutine leakage

I'm under the suspicion that each call to `withPreview` is creating a new signal
handler goroutine without properly stoping it before returning. To verify this,
I added these two log lines:

```go
fmt.Fprintln(os.Stderr, "Starting signal handler for file: ", filePath)
go func() {
	<-c
	ppp.Close()
	fmt.Fprintln(os.Stderr, "Stoping signal handler for file: ", filePath)
	os.Exit(1)
}()
```

This allowed me to verify that those processes live until the program ends. This
can be a problem given a large list of files. There are two ways that come to my
mind in order to clean up those processes:

- Add a `stopChannel` that gracefully stops the signal handler.
- Move the signal handler away from the `withPreview` function. This solves the
  aditional problem that might arise when the program has another signal handler
  that does more than just closing the previewer process.

I decided that the best solution is to leave the interruption handling to the
caller. This prevents both the goroutine leakage and race conditions between the
libraries interrupt handlers and any other interrupt handler that the user might
implement. This increments flexibility with some added burden, but I think that
it is a good trade off.

To fix this I needed to convert `withPreview` into a method. This is necessary
to allow outside access to the `PdfPreviewProcess.Close()` method.

Then, I moved the signal handler to the main loop. This forces the user to
implement a handler, but gives them the opportunity to extend it's behavior.

Now, `main()` looks like this:

```go
func main() {
	// ...

	ppp := PdfPreviewProcess{}

	// Set up a signal handler to clean up child processes on SIGINT (Ctrl-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		ppp.Close()
		os.Exit(1)
	}()

	files := os.Args[1:]
	for _, file := range files {
		// Call to ppp.withPreview
	}
}
```

And now, `withPreview` looks like this.

```go
func (ppp *PdfPreviewProcess) withPreview(filePath string, action func(string) error) error {
	// Preliminary checks

	// Open Zathura
	ppp.cmd = exec.Command("zathura", filePath)
	ppp.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := ppp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start zathura: %v", err)
	}

	defer ppp.Close()

	return action(filePath)
}
```

### Running a generic preview script

I wanted to simplify the program by replacing Zathura with a generic preview
program supplied by the user. This moves some of the complexity to an external
program and simplifies the code providing better readability.

- I renamed `PdfPreviewProcess` to `PreviewProcess`.
- I added a field called `programName`.

For the user, the only change is that the constructor requires a `programName`
to be set. The rest of the API remains the same.

```go
func main() {
	// Argument check.
    // ...

	programName := os.Args[1]
	files := os.Args[2:]

	pp := PreviewProcess{programName: programName}

	// Set up a signal handler to clean up child processes on SIGINT (Ctrl-C)
    // ...

	for _, file := range files {
		// Unchanged
	}
}
```

The `withPreview` function became simpler now that the preliminary checks are
responsibility of the preview program.

```go
func (pp *PreviewProcess) withPreview(filePath string, action func(string) error) error {
	pp.cmd = exec.Command(pp.programName, filePath)
	pp.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Give the process it's own group.
	if err := pp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %v", pp.programName, err)
	}

	defer pp.Close()

	return action(filePath)
}
```

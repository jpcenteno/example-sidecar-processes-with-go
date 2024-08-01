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

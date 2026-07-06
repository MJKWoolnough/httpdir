# httpdir

[![CI](https://github.com/MJKWoolnough/httpdir/actions/workflows/go-checks.yml/badge.svg)](https://github.com/MJKWoolnough/httpdir/actions)
[![Go Reference](https://pkg.go.dev/badge/vimagination.zapto.org/httpdir.svg)](https://pkg.go.dev/vimagination.zapto.org/httpdir)

--
    import "vimagination.zapto.org/httpdir"

Package httpdir provides an in-memory implementation of http.FileSystem.

## Highlights

 - Implements http.FileSystem for seamless integration with Go’s net/http.
 - Supports creating, removing, and opening files and directories in memory.
 - Provides convenience functions for working with byte slices, strings, and OS-based files.
 - Allows directory listings with optional index.html overrides.

## Usage

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"vimagination.zapto.org/httpdir"
)

func main() {
	dir := httpdir.New(time.Now())

	dir.Create("index.html", httpdir.FileString("<html><head><title>Example</title></head><body><h1>Hello from httpdir!</h1></body></html>", time.Now()))
	dir.Create("style.css", httpdir.FileString("body { background: #f0f0f0; }", time.Now()))

	http.Handle("/", http.FileServer(dir))

	srv := httptest.NewServer(http.DefaultServeMux)

	resp, err := srv.Client().Get(srv.URL + "/")
	if err != nil {
		fmt.Println(err)

		return
	}

	io.Copy(os.Stdout, resp.Body)

	// Output:
	// <html><head><title>Example</title></head><body><h1>Hello from httpdir!</h1></body></html>
}
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/vimagination.zapto.org/httpdir

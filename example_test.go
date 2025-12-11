package httpdir_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"vimagination.zapto.org/httpdir"
)

func Example() {
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

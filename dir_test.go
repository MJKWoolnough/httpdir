package httpdir

import (
	"io/fs"
	"testing"
	"time"
)

func TestDir(t *testing.T) {
	d := New(time.Now())
	err := d.Mkdir("/dir", time.Now(), true)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	err = d.Mkdir("/dir2", time.Now(), false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	_, err = d.Open("/dir")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	_, err = d.Open("/dir2")
	if err != fs.ErrPermission {
		t.Errorf("expecting permission error, got %q", err)
	}
	err = d.Create("/dir3/test/hello", FileString("Hello, World!", time.Now()))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	f, err := d.Open("/dir3/test/hello")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	stat, err := f.Stat()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if stat.Name() != "hello" {
		t.Errorf("expecting name \"hello\", got %q", stat.Name())
	}
	data := make([]byte, 20)
	n, err := f.Read(data)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if n != 13 {
		t.Errorf("expecting to read 13 bytes, read %d", n)
	}
	if string(data[:n]) != "Hello, World!" {
		t.Errorf("expecting string \"Hello, World!\", got %q", data[:n])
	}
	err = d.Remove("/dir3/test/hello")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if len(d.d.contents["dir3"].(dir).contents["test"].(dir).contents) != 0 {
		t.Errorf("did not delete '/dir3/test/hello'")
		return
	}
	err = d.Remove("/dir3")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if len(d.d.contents) != 2 { // dir && dir2 remain
		t.Errorf("did not delete '/dir3'")
	}
}

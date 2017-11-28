package httpdir

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestDirectory(t *testing.T) {
	d := dir{}
	_, err := d.Open()
	if err != os.ErrPermission {
		t.Errorf("expecting permission error, got: %s", err)
	}
	if d.Mode() != ModeDir {
		t.Errorf("expecting mode %v, got %v", ModeDir, d.Mode())
	}
	mt := time.Now()
	d = dir{
		index: true,
		contents: map[string]Node{
			"file1": FileString("Hello, World!", mt.Add(-1*time.Hour)),
			"file2": FileString("FooBarBaz", mt.Add(-2*time.Hour)),
		},
		modTime: mt,
	}
	if !d.ModTime().Equal(mt) {
		t.Errorf("expecting modTime %v, got %v", mt, d.ModTime())
	}
	f, err := d.Open()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	c, err := f.Readdir(-1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(c) != 2 {
		t.Errorf("expecting 2 items, got %d", len(c))
		return
	}
	if c[0].Name() != "file1" {
		t.Errorf("expecting file with name \"file1\", got %q", c[0].Name())
	}
	if c[1].Name() != "file2" {
		t.Errorf("expecting file with name \"file2\", got %q", c[1].Name())
	}
	_, err = f.Readdir(1)
	if err != io.EOF {
		t.Errorf("expected error EOF, got %s", err)
		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	c, err = f.Readdir(1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(c) != 1 {
		t.Errorf("expecting 1 item, got %d", len(c))
		return
	}
	if c[0].Name() != "file1" {
		t.Errorf("expecting file with name \"file1\", got %q", c[0].Name())
	}
	c, err = f.Readdir(1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(c) != 1 {
		t.Errorf("expecting 1 item, got %d", len(c))
		return
	}
	if c[0].Name() != "file2" {
		t.Errorf("expecting file with name \"file2\", got %q", c[0].Name())
	}
	_, err = f.Readdir(1)
	if err != io.EOF {
		t.Errorf("expected error EOF, got %s", err)
		return
	}
}

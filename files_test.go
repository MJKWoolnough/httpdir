package httpdir

import (
	"testing"
	"time"
)

func TestBytes(t *testing.T) {
	mt := time.Now()
	b := FileBytes([]byte("Hello, World!"), mt)

	if !b.ModTime().Equal(mt) {
		t.Errorf("expecting time %v, got %v", mt, b.ModTime())
	}

	if b.Mode() != ModeFile {
		t.Errorf("expecting mode %v, got %v", ModeFile, b.Mode())
	}

	if b.Size() != 13 {
		t.Errorf("expecting size 13, got %d", b.Size())
	}

	f, err := b.Open()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	tr := make([]byte, 6)
	n, err := f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 6 {
		t.Errorf("expecting to read 6 bytes, read %d byte(s)", n)
	}
	if string(tr[:n]) != "Hello," {
		t.Errorf("expecting to have read \"Hello,\", read %s", tr)
	}

	n, err = f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 6 {
		t.Errorf("expecting to read 6 bytes, read %d byte(s)", n)
	}
	if string(tr[:n]) != " World" {
		t.Errorf("expecting to have read \" World\", read %s", tr)
	}

	n, err = f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 1 {
		t.Errorf("expecting to read 1 byte, read %d byte(s)", n)
	}
	if string(tr[:n]) != "!" {
		t.Errorf("expecting to have read \"!\", read %s", tr)
	}
}

func TestString(t *testing.T) {
	mt := time.Now()
	s := FileString("Hello, World!", mt)

	if !s.ModTime().Equal(mt) {
		t.Errorf("expecting time %v, got %v", mt, s.ModTime())
	}

	if s.Mode() != ModeFile {
		t.Errorf("expecting mode %v, got %v", ModeFile, s.Mode())
	}

	if s.Size() != 13 {
		t.Errorf("expecting size 13, got %d", s.Size())
	}

	f, err := s.Open()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	tr := make([]byte, 6)
	n, err := f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 6 {
		t.Errorf("expecting to read 6 bytes, read %d byte(s)", n)
	}
	if string(tr[:n]) != "Hello," {
		t.Errorf("expecting to have read \"Hello,\", read %s", tr)
	}

	n, err = f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 6 {
		t.Errorf("expecting to read 6 bytes, read %d byte(s)", n)
	}
	if string(tr[:n]) != " World" {
		t.Errorf("expecting to have read \" World\", read %s", tr)
	}

	n, err = f.Read(tr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if n != 1 {
		t.Errorf("expecting to read 1 byte, read %d byte(s)", n)
	}
	if string(tr[:n]) != "!" {
		t.Errorf("expecting to have read \"!\", read %s", tr)
	}
}

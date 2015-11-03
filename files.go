package httpdir

import (
	"bytes"
	"os"
	"strings"
	"time"
)

type fileBytes struct {
	data    []byte
	modTime time.Time
}

func FileBytes(data []byte, modTime time.Time) Node {
	return fileBytes{
		data,
		modTime,
	}
}

func (f fileBytes) Size() int64 {
	return int64(len(f.data))
}

func (fileBytes) Mode() os.FileMode {
	return ModeFile
}

func (f fileBytes) ModTime() time.Time {
	return f.modTime
}

func (f fileBytes) Open() (File, error) {
	return fileBytesOpen{bytes.NewReader(f.data)}, nil
}

type fileBytesOpen struct {
	*bytes.Reader
}

func (fileBytesOpen) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (fileBytesOpen) Close() error {
	return nil
}

type fileString struct {
	data    string
	modTime time.Time
}

func FileString(data string, modTime time.Time) Node {
	return fileString{
		data,
		modTime,
	}
}

func (f fileString) Size() int64 {
	return int64(len(f.data))
}

func (fileString) Mode() os.FileMode {
	return ModeFile
}

func (f fileString) ModTime() time.Time {
	return f.modTime
}

func (f fileString) Open() (File, error) {
	return fileStringOpen{strings.NewReader(f.data)}, nil
}

type fileStringOpen struct {
	*strings.Reader
}

func (fileStringOpen) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (fileStringOpen) Close() error {
	return nil
}

type OSFile string

func (o OSFile) Size() int64 {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}
	return s.Size()
}

func (o OSFile) Mode() os.FileMode {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}
	return s.Mode()
}

func (o OSFile) ModTime() time.Time {
	s, err := os.Stat(string(o))
	if err != nil {
		return time.Time{}
	}
	return s.ModTime()
}

func (o OSFile) Open() (File, error) {
	return os.Open(string(o))
}

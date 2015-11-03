package httpdir

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Dir map[string]File

func (d Dir) Open(name string) (http.File, error) {
	f, ok := d[name]
	if !ok {
		return nil, NotFound
	}
	return f.Open()
}

type File interface {
	Open() (http.File, error)
}

type stat struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (s *stat) Stat() (os.FileInfo, error) {
	return s, nil
}

func (s *stat) Name() string {
	return s.name
}

func (s *stat) Size() int64 {
	return s.size
}

func (s *stat) Mode() os.FileMode {
	return s.mode
}

func (s *stat) ModTime() time.Time {
	return s.modTime
}

func (s *stat) IsDir() bool {
	return s.mode.IsDir()
}

func (s *stat) Sys() interface{} {
	return s
}

type file interface {
	io.Closer
	io.Reader
	io.Seeker
}

type fileWrapper struct {
	file
	*stat
}

func (fileWrapper) Readdir(int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

type fileBytes struct {
	data []byte
	stat
}

func FileBytes(name string, data []byte, modTime time.Time) File {
	return &fileBytes{
		data,
		stat{
			name,
			int64(len(data)),
			0644,
			modTime,
		},
	}
}

type bytesCloser struct {
	*bytes.Reader
}

func (bytesCloser) Close() error {
	return nil
}

func (f *fileBytes) Open() (http.File, error) {
	return fileWrapper{
		bytesCloser{bytes.NewReader(f.data)},
		&f.stat,
	}, nil
}

type fileString struct {
	data string
	stat
}

func FileString(name, data string, modTime time.Time) File {
	return &fileString{
		data,
		stat{
			name,
			int64(len(data)),
			0644,
			modTime,
		},
	}
}

type stringCloser struct {
	*strings.Reader
}

func (stringCloser) Close() error {
	return nil
}

func (f *fileString) Open() (http.File, error) {
	return &fileWrapper{
		stringCloser{strings.NewReader(f.data)},
		&f.stat,
	}, nil
}

type dirWrapper struct {
	*directory
	pos int
}

func (d *dirWrapper) Close() error {
	return nil
}

func (d *dirWrapper) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (d *dirWrapper) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (dirWrapper) Read([]byte) (int, error) {
	return 0, os.ErrInvalid
}

type directory struct {
	contents []string
	stat
}

func Directory(name string, contents []string, modTime time.Time) File {
	return &directory{
		contents,
		stat{
			name,
			0,
			os.ModeDir | 0755,
			modTime,
		},
	}
}

func (d *directory) Open() (http.File, error) {
	return &dirWrapper{directory: d}, nil
}

var NotFound = errors.New("file not found")

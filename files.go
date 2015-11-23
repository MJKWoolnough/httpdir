package httpdir

import (
	"bytes"
	"compress/gzip"
	"os"
	"strings"
	"time"
)

type fileBytes struct {
	data    []byte
	modTime time.Time
}

// FileBytes provides an implementation of Node that takes a byte slice as its
// data source
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

// FileString provides an implementation of Node that takes a string as its
// data source
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

// OSFile is the path of a file in the real filesystem to be put into the
// im-memory filesystem
type OSFile string

// Size returns the size of the file
func (o OSFile) Size() int64 {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}
	return s.Size()
}

// Mode returns the Mode of the file
func (o OSFile) Mode() os.FileMode {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}
	return s.Mode()
}

// ModTime returns the ModTime of the file
func (o OSFile) ModTime() time.Time {
	s, err := os.Stat(string(o))
	if err != nil {
		return time.Time{}
	}
	return s.ModTime()
}

// Open opens the file, returning it as a File
func (o OSFile) Open() (File, error) {
	return os.Open(string(o))
}

// Decompressed creates a new node that contains the decompressed gzip data
// from the given node
func Decompressed(node Node, size int) (Node, error) {
	f, err := node.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer g.Close()
	buf := make([]byte, size)

	var read int
	for read < size {
		n, err := g.Read(buf[read:])
		if err != nil {
			return nil, err
		}
		read += n
	}

	return FileBytes(buf, node.ModTime()), nil
}

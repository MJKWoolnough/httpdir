package httpdir

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"time"
)

type fileBytes struct {
	data    []byte
	modTime time.Time
}

// FileBytes provides an implementation of Node that takes a byte slice as its
// data source.
func FileBytes(data []byte, modTime time.Time) Node {
	return fileBytes{
		data,
		modTime,
	}
}

func (f fileBytes) Size() int64 {
	return int64(len(f.data))
}

func (fileBytes) Mode() fs.FileMode {
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

func (fileBytesOpen) Readdir(int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (fileBytesOpen) Close() error {
	return nil
}

type fileString struct {
	data    string
	modTime time.Time
}

// FileString provides an implementation of Node that takes a string as its
// data source.
func FileString(data string, modTime time.Time) Node {
	return fileString{
		data,
		modTime,
	}
}

func (f fileString) Size() int64 {
	return int64(len(f.data))
}

func (fileString) Mode() fs.FileMode {
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

func (fileStringOpen) Readdir(int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (fileStringOpen) Close() error {
	return nil
}

// OSFile is the path of a file in the real filesystem to be put into the
// im-memory filesystem.
type OSFile string

// Size returns the size of the file.
func (o OSFile) Size() int64 {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}

	return s.Size()
}

// Mode returns the Mode of the file.
func (o OSFile) Mode() fs.FileMode {
	s, err := os.Stat(string(o))
	if err != nil {
		return 0
	}

	return s.Mode()
}

// ModTime returns the ModTime of the file.
func (o OSFile) ModTime() time.Time {
	s, err := os.Stat(string(o))
	if err != nil {
		return time.Time{}
	}

	return s.ModTime()
}

// Open opens the file, returning it as a File.
func (o OSFile) Open() (File, error) {
	return os.Open(string(o))
}

/*
// Compressed adds the given node to the Directory tree and gzip decompresses
// it into a FileBytes and also adds it to the tree.
//
// NB: The compressed version has .gz appended to its name
func Compressed(d Dir, name string, node Node, size int) error {
	f, err := node.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer g.Close()
	buf := make([]byte, size)

	_, err = io.ReadFull(g, buf)
	if err != nil {
		return err
	}

	d.Create(name+".gz", node)
	d.Create(name, FileBytes(buf, node.ModTime()))
	return nil
}
*/

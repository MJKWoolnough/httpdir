// Package httpdir provides an in-memory implementation of http.FileSystem.
package httpdir // import "vimagination.zapto.org/httpdir"

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// Convenient FileMode constants.
const (
	ModeDir  fs.FileMode = fs.ModeDir | 0o755
	ModeFile fs.FileMode = 0o644
)

// Default is the Dir used by the top-level functions.
var Default = New(time.Now())

// Mkdir is a convenience function for Default.Mkdir.
func Mkdir(name string, modTime time.Time, index bool) error {
	return Default.Mkdir(name, modTime, index)
}

// Create is a convenience function for Default.Create.
func Create(name string, n Node) error {
	return Default.Create(name, n)
}

// Remove is a convenience function for Default.Remove.
func Remove(name string) error {
	return Default.Remove(name)
}

// Dir is the start of a simple in-memory filesystem tree.
type Dir struct {
	d dir
}

// New creates a new, initialised, Dir.
func New(t time.Time) Dir {
	return Dir{
		dir{
			modTime:  t,
			contents: make(map[string]Node),
		},
	}
}

// Open returns the file, or directory, specified by the given name.
//
// This method is the implementation of http.FileSystem and isn't intended to
// be used by clients of this package.
func (d Dir) Open(name string) (http.File, error) {
	n, err := d.get(name)
	if err != nil {
		return nil, err
	}

	return n.Open()
}

func (d Dir) get(name string) (namedNode, error) {
	name = path.Clean(name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	n := namedNode{"", d.d}

	if len(name) > 0 {
		for _, part := range strings.Split(name, "/") {
			nd, ok := n.Node.(dir)
			if !ok {
				return namedNode{}, fs.ErrInvalid
			}

			dn, ok := nd.contents[part]
			if !ok {
				return namedNode{}, fs.ErrNotExist
			}

			n = namedNode{part, dn}
		}
	}

	return n, nil
}

// Mkdir creates the named directory, and any parent directories required.
//
// modTime is the modification time of the directory, used in caching
// mechanisms.
//
// index specifies whether or not the directory allows a directory listing.
// NB: if the directory contains an index.html file, then that will be
// displayed instead, regardless the value of index.
//
// All directories created will be given the specified modification time and
// index bool.
//
// Directories already existing will not be modified.
func (d Dir) Mkdir(name string, modTime time.Time, index bool) error {
	_, err := d.makePath(path.Clean(name), modTime, index)

	return err
}

func (d Dir) makePath(name string, modTime time.Time, index bool) (dir, error) {
	name = strings.TrimPrefix(name, "/")
	td := d.d

	for _, part := range strings.Split(name, "/") {
		if part == "" {
			continue
		}

		if n, ok := td.contents[part]; ok {
			switch f := n.(type) {
			case dir:
				td = f
			default:
				return dir{}, fs.ErrInvalid
			}
		} else {
			nd := dir{
				index:    index,
				contents: make(map[string]Node),
				modTime:  modTime,
			}
			td.contents[part] = nd
			td = nd
		}
	}

	return td, nil
}

// Create places a Node into the directory tree.
//
// Any non-existent directories will be created automatically, setting the
// modTime to that of the Node and the index to false.
//
// If you want to specify alternate modTime/index values for the directories,
// then you should create them first with Mkdir.
func (d Dir) Create(name string, n Node) error {
	dname, fname := path.Split(name)

	dn, err := d.makePath(dname, n.ModTime(), false)
	if err != nil {
		return err
	}

	if _, ok := dn.contents[fname]; ok {
		return fs.ErrExist
	}

	dn.contents[fname] = n

	return nil
}

// Remove will remove a node from the tree.
//
// It will remove files and any directories, whether they are empty or not.
//
// Caution: httpdir does no internal locking, so you should provide your own if
// you intend to call this method.
func (d Dir) Remove(name string) error {
	dname, fname := path.Split(name)

	nn, err := d.get(dname)
	if err != nil {
		return err
	}

	if nd, ok := nn.Node.(dir); ok {
		return nd.Remove(fname)
	}

	return fs.ErrInvalid
}

// Node represents a data file in the tree.
type Node interface {
	Size() int64
	Mode() fs.FileMode
	ModTime() time.Time
	Open() (File, error)
}

type namedNode struct {
	name string
	Node
}

func (n namedNode) Name() string {
	return n.name
}

func (n namedNode) IsDir() bool {
	return n.Mode().IsDir()
}

func (n namedNode) Sys() interface{} {
	return n.Node
}

func (n namedNode) Open() (http.File, error) {
	f, err := n.Node.Open()
	if err != nil {
		return nil, err
	}

	return wrapped{n, f}, nil
}

// File represents an opened data Node.
type File interface {
	io.Reader
	io.Seeker
	io.Closer
	Readdir(int) ([]fs.FileInfo, error)
}

type wrapped struct {
	fs.FileInfo
	File
}

func (w wrapped) Stat() (fs.FileInfo, error) {
	return w.FileInfo, nil
}

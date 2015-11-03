package httpdir

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	ModeDir  os.FileMode = os.ModeDir | 0755
	ModeFile os.FileMode = 0644
)

type Dir struct {
	dir
}

func NewDir(t time.Time) Dir {
	return Dir{
		dir: dir{
			modTime:  t,
			contents: make(map[string]Node),
		},
	}
}

func (d *Dir) Mkdir(name string, modTime time.Time, index bool) error {
	_, err := d.makePath(path.Clean(name), modTime, index)
	return err
}

func (d *Dir) makePath(name string, modTime time.Time, index bool) (dir, error) {
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if len(name) == 0 {
		return dir{}, nil
	}
	td := d.dir
	for _, part := range strings.Split(name, "/") {
		n, ok := td.contents[part]
		if ok {
			switch f := n.(type) {
			case dir:
				td = f
			default:
				return dir{}, os.ErrInvalid
			}
		} else {
			nd := dir{
				index,
				make(map[string]Node),
				modTime,
			}
			td.contents[part] = nd
			td = nd
		}
	}
	return dir{}, nil
}

func (d *Dir) Create(name string, n Node) error {
	dname, fname := path.Split(name)
	dn, err := d.makePath(dname, n.ModTime(), false)
	if err != nil {
		return nil
	}
	if _, ok := dn.contents[fname]; ok {
		return os.ErrExist
	}
	dn.contents[fname] = n
	return nil
}

type dir struct {
	index    bool
	contents map[string]Node
	modTime  time.Time
}

func (d dir) Size() int64 {
	return 0
}

func (dir) Mode() os.FileMode {
	return ModeDir
}

func (d dir) ModTime() time.Time {
	return d.modTime
}

func (d dir) Open() (File, error) {
	if !d.index {
		return nil, os.ErrPermission
	}
	return newDirectory(d), nil
}

type Node interface {
	Size() int64
	Mode() os.FileMode
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

type File interface {
	io.Reader
	io.Seeker
	io.Closer
	Readdir(int) ([]os.FileInfo, error)
}

type wrapped struct {
	os.FileInfo
	File
}

func (w wrapped) Stat() (os.FileInfo, error) {
	return w.FileInfo, nil
}

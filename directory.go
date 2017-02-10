package httpdir

import (
	"io"
	"os"
	"sort"
	"time"
)

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
		if f, ok := d.contents["index.html"]; ok {
			return f.Open()
		}
		return nil, os.ErrPermission
	}
	contents := make([]os.FileInfo, 0, len(d.contents))
	for name, node := range d.contents {
		contents = append(contents, namedNode{name, node})
	}
	dir := &directory{
		contents: contents,
	}
	sort.Sort(dir)
	return dir, nil
}

type directory struct {
	contents []os.FileInfo
	pos      int
}

func (directory) Close() error {
	return nil
}

func (directory) Read([]byte) (int, error) {
	return 0, os.ErrInvalid
}

func (d *directory) Seek(offset int64, whence int) (int64, error) {
	pos := int64(d.pos)
	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos += offset
	case io.SeekEnd:
		pos = int64(len(d.contents)) - offset
	}
	if pos != 0 {
		return 0, os.ErrInvalid
	}
	d.pos = 0
	return 0, nil
}

func (d *directory) Readdir(n int) ([]os.FileInfo, error) {
	if n < 0 || d.pos+n > len(d.contents) {
		n = len(d.contents) - d.pos
	}
	last := d.pos + n
	toRet := d.contents[d.pos:last]
	if len(toRet) == 0 {
		return nil, io.EOF
	}
	d.pos = last
	return toRet, nil
}

func (d *directory) Len() int {
	return len(d.contents)
}

func (d *directory) Less(i, j int) bool {
	return d.contents[i].Name() < d.contents[j].Name()
}

func (d *directory) Swap(i, j int) {
	d.contents[i], d.contents[j] = d.contents[j], d.contents[i]
}

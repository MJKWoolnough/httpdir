package httpdir

import (
	"io"
	"os"
	"sort"
)

type directory struct {
	contents []os.FileInfo
	pos      int
}

func newDirectory(d dir) File {
	contents := make([]os.FileInfo, 0, len(d.contents))
	for name, node := range d.contents {
		contents = append(contents, namedNode{name, node})
	}
	dir := &directory{
		contents: contents,
	}
	sort.Sort(dir)
	return dir
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
	case os.SEEK_SET:
		pos = offset
	case os.SEEK_CUR:
		pos += offset
	case os.SEEK_END:
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

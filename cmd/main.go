package main

import (
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

var (
	pkg     = flag.String("p", "main", "package name")
	in      = flag.String("i", "", "input filename")
	out     = flag.String("o", "", "output filename")
	varname = flag.String("v", "httpdir", "http dir variable name")
	help    = flag.Bool("h", false, "show help")
	comp    = flag.Bool("c", false, "compress using gzip")
)

func errHandler(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type replacer struct {
	f *os.File
}

var hexArr = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

func (r replacer) Write(p []byte) (int, error) {
	var (
		toWrite [5]byte
		t       []byte
		n       int
	)
	toWrite[0] = '\\'
	for len(p) > 0 {
		dr, s := utf8.DecodeRune(p)
		if dr != utf8.RuneError && (s > 1 || (p[0] >= 0x20 && p[0] < 0x7f)) {
			t = append(toWrite[1:1], p[:s]...)
			if p[0] == '"' || p[0] == '\\' {
				t = toWrite[:2]
			}
		} else {
			t = toWrite[:2]
			switch p[0] {
			case '\a':
				toWrite[1] = 'a'
			case '\b':
				toWrite[1] = 'b'
			case '\f':
				toWrite[1] = 'f'
			case '\n':
				toWrite[1] = 'n'
			case '\t':
				toWrite[1] = '	'
				t = toWrite[1:2]
			case '\r':
				toWrite[1] = 'r'
			case '\v':
				toWrite[1] = 'v'
			default:
				toWrite[1] = 'x'
				toWrite[2] = hexArr[p[0]>>4]
				toWrite[3] = hexArr[p[0]&15]
				t = toWrite[:4]
			}
		}
		_, err := r.f.Write(t)
		n += s
		if err != nil {
			return n, err
		}
		p = p[s:]
	}
	return n, nil
}

type tickReplacer struct {
	f *os.File
}

func (r tickReplacer) Write(p []byte) (int, error) {
	var (
		m, n int
		err  error
	)
	hexes := make([]byte, 0, 32)
	toWrite := make([]byte, 0, 1024)
	for len(p) > 0 {
		dr, s := utf8.DecodeRune(p)
		if dr == utf8.RuneError || dr == '`' || dr == '\r' {
			hexes = append(hexes, p[0])
		} else {
			if len(hexes) > 0 {
				toWrite = toWrite[:0]
				toWrite = append(toWrite, '`', '+', '"')
				for _, b := range hexes {
					if b == '`' {
						toWrite = append(toWrite, '`')
					} else if b == '\r' {
						toWrite = append(toWrite, '\\', 'r')
					} else {
						toWrite = append(toWrite, '\\', 'x', hexArr[b>>4], hexArr[b&15])
					}
				}
				toWrite = append(toWrite, '"', '+', '`')
				_, err = r.f.Write(toWrite)
				n += len(hexes)
				if err != nil {
					break
				}
				hexes = hexes[:0]
			}
			m, err = r.f.Write(p[:s])
			n += m
			if err != nil {
				break
			}
		}
		p = p[s:]
	}
	return n, err
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	if *in == "" || *out == "" {
		errHandler(errors.New("missing in/out file"))
	}
	fi, err := os.Open(*in)
	errHandler(err)
	defer fi.Close()
	stat, err := fi.Stat()
	errHandler(err)
	fo, err := os.Create(*out)
	errHandler(err)
	defer fo.Close()
	if *comp {
		_, err = fmt.Fprintf(fo, compressedStart, *pkg)
		errHandler(err)
		w, err := gzip.NewWriterLevel(replacer{fo}, gzip.BestCompression)
		errHandler(err)
		_, err = io.Copy(w, fi)
		errHandler(err)
		errHandler(w.Close())
		_, err = fmt.Fprintf(fo, compressedEnd, stat.ModTime().Unix(), *varname, *in+".gz", *varname, *in, stat.Size())
		errHandler(err)
	} else {
		_, err = fmt.Fprintf(fo, uncompressedStart, *pkg, *varname, *in)
		errHandler(err)
		_, err = io.Copy(tickReplacer{fo}, fi)
		errHandler(err)
		_, err = fmt.Fprintf(fo, uncompressedEnd, stat.ModTime().Unix())
		errHandler(err)
	}
}

const (
	uncompressedStart = `package %s

import (
	"time"

	"github.com/MJKWoolnough/httpdir"
)

func init() {
	%s.Create(%q, httpdir.FileString(` + "`"
	uncompressedEnd = "`" + `, time.Unix(%d, 0)))
}
`

	compressedStart = `package %s

import (
	"time"

	"github.com/MJKWoolnough/httpdir"
)

func init() {
	n := httpdir.FileString("`
	compressedEnd = `", time.Unix(%d, 0))
	%s.Create(%q, n)
	%s.Create(%q, httpdir.Decompressed(n, %d))
}
`
)

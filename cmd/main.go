package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	pkg     = flag.String("p", "main", "package name")
	in      = flag.String("i", "", "input filename")
	out     = flag.String("o", "", "output filename")
	varname = flag.String("v", "dir", "http dir variable name")
	help    = flag.Bool("h", false, "show help")
	//comp = flag.String("-c", "", "compression method [gzip, bzip2, lzw, zlib, flate]")
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

func (r replacer) Write(p []byte) (int, error) {
	n := 0
	toWrite := make([]byte, 0, 32)
	toWrite = append(toWrite, "`+\""...)
	for len(p) > 0 {
		pos := bytes.IndexAny(p, "`\r")
		if pos == -1 {
			m, err := r.f.Write(p)
			return n + m, err
		}
		if pos > 0 {
			m, err := r.f.Write(p[:pos])
			n += m
			if err != nil {
				return n, err
			}
			p = p[pos:]
		}
		toWrite = toWrite[:3]
		for len(p) > 0 && (p[0] == '`' || p[0] == '\r') {
			if p[0] == '\r' {
				toWrite = append(toWrite, "\r"...)
			} else {
				toWrite = append(toWrite, p[0])
			}
			p = p[1:]
			n++
		}
		toWrite = append(toWrite, "\"+`"...)
		_, err := r.f.Write(toWrite)
		if err != nil {
			return n, err
		}
	}
	return n, nil
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
	fo, err := os.Create(*out)
	errHandler(err)
	_, err = fmt.Fprintf(fo, tStart, *pkg, *varname, *in)
	errHandler(err)
	_, err = io.Copy(replacer{fo}, fi)
	errHandler(err)
	stat, err := fi.Stat()
	errHandler(err)
	t := stat.ModTime().Unix()
	fi.Close()
	_, err = fmt.Fprintf(fo, tEnd, t)
	errHandler(err)
	fo.Close()
}

const (
	test   = "````````"
	tStart = `package %s

import (
	"time"

	"github.com/MJKWoolnough/httpdir"
)

func init() {
	%s.Create(%q, httpdir.FileString(` + "`"
	tEnd = "`" + `, time.Unix(%d, 0)))
}
`
)

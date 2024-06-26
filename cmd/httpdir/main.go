package main // import "vimagination.zapto.org/httpdir/cmd/httpdir"

import (
	"compress/flate"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/foobaz/go-zopfli/zopfli"
	"github.com/google/brotli/go/cbrotli"
	"vimagination.zapto.org/memio"
)

var (
	pkg      = flag.String("p", "main", "package name")
	in       = flag.String("i", "", "input filename")
	out      = flag.String("o", "", "output filename")
	odate    = flag.String("d", "", "modified date [seconds since epoch]")
	path     = flag.String("w", "", "http path")
	varname  = flag.String("v", "httpdir.Default", "http dir variable name")
	cvarname = flag.String("c", "httpdir.Default", "http dir compressed variable name")
	help     = flag.Bool("h", false, "show help")
	gzcomp   = flag.Bool("g", false, "compress using gzip")
	brcomp   = flag.Bool("b", false, "compress using brotli")
	flcomp   = flag.Bool("f", false, "compress using flate/deflate")
	single   = flag.Bool("s", false, "use single source var and decompress/compress for others")
	zpfcomp  = flag.Bool("z", false, "replace gzip with zopfli compression")
)

type replacer struct {
	f *os.File
}

var hexArr = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

func (r replacer) Write(p []byte) (int, error) {
	n := 0
	toWrite := make([]byte, 0, 5)

	for len(p) > 0 {
		rn, s := utf8.DecodeRune(p)
		if rn == utf8.RuneError {
			s = 1
		}

		if s > 1 || (p[0] > 0 && p[0] < 0x7f && p[0] != '\n' && p[0] != '\\' && p[0] != '"') {
			toWrite = append(toWrite[:0], p[:s]...)
		} else {
			switch p[0] {
			case '\n':
				toWrite = append(toWrite[:0], '\\', 'n')
			case '\\':
				toWrite = append(toWrite[:0], '\\', '\\')
			case '"':
				toWrite = append(toWrite[:0], '\\', '"')
			default:
				toWrite = append(toWrite[:0], '\\', 'x', hexArr[p[0]>>4], hexArr[p[0]&15])
			}
		}

		n += s

		if _, err := r.f.Write(toWrite); err != nil {
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
				n += len(hexes)

				if _, err = r.f.Write(toWrite); err != nil {
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

func e(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func ne(_ int, err error) {
	e(err)
}

type imports []string

func (im imports) Len() int {
	return len(im)
}

func (im imports) Less(i, j int) bool {
	si := strings.HasPrefix(im[i], "\"github.com")
	sj := strings.HasPrefix(im[j], "\"github.com")
	vi := strings.HasPrefix(im[i], "\"vimagination.zapto.org")
	vj := strings.HasPrefix(im[j], "\"vimagination.zapto.org")

	if si == sj && vi == vj {
		return im[i] < im[j]
	}

	return !si && !vi || si && vj
}

func (im imports) Swap(i, j int) {
	im[i], im[j] = im[j], im[i]
}

type encoding struct {
	Buffer                    []byte
	Compress, Decompress, Ext string
}

type encodings []encoding

func (e encodings) Len() int {
	return len(e)
}

func (e encodings) Less(i, j int) bool {
	return len(e[i].Buffer) < len(e[j].Buffer)
}

func (e encodings) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	var (
		f    *os.File
		err  error
		date int64
	)

	if *in == "-" || *in == "" {
		f = os.Stdin
		date = time.Now().Unix()
	} else {
		f, err = os.Open(*in)
		e(err)
		fi, err := f.Stat()
		e(err)
		date = fi.ModTime().Unix()
	}
	if *odate != "" {
		date, err = strconv.ParseInt(*odate, 10, 64)
		e(err)
	}

	data := make(memio.Buffer, 0, 1<<20)

	_, err = io.Copy(&data, f)
	e(err)
	e(f.Close())

	im := imports{"\"vimagination.zapto.org/httpdir\"", "\"time\""}

	encs := make(encodings, 1, 4)
	encs[0] = encoding{
		Buffer:     data,
		Compress:   identCompress,
		Decompress: identDecompress,
		Ext:        "",
	}

	if *brcomp {
		var b memio.Buffer

		br := cbrotli.NewWriter(&b, cbrotli.WriterOptions{Quality: 11})
		br.Write(data)
		br.Close()

		if *single {
			im = append(im, brotliImport)
		}

		encs = append(encs, encoding{
			Buffer:     b,
			Compress:   brotliCompress,
			Decompress: brotliDecompress,
			Ext:        ".br",
		})
	}
	if *flcomp {
		var b memio.Buffer

		fl, _ := flate.NewWriter(&b, flate.BestCompression)
		fl.Write(data)
		fl.Close()

		if *single {
			im = append(im, strings.Split(flateImport, "\n")...)
		}

		encs = append(encs, encoding{
			Buffer:     b,
			Compress:   flateCompress,
			Decompress: flateDecompress,
			Ext:        ".fl",
		})
	}
	if *gzcomp || *zpfcomp {
		var b memio.Buffer

		if *zpfcomp {
			zopfli.GzipCompress(&zopfli.Options{
				NumIterations:  100,
				BlockSplitting: true,
				BlockType:      2,
			}, data, &b)
		} else {
			gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
			gz.Write(data)
			gz.Close()
		}

		if *single {
			im = append(im, gzipImport)
		}

		encs = append(encs, encoding{
			Buffer:     b,
			Compress:   gzipCompress,
			Decompress: gzipDecompress,
			Ext:        ".gz",
		})
	}

	sort.Sort(encs)

	if *single && (*gzcomp || *brcomp || *flcomp) {
		im = append(im, "\"vimagination.zapto.org/memio\"")
		if encs[0].Ext != "" {
			im = append(im, "\"strings\"")
		}
	}

	if encs[0].Ext == ".fl" {
		im = append(im, "\"io\"")
	}

	sort.Sort(im)

	var (
		imports string
		ext     bool
	)

	for _, i := range im {
		if !ext && (strings.HasPrefix(i, "\"github.com") || strings.HasPrefix(i, "\"vimagination")) {
			imports += "\n"
			ext = true
		}
		imports += "	" + i + "\n"
	}

	if *out == "-" || *out == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(*out)
		e(err)
	}

	fmt.Fprintf(f, packageStart, *pkg, imports, date)

	if *single {
		f.WriteString(stringStart)

		if encs[0].Ext == "" {
			ne(f.WriteString("`"))
			ne(tickReplacer{f}.Write(encs[0].Buffer))
			ne(f.WriteString("`"))
		} else {
			ne(f.WriteString("\""))
			ne(replacer{f}.Write(encs[0].Buffer))
			ne(f.WriteString("\""))
		}

		f.WriteString(stringEnd)

		for n, enc := range encs {
			var (
				templ string
				vars  = []interface{}{0, *cvarname, *path + enc.Ext}
			)

			if enc.Ext == "" {
				vars = vars[1:]
				vars[0] = *varname

				if n == 0 {
					templ = identDecompress
				} else {
					templ = identCompress
				}
			} else {
				if n == 0 {
					vars[0] = len(data)
					templ = enc.Decompress
				} else {
					vars[0] = len(enc.Buffer)
					templ = enc.Compress
				}
			}

			fmt.Fprintf(f, templ, vars...)
		}
	} else {
		for _, enc := range encs {
			filename := *path + enc.Ext
			ne(fmt.Fprintf(f, soloStart, *varname, filename))

			if enc.Ext == "" {
				ne(f.WriteString("`"))
				ne(tickReplacer{f}.Write(enc.Buffer))
				ne(f.WriteString("`"))
			} else {
				ne(f.WriteString("\""))
				ne(replacer{f}.Write(enc.Buffer))
				ne(f.WriteString("\""))
			}

			ne(f.WriteString(soloEnd))
		}
	}

	ne(fmt.Fprint(f, packageEnd))
	e(f.Close())
}

const (
	packageStart = `package %s

import (
%s)

func init() {
	date := time.Unix(%d, 0)
`
	stringStart = `	s := `
	stringEnd   = `
`
	packageEnd = `}
`
	soloStart = `	%s.Create(%q, httpdir.FileString(`
	soloEnd   = `, date))
`
	identDecompress = `	b := []byte(s)
	%s.Create(%q, httpdir.FileString(s, date))
`
	identCompress = `	%s.Create(%q, httpdir.FileBytes(b, date))
`

	brotliImport     = "\"github.com/google/brotli/go/cbrotli\""
	brotliDecompress = `	b := make([]byte, %d)
	br := cbrotli.NewReader(strings.NewReader(s))
	br.Read(b)
	br.Close()
	%s.Create(%q, httpdir.FileString(s, date))
`
	brotliCompress = `	brb := make(memio.Buffer, 0, %d)
	br := cbrotli.NewWriter(&brb, cbrotli.WriterOptions{Quality: 11})
	br.Write(b)
	br.Close()
	%s.Create(%q, httpdir.FileBytes(brb, date))
`
	flateImport     = "\"compress/flate\""
	flateDecompress = `	b := make([]byte, %d)
	fl := flate.NewReader(strings.NewReader(s))
	io.ReadFull(fl, b)
	fl.Close()
	%s.Create(%q, httpdir.FileString(s, date))
`
	flateCompress = `	flb := make(memio.Buffer, 0, %d)
	fl, _ := flate.NewWriter(&flb, flate.BestCompression)
	fl.Write(b)
	fl.Close()
	%s.Create(%q, httpdir.FileBytes(flb, date))
`
	gzipImport     = "\"compress/gzip\""
	gzipDecompress = `	b := make([]byte, %d)
	gz, _ := gzip.NewReader(strings.NewReader(s))
	gz.Read(b)
	gz.Close()
	%s.Create(%q, httpdir.FileString(s, date))
`
	gzipCompress = `	gzb := make(memio.Buffer, 0, %d)
	gz, _ := gzip.NewWriterLevel(&gzb, gzip.BestCompression)
	gz.Write(b)
	gz.Close()
	%s.Create(%q, httpdir.FileBytes(gzb, date))
`
)

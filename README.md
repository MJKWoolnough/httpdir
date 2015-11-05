# httpdir
--
    import "github.com/MJKWoolnough/httpdir"

Package httpdir provides an in-memory implementation of http.FileSystem

## Usage

```go
const (
	ModeDir  os.FileMode = os.ModeDir | 0755
	ModeFile os.FileMode = 0644
)
```
Convenient FileMode constants

```go
var Default = New(time.Now())
```
Default is the Dir used by the top-level functions

#### func  Create

```go
func Create(name string, n Node) error
```
Create is a convenience function for Default.Create

#### func  Mkdir

```go
func Mkdir(name string, modTime time.Time, index bool) error
```
Mkdir is a convenience function for Default.Mkdir

#### type Dir

```go
type Dir struct {
}
```

Dir is the start of a simple in-memory filesystem tree

#### func  New

```go
func New(t time.Time) Dir
```
New creates a new, initialised, Dir

#### func (Dir) Create

```go
func (d Dir) Create(name string, n Node) error
```
Create places a Node into the directory tree.

Any non-existant directories will be created automatically, setting the modTime
to that of the Node and the index to false.

If you want to specify alternate modTime/index values for the directories, then
you should create them first with Mkdir

#### func (Dir) Mkdir

```go
func (d Dir) Mkdir(name string, modTime time.Time, index bool) error
```
Mkdir creates the named directory, and any parent directories required.

modTime is the modification time of the directory, used in caching mechanisms.

index specifies whether or not the directory allows a directory listing. NB: if
the directory contains an index.html file, then that will be displayed instead,
regardless the value of index.

All directories created will be given the specified modification time and index
bool.

Directories already existing will not be modified.

#### func (Dir) Open

```go
func (d Dir) Open(name string) (http.File, error)
```
Open returns the file, or directory, specified by the given name.

This method is the implementation of http.FileSystem and isn't intended to be
used by clients of this package.

#### type File

```go
type File interface {
	io.Reader
	io.Seeker
	io.Closer
	Readdir(int) ([]os.FileInfo, error)
}
```

File represents an opened data Node

#### type Node

```go
type Node interface {
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	Open() (File, error)
}
```

Node represents a data file in the tree

#### func  FileBytes

```go
func FileBytes(data []byte, modTime time.Time) Node
```
FileBytes provides an implementation of Node that takes a byte slice as its data
source

#### func  FileString

```go
func FileString(data string, modTime time.Time) Node
```
FileString provides an implementation of Node that takes a string as its data
source

#### type OSFile

```go
type OSFile string
```

OSFile is the path of a file in the real filesystem to be put into the im-memory
filesystem

#### func (OSFile) ModTime

```go
func (o OSFile) ModTime() time.Time
```
ModTime returns the ModTime of the file

#### func (OSFile) Mode

```go
func (o OSFile) Mode() os.FileMode
```
Mode returns the Mode of the file

#### func (OSFile) Open

```go
func (o OSFile) Open() (File, error)
```
Open opens the file, returning it as a File

#### func (OSFile) Size

```go
func (o OSFile) Size() int64
```
Size returns the size of the file

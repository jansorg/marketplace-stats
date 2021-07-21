// Code generated by "esc -o static.go -pkg report static"; DO NOT EDIT.

package report

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is not directory", f.name)
	}

	fis, ok := _escDirs[f.local]
	if !ok {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is directory, but we have no info about content of this dir, local=%s", f.name, f.local)
	}
	limit := count
	if count <= 0 || limit > len(fis) {
		limit = len(fis)
	}

	if len(fis) == 0 && count > 0 {
		return nil, io.EOF
	}

	return fis[0:limit], nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/static/report.gohtml": {
		name:    "report.gohtml",
		local:   "static/report.gohtml",
		size:    33615,
		modtime: 1626861424,
		compressed: `
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
QKzDxf8FAAD//4Vd829PgwAA
`,
	},

	"/static": {
		name:  "static",
		local: `static`,
		isDir: true,
	},
}

var _escDirs = map[string][]os.FileInfo{

	"static": {
		_escData["/static/report.gohtml"],
	},
}

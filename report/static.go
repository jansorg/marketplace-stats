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
		size:    32271,
		modtime: 1614800926,
		compressed: `
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
taqQGmUYiFU4+98AAAD//4pfWaYPfgAA
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

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
		size:    32412,
		modtime: 1618301985,
		compressed: `
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
neLZyVCrCqlRhoFYhrP/DQAA//8kmyZhnH4AAA==
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

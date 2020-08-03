package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// fileWriter writes to a temp file and later atomically renames it.
type fileWriter struct {
	p    string   // target filename
	f    *os.File // temp file
	werr error    // first error encountered while writing
}

func newFileWriter(p string) (*fileWriter, error) {
	f, err := ioutil.TempFile(filepath.Dir(p), filepath.Base(p)+".*")
	if err != nil {
		return nil, err
	}
	return &fileWriter{p, f, nil}, nil
}

func (fw *fileWriter) printf(format string, args ...interface{}) {
	if fw.werr == nil {
		_, fw.werr = fmt.Fprintf(fw.f, format, args...)
	}
}

func (fw *fileWriter) close() error {
	defer os.Remove(fw.f.Name()) // no-op on success
	cerr := fw.f.Close()
	if fw.werr != nil {
		return fw.werr
	}
	if cerr != nil {
		return cerr
	}
	return os.Rename(fw.f.Name(), fw.p)
}

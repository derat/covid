// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

// Package filewriter safely writes files.
package filewriter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileWriter writes to a temp file and later atomically renames it.
// If a write error occurs, it is saved internally and future writes become no-ops.
type FileWriter struct {
	p    string   // target filename
	f    *os.File // temp file
	werr error    // first error encountered while writing
}

// New returns a new FileWriter that will write to the supplied path.
func New(p string) (*FileWriter, error) {
	f, err := ioutil.TempFile(filepath.Dir(p), filepath.Base(p)+".*")
	if err != nil {
		return nil, err
	}
	return &FileWriter{p, f, nil}, nil
}

// Printf writes the supplied formatted data and returns the number of bytes written.
func (fw *FileWriter) Printf(format string, args ...interface{}) int {
	var n int
	if fw.werr == nil {
		n, fw.werr = fmt.Fprintf(fw.f, format, args...)
	}
	return n
}

// Close renames the temp file to the path originally supplied to New.
// If a write error occurred earlier, it is returned and no other action is taken.
func (fw *FileWriter) Close() error {
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

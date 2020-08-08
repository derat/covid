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
// Callers can ignore write errors and just check the return value from Close.
type FileWriter struct {
	p   string   // target filename
	f   *os.File // temp file; nil if creation failed
	err error    // first error encountered
}

// New returns a new FileWriter to write to the supplied path.
func New(p string) *FileWriter {
	f, err := ioutil.TempFile(filepath.Dir(p), filepath.Base(p)+".*")
	if err != nil {
		return &FileWriter{"", nil, err}
	}
	return &FileWriter{p, f, nil}
}

// Write writes the supplied bytes to the file as in io.Writer.
// If an error occurred earlier, nothing is written and the earlier error is returned.
func (fw *FileWriter) Write(p []byte) (int, error) {
	var n int
	if fw.err == nil {
		n, fw.err = fw.f.Write(p)
	}
	return n, fw.err
}

// Printf writes the supplied formatted data as in fmt.Fprintf.
// If an error occurred earlier, nothing is written and the earlier error is returned.
func (fw *FileWriter) Printf(format string, args ...interface{}) (int, error) {
	var n int
	if fw.err == nil {
		n, fw.err = fmt.Fprintf(fw.f, format, args...)
	}
	return n, fw.err
}

// Close renames the temp file to the path originally supplied to New.
// If a write error occurred earlier, it is returned and no other action is taken.
func (fw *FileWriter) Close() error {
	if fw.f == nil { // failed to create temp file
		return fw.err
	}
	defer os.Remove(fw.f.Name()) // no-op if we successfully rename temp file
	cerr := fw.f.Close()
	if fw.err != nil {
		return fw.err
	}
	if cerr != nil {
		return cerr
	}
	return os.Rename(fw.f.Name(), fw.p)
}

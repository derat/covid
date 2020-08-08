// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

// Package gnuplot makes it slightly easier to generate plots using gnuplot.
package gnuplot

import (
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

// ExecTemplate executes the supplied Go template and data to write a .gnuplot file,
// which it then passes to gnuplot.
func ExecTemplate(tmpl string, data interface{}) error {
	// Execute the template to write the .gnuplot file.
	gf, err := ioutil.TempFile("", "gnuplot.")
	if err != nil {
		return err
	}
	defer os.Remove(gf.Name())

	terr := template.Must(template.New("").Parse(tmpl)).Execute(gf, data)
	cerr := gf.Close()
	if terr != nil {
		return terr
	}
	if cerr != nil {
		return cerr
	}
	return exec.Command("gnuplot", gf.Name()).Run()
}

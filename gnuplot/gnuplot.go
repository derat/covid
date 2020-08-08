// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

// Package gnuplot makes it slightly easier to generate plots using gnuplot.
package gnuplot

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

	if err := exec.Command("gnuplot", gf.Name()).Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) != 0 {
			return fmt.Errorf("%v: %q", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return err
	}
	return nil
}

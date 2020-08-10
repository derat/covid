// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"fmt"
	"time"
)

func templateData(dataPath, imgPath string, now time.Time, vars map[string]interface{}) interface{} {
	return struct {
		DataPath    string // path to gnuplot data file
		SetTerm     string // 'set term' command for writing PNG image data
		SetOutput   string // 'set output' command for writing to image file
		LineCfg     string // 'lines' default args
		FooterLabel string // 'set label' command for writing footer label

		Vars map[string]interface{} // extra variables
	}{
		DataPath:  dataPath,
		SetTerm:   "set term pngcairo font 'Roboto,22' size 1280,960",
		SetOutput: fmt.Sprintf("set output '%s'", imgPath),
		LineCfg:   "lc rgb '#0D47A1' lw 3",
		FooterLabel: fmt.Sprintf(
			"set label front '{/*0.7 Generated on %s by https://github.com/derat/covid}' at screen 0.99,0.025 right",
			time.Now().Format("2006-01-02")),
		Vars: vars,
	}
}

const (
	posAgeTmpl = `
set title 'Puerto Rico Bioportal positive COVID-19 tests by age'

# Plot data initially to set GPVAL_DATA_* variables:
# http://www.phyast.pitt.edu/~zov1/gnuplot/html/statistics.html
set term unknown
plot '{{.DataPath}}' using 1:3

{{.SetTerm}}
{{.SetOutput}}

set view map
set xtics scale 0 rotate by 90 right
set xlabel 'Reporting week' offset 0,-1.5
set xrange [GPVAL_DATA_X_MIN-0.5:GPVAL_DATA_X_MAX+0.5]
set yrange [GPVAL_DATA_Y_MIN-5:GPVAL_DATA_Y_MAX+5]
set ylabel 'Age'
set ytics scale 0 offset 0,-0.75
set bmargin 5
{{.FooterLabel}}

splot '{{.DataPath}}' using 1:3:4:xtic(2) with image notitle
`

	reportsTmpl = `
set title 'Puerto Rico Bioportal reported COVID-19 tests'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting day'
set ylabel 'Reported results (7-day average)'
set yrange [0:*]
set key off
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:2 with lines {{.LineCfg}} notitle
`

	delaysTmpl = `
set title 'Puerto Rico Bioportal COVID-19 {{.Vars.TestType}} test result delays'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting week'
set ylabel 'Result delay (days)'
set yrange [0:{{.Vars.MaxDelay}}]
set key top left
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:2:6 with filledcurves lc rgb '#DDDDDD' title '10th-90th', \
	 '{{.DataPath}}' using 1:3:5 with filledcurves lc rgb '#BBBBBB' title '25th-75th', \
     '{{.DataPath}}' using 1:4 with lines lc black lw 3 title 'Median'
`
)

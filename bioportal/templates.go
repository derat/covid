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
		FooterLabel string // 'set label' command for writing footer label

		Vars map[string]interface{} // extra variables
	}{
		DataPath:  dataPath,
		SetTerm:   "set term pngcairo font 'Roboto,22' size 1280,960 linewidth 2",
		SetOutput: fmt.Sprintf("set output '%s'", imgPath),
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
set title 'Puerto Rico Bioportal COVID-19 molecular tests'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting date'
set ylabel 'Reported results (7-day average)'
set yrange [0:*]
set grid front xtics ytics
set key off
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:2 with lines lc black lw 2 notitle
`

	typesTmpl = `
set title 'Puerto Rico Bioportal COVID-19 tests by type'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting date'
set ylabel 'Reported results'
set yrange [0:*]
set grid front xtics ytics
set key top left invert
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:5 with lines lc rgb '#ef9a9a' lw 2 title 'Unknown', \
     '{{.DataPath}}' using 1:3 with lines lc rgb '#cccccc' lw 2 title 'Serological', \
     '{{.DataPath}}' using 1:4 with lines lc rgb '#009688' lw 2 title 'Antigen', \
     '{{.DataPath}}' using 1:2 with lines lc rgb '#3f51b5' lw 2 title 'Molecular'
`

	posRateTmpl = `
set title 'Puerto Rico Bioportal COVID-19 test positivity rate'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Sample collection date'
set ylabel 'Percent positive (7-day average)'
set yrange [0:*]
set grid front xtics ytics
set key off
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:2 with lines lc black lw 2 notitle
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
set grid front xtics ytics
set key top left
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:2:6 with filledcurves lc rgb '#dddddd' title '10th-90th', \
	 '{{.DataPath}}' using 1:3:5 with filledcurves lc rgb '#bbbbbb' title '25th-75th', \
     '{{.DataPath}}' using 1:4 with lines lc black lw 2 title 'Median'
`
)

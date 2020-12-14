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
	ageHeatTmpl = `
set title 'Puerto Rico Bioportal {{.Vars.Units}} by age'

# Plot data initially to set GPVAL_DATA_* variables:
# http://www.phyast.pitt.edu/~zov1/gnuplot/html/statistics.html
set term unknown
plot '{{.DataPath}}' using 1:3

{{.SetTerm}}
{{.SetOutput}}

set view map
set size ratio 0.4
set tics font ', 20'
set xtics scale 0 rotate by 90 right
set xlabel '{{if .Vars.Collect}}Sample collection{{else}}Reporting{{end}} week' offset 0,-1.5
set xrange [GPVAL_DATA_X_MIN-0.5:GPVAL_DATA_X_MAX+0.5]
set yrange [GPVAL_DATA_Y_MIN-5:GPVAL_DATA_Y_MAX+5]
set ylabel 'Age'
set ytics scale 0 offset 0,-0.5
set bmargin 5
set lmargin at screen 0.12
set rmargin at screen 0.85
{{.FooterLabel}}

splot '{{.DataPath}}' using 1:3:4:xtic(2) with image notitle
`

	typesTmpl = `
set title 'Puerto Rico Bioportal COVID-19 daily reported tests'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set format x '%m/%d'
set xlabel 'Reporting date'
set ylabel 'Reported results (7-day average)'
set yrange [0:*]
set grid front xtics ytics
set key top left invert
set bmargin 5
{{.FooterLabel}}

plot '{{.DataPath}}' using 1:5 with lines lc rgb '#ef9a9a' lw 2 title 'Unknown', \
     '{{.DataPath}}' using 1:3 with lines lc rgb '#dddddd' lw 2 title 'Serological', \
     '{{.DataPath}}' using 1:4 with lines lc rgb '#009688' lw 2 title 'Antigen', \
     '{{.DataPath}}' using 1:2 with lines lc rgb '#3f51b5' lw 2 title 'Molecular'
`

	posRateTmpl = `
set title 'Puerto Rico Bioportal COVID-19 test positivity rate'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set format x '%m/%d'
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
set format x '%m/%d'
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

	ageDistTmpl = `
set title 'Puerto Rico Bioportal COVID-19 positive test distribution by age'

{{.SetTerm}}
{{.SetOutput}}

set timefmt '%Y-%m-%d'
set xdata time
set format x '%m/%d'
set autoscale xfix
set xlabel 'Sample collection date'
set ylabel 'Fraction of all positives (7-day average)'
set yrange [0:*]
set grid front xtics ytics
set key outside autotitle columnheader
set bmargin 5
{{.FooterLabel}}

# Based on Anna Schneider's https://github.com/aschn/gnuplot-colorbrewer/blob/master/sequential/YlOrRd.plt
set linetype 1 lc rgb '#FFFFCC' # very light yellow-orange-red
set linetype 2 lc rgb '#FFEDA0' #
set linetype 3 lc rgb '#FED976' # light yellow-orange-red
set linetype 4 lc rgb '#FEB24C' #
set linetype 5 lc rgb '#FD8D3C' #
set linetype 6 lc rgb '#FC4E2A' # medium yellow-orange-red
set linetype 7 lc rgb '#E31A1C' #
set linetype 8 lc rgb '#B10026' # dark yellow-orange-red
set linetype 9 lc rgb '#80001c'
set linetype 10 lc rgb '#660016'

plot for [i=11:2:-1] '{{.DataPath}}' using 1:i with filledcurves x1 linestyle i-1
`
)

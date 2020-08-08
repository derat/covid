// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

const (
	posAgeTmpl = `
set title 'Puerto Rico weekly reported positive COVID-19 tests by age'

# Plot data initially to set GPVAL_DATA_* variables:
# http://www.phyast.pitt.edu/~zov1/gnuplot/html/statistics.html
set term unknown
plot '{{.DataPath}}' using 1:3

set term pngcairo
set output '{{.ImagePath}}'

set view map
set xtics scale 0 rotate by 90 right
set xrange [GPVAL_DATA_X_MIN-0.5:GPVAL_DATA_X_MAX+0.5]
set yrange [GPVAL_DATA_Y_MIN-5:GPVAL_DATA_Y_MAX+5]
set ylabel 'Age'
set ytics offset 0,-1

splot '{{.DataPath}}' using 1:3:4:xtic(2) with image notitle`

	delaysTmpl = `
set title 'Puerto Rico weekly COVID-19 test delays'

set term pngcairo
set output '{{.ImagePath}}'

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting week'
set ylabel 'Reporting delay (days)'
set yrange [0:*]
set key top left

plot '{{.DataPath}}' using 1:3 with lines lc 'blue' title 'Median', \
     '{{.DataPath}}' using 1:2:4 with filledcurves lc 'skyblue' fs transparent solid 0.25 title '1st-3rd Quartile'
`
)

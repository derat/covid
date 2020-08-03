set title 'Puerto Rico weekly positive reported tests by age'

# Plot data initially to set GPVAL_DATA_* variables:
# http://www.phyast.pitt.edu/~zov1/gnuplot/html/statistics.html
plot 'positives-age.data' u 1:3

set view map
set xtics scale 0 rotate by 90 right
set xrange [GPVAL_DATA_X_MIN-0.5:GPVAL_DATA_X_MAX+0.5]
set yrange [GPVAL_DATA_Y_MIN-5:GPVAL_DATA_Y_MAX+5]
set ylabel 'Age'
set ytics offset 0,-1

splot 'positives-age.data' using 1:3:4:xtic(2) with image notitle

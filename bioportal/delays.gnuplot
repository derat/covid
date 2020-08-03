set title 'Puerto Rico weekly COVID-19 test delays'

set timefmt '%Y-%m-%d'
set xdata time
set xlabel 'Reporting week'
set ylabel 'Reporting delay (days)'
set yrange [0:*]
set key top left

plot 'delays.data' using 1:3 smooth csplines with lines lc 'blue' title 'Median', \
     'delays.data' using 1:2:4 with filledcurves lc 'skyblue' fs transparent solid 0.25 title '20th-80th'

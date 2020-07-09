states="AZ CA FL NJ NV NY TX"

set title "Cumulative COVID-19 deaths in " . states
set timefmt '%Y%m%d'
set xdata time
set xlabel "Date"
set yrange [0:*]
set ylabel "Cumulative COVID-19 deaths"
set key top left

state(n) = word(states,n)

plot for [i=1:words(states)] \
  "<csvtool namedcol date,state,death daily.csv | " . \
  "awk -F, 'NR>1 && $2==\"".state(i)."\" {print $1,$3}'" \
  using 1:2 with lines title state(i)

states="AZ CA FL NV TX"

set title "COVID-19 hospitalizations in " . states
set timefmt '%Y%m%d'
set xdata time
set xlabel "Date"
set yrange [0:*]
set ylabel "Current COVID-19 hospitalizations"
set key top left

state(n) = word(states,n)

plot for [i=1:words(states)] \
  "<csvtool namedcol date,state,hospitalizedCurrently daily.csv | " . \
  "awk -F, 'NR>1 && $2==\"".state(i)."\" {print $1,$3}'" \
  using 1:2 with lines title state(i)

weeks = 8

set title "Per-state COVID-19 hospitalizations"

dates = system("for i in $(seq ".(weeks-1)." -1 0); do date +%Y%m%d -d \"today -$i week -1 day\"; done")
date(n) = word(dates,n)

# http://gnuplot.sourceforge.net/demo/boxplot.html
set style fill solid 0.5 border -1
set style boxplot outliers pointtype 7
set style data boxplot
set boxwidth  0.4
set pointsize 0.5

set key off
set border 2
set ytics nomirror
set grid ytics
set ylabel "Current COVID-19 hospitalizations"

# https://stackoverflow.com/a/37453347
set xtics () scale 0
set for [i=1:words(dates)] xtics add (date(i)[5:6]."/".date(i)[7:8] i)
set xlabel "Date"

plot for [i=1:words(dates)] \
  "<csvtool namedcol date,hospitalizedCurrently  daily.csv | " . \
  "awk -F, 'NR>1 && $1==\"".date(i)."\" && $2!=\"\" {print $2}'" \
  using (i):1 title date(i)

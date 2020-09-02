set xlabel 'Month'
set xdata time
set timefmt '%m-%d'
set format x '%b'

set ylabel 'Deaths'
set yrange [*<0:*]
set grid xtics ytics

filename='20200902.csv'
state='United States'

set key autotitle top right title 'Year'

plot for [year=2017:2020] \
  "<cat '".filename."' | ./year-over-year.sh '".state."' ".year \
  using 1:2 with lines title "".year

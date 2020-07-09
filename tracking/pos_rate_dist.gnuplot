weeks = 8
min_tests = 1000

set title "Distribution of per-state weekly positive COVID-19 testing rates\n".\
  "{/*0.8 Only includes days where state reported at least ".min_tests." new tests}"

dates = system("for i in $(seq ".(weeks-1)." -1 0); do date +%Y%m%d -d \"today -$i week -1 day\"; done")
date(n) = word(dates,n)

start_date = system("date +%Y%m%d -d \"".date(1)." -1 week\"")
pdate(n) = n>1 ? date(n-1) : start_date

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
set ylabel "Positivity Rate"

# https://stackoverflow.com/a/37453347
set xtics () scale 0
set for [i=1:words(dates)] xtics add (date(i)[5:6]."/".date(i)[7:8] i)
set xlabel "Week Ending"

plot for [i=1:words(dates)] \
  "<csvtool namedcol date,state,positiveIncrease,totalTestResultsIncrease daily.csv | " . \
  "awk -F, '" . \
  "NR>1 && $1>\"".pdate(i)."\" && $1<=\"".date(i)."\" && $3>0 && $4>=".min_tests." && $3!=$4 " . \
  "{ state=$2; pos[state] += $3; tot[state] += $4 } " . \
  "END { for (state in pos) { print pos[state]/tot[state] } }'" \
  using (i):1 title date(i)

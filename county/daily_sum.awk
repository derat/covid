# Accepts daily_diff.awk's output as input and produces CSV with daily sums:
#
# 1/23/20,1/24/20,...
# 0,1,...

BEGIN {
  FS = ","
  first = 3  # first date column in input (1-based)
}
{
  for (i = first; i <= NF; i++) {
    if (NR==1) {
      if (i > firstcol) printf ","
      printf "%s", $i
    } else {
      sums[i-first] += $i
    }
  }
  if (NR==1) printf "\n"
}
END {
  for (i = 0; i <  length(sums); i++) {
      if (i > 0) printf ","
      printf "%s", sums[i]
  }
  printf "\n"
}

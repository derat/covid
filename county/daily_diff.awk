# Accepts covid_confirmed_usafacts.csv or covid_deaths_usafacts.csv as input and
# produces CSV with per-county daily increases:
#
# County Name,State,1/23/20,1/24/20,...
# Statewide Unallocated,AL,0,0,...
# ...

BEGIN { FS="," }
{
  printf "%s,%s", $2, $3          # county and state
  for (i = 6; i <= NF; i++) {     # second day to final day
    if (NR==1) printf ",%s", $i   # preserve headers
    else printf ",%d", $i-$(i-1)  # subtract previous day
  }
  printf "\n"
}

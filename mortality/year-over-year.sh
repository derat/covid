#!/bin/sh

est='Unweighted'
#est='Predicted (weighted)'

sed '1s/^\xEF\xBB\xBF//' | \
  csvtool namedcol 'Week Ending Date,State,Observed Number,Year,Type,Outcome' - | \
  csvtool drop 1 - | \
  awk -F, "\$2==\"$1\" && \$4==\"$2\" && \$5==\"${est}\" && \$6==\"All causes\" {printf \"%s %s\n\",substr(\$1,6,5),\$3 }"

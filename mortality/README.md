# CDC Mortality Data Latency

This code visualizes the latency of the National Vital Statistics System
mortality data in the CDC's [Excess Deaths Associated with COVID-19] dataset.

It parses CSV snapshots downloaded at different times and graphs the increase in
each week's reported deaths over time.

[Excess Deaths Associated with COVID-19]: https://data.cdc.gov/NCHS/Excess-Deaths-Associated-with-COVID-19/xkkf-xrst/

## Data

The CSV data can be downloaded from
<https://data.cdc.gov/api/views/xkkf-xrst/rows.csv?accessType=DOWNLOAD&bom=true&format=true%20target=>,
and several historical snapshots are available at
<http://web.archive.org/web/*/https://data.cdc.gov/api/views/xkkf-xrst/rows.csv?accessType=DOWNLOAD&bom=true&format=true%20target=>.
The dataset is updated roughly weekly.

The CDC also provides provides [Technical Notes] with more information about the
data.

[Technical Notes]: https://www.cdc.gov/nchs/nvss/vsrr/covid19/tech_notes.htm

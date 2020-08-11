# Puerto Rico Bioportal COVID-19 test results

COVID-19 test results produced by private laboratories can be downloaded from
the (undocumented?) [Bioportal API].

As of 2020-08-08, this URL hangs for a few minutes before producing an 80 MB
JSON array of objects describing results (represented by the `test` struct in
[test.go](./test.go)).

As of 2020-08-10, the Bioportal API reports results from antigen, molecular, and
serological tests. The plots below only include molecular tests.

[BioPortal API]: https://bioportal.salud.gov.pr/api/administration/reports/minimal-info-unique-tests

## Positive tests by age

This heatmap displays weekly positive test results by patient age.

![positive tests by age](https://github.com/derat/covid-plots/raw/master/bioportal/positives-age.png)

## Result delays

Per [reporting by Primera Hora], positivity rates computed from the Bioportal's
data are inaccurate due to negative results often lagging behind positive
results.

![total test result delays](https://github.com/derat/covid-plots/raw/master/bioportal/result-delays.png)

![positive test result delays](https://github.com/derat/covid-plots/raw/master/bioportal/positive-result-delays.png)

![negative test result delays](https://github.com/derat/covid-plots/raw/master/bioportal/negative-result-delays.png)

[reporting by Primera Hora]: https://www.primerahora.com/noticias/gobierno-politica/notas/incierto-el-por-ciento-de-positividad-del-coronavirus-en-la-isla/

## Positivity rate

This plot attempts to work around the different latencies for positive and
negative test results by using the sample collection date (rather than reporting
date) on the X-axis and excluding the last 14 days of testing.

![test positivity rate](https://github.com/derat/covid-plots/raw/master/bioportal/result-delays.png)

## Tests reported per day

![tests reported per day](https://github.com/derat/covid-plots/raw/master/bioportal/reports-daily.png)

See also [Dr. Rafael Irrizary's dashboard], which presents data from the same
source.

[Dr. Rafael Irrizary's dashboard]: https://rconnect.dfci.harvard.edu/covidpr/

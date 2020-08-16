# Puerto Rico Bioportal COVID-19 test results

COVID-19 test results produced by private laboratories can be downloaded from
the (undocumented?) [Bioportal API].

This URL hangs for a few minutes before producing a 100+ MB JSON array of
objects describing results (represented by the `test` struct in
[test.go](./test.go)).

As of 2020-08-10, the Bioportal API reports results from antigen, molecular
(PCR), and serological (antibody) tests. The plots below include only molecular
tests unless indicated otherwise.

[BioPortal API]: https://bioportal.salud.gov.pr/api/administration/reports/minimal-info-unique-tests

## Results by age

These heatmaps display data based on weekly test results grouped by patient age.
Demographic data used to calculate per-100k numbers is from a 2017 UN dataset.

![positive tests by age](https://github.com/derat/covid-plots/raw/master/bioportal/positives-age.png)

![positive tests per 100k by age](https://github.com/derat/covid-plots/raw/master/bioportal/positives-age-scaled.png)

![total tests per 100k by age](https://github.com/derat/covid-plots/raw/master/bioportal/results-age-scaled.png)

![positivity rate by age](https://github.com/derat/covid-plots/raw/master/bioportal/positivity-age.png)

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

![test positivity rate](https://github.com/derat/covid-plots/raw/master/bioportal/positivity.png)

## Testing volume

![tests reported per day by type](https://github.com/derat/covid-plots/raw/master/bioportal/test-types.png)

---

See also [Dr. Rafael Irrizary's dashboard], which presents data from the same
source.

[Dr. Rafael Irrizary's dashboard]: https://rconnect.dfci.harvard.edu/covidpr/

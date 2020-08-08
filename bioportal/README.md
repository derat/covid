# Puerto Rico Bioportal COVID-19 test results

COVID-19 test results produced by private laboratories can be downloaded from
the (undocumented?) [Bioportal API]. As of 2020-08-08, this URL hangs for a few
minutes before producing an 80 MB JSON array of objects describing results
(represented by the `test` struct in [test.go](./test.go)).

Per [reporting by Primera Hora], positivity rates computed from the Bioportal's
data are inaccurate due to negative results often lagging behind positive
results.

![positive COVID-19 tests by age](https://github.com/derat/covid-plots/raw/master/bioportal/positives-age.png)

![COVID-19 test result delays](https://github.com/derat/covid-plots/raw/master/bioportal/result-delays.png)

See also [Dr. Rafael Irrizary's dashboard], which presents data from the same
source.

[BioPortal API]: https://bioportal.salud.gov.pr/api/administration/reports/minimal-info-unique-tests
[reporting by Primera Hora]: https://www.primerahora.com/noticias/gobierno-politica/notas/incierto-el-por-ciento-de-positividad-del-coronavirus-en-la-isla/
[Dr. Rafael Irrizary's dashboard]: https://rconnect.dfci.harvard.edu/covidpr/

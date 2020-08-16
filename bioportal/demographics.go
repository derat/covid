package main

// From http://data.un.org/Data.aspx?d=POP&f=tableCode%3a22, with 'Country or Area' set to
// 'Puerto Rico' and 'Year' set to '2017' (the latest available). Collected 20200816, with
// the page reporting 'Last update in UNdata: 2020/02/12'.
var unAgePop = map[ageRange]int{
	age0To9:   147970 + 177739, // 0-4 + 5-9
	age10To19: 198257 + 222678, // 10-14 + 15-19
	age20To29: 232150 + 223828, // ...
	age30To39: 190755 + 207678,
	age40To49: 208209 + 214945,
	age50To59: 224402 + 219484,
	age60To69: 210332 + 195563,
	age70To79: 171623 + 123063,
	age80To89: 84508 + 83993, // 80-84 + 85+
}

// https://data.census.gov/cedsci/table?q=puerto%20rico&tid=ACSDP1Y2018.DP05&hidePreview=false
// provides estimated population by age range from 2018, but some of the ranges span multiple decades
// (e.g. 25-34 or 35-44), making them not directly comparable to the ranges used by the Bioportal
// (20-29, 30-39, etc.).

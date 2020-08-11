// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// test represents the result of an individual test reported to the Bioportal.
// Each test is described as a JSON object:
//
//  {
//    "collectedDate": "6/25/2020",
//    "reportedDate": "6/25/2020",
//    "ageRange": "30 to 39",
//    "testType": "Molecular",
//    "result": "Negative",
//    "patientCity": "Las Piedras",
//    "patientId": "5161e02e-8aca-4c50-9a16-0007ec5f5e51",
//    "createdAt": "06/30/2020 13:49"
//  }
type test struct {
	Collected   jsonDate `json:"collectedDate"`
	Reported    jsonDate `json:"reportedDate"`
	AgeRange    ageRange `json:"ageRange"`
	Result      result   `json:"result"`
	PatientID   string   `json:"patientId"`
	PatientCity string   `json:"patientCity"`
	Created     jsonTime `json:"createdAt"`
}

type jsonDate time.Time

func (j *jsonDate) UnmarshalJSON(b []byte) error {
	t, err := unmarshalTime(b, "1/2/2006", true)
	*j = jsonDate(t)
	return err
}

type jsonTime time.Time

func (j *jsonTime) UnmarshalJSON(b []byte) error {
	t, err := unmarshalTime(b, "01/02/2006 15:04", false)
	*j = jsonTime(t)
	return err
}

func unmarshalTime(b []byte, layout string, allowEmpty bool) (time.Time, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return time.Time{}, err
	}
	if s == "" && allowEmpty {
		return time.Time{}, nil
	}
	return time.ParseInLocation(layout, s, loc)
}

type ageRange int

const (
	ageNA ageRange = iota
	age0To9
	age10To19
	age20To29
	age30To39
	age40To49
	age50To59
	age60To69
	age70To79
	age80To89
	age90To99
	age100To109
	age110To119
	age120To129
	age130To139
	age140To149 // Huh?

	ageMin = ageNA
	ageMax = age140To149
)

var ageRangeStrings = map[string]ageRange{
	"N/A":        ageNA,
	"0 to 9":     age0To9,
	"10 to 19":   age10To19,
	"20 to 29":   age20To29,
	"30 to 39":   age30To39,
	"40 to 49":   age40To49,
	"50 to 59":   age50To59,
	"60 to 69":   age60To69,
	"70 to 79":   age70To79,
	"80 to 89":   age80To89,
	"90 to 99":   age90To99,
	"100 to 109": age100To109,
	"110 to 119": age110To119,
	"120 to 129": age120To129,
	"130 to 139": age130To139,
	"140 to 149": age140To149,
}

func (a *ageRange) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*a = ageNA
		return nil
	}
	var ok bool
	if *a, ok = ageRangeStrings[s]; !ok {
		return fmt.Errorf("invalid age range %q", s)
	}
	return nil
}

func (a *ageRange) min() int {
	if *a == ageNA {
		return -1
	}
	return (int(*a) - 1) * 10
}

type testType int

const (
	molecular testType = iota
)

type result int

const (
	positive result = iota
	negative
	other
)

// The strings used for the "result" property are all over the board.
// Descending occurrence as of 20200726:
//
//  "Negative"             283998
//  "Not Detected"          27997
//  "Positive"               5820
//  "COVID-19 Negative"      2001
//  "Positive 2019-nCoV"     1516
//  "Presumptive Positive"    301
//  "Not Tested"              231
//  "Inconclusive"            142
//  "Other"                    62
//  "Not Valid"                23
//  "COVID-19 Positive"        19
//  "Invalid"                  13
//  "Positive IgM and IgG"      4
//  "Positive IgM Only"         4
var resultStrings = map[string]result{
	"Positive":             positive,
	"Positive 2019-nCoV":   positive,
	"Presumptive Positive": positive,
	"COVID-19 Positive":    positive,

	"Negative":          negative,
	"Not Detected":      negative,
	"COVID-19 Negative": negative,

	"Positive IgM and IgG": other, // serological?
	"Positive IgM Only":    other, // serological?
	"Not Tested":           other,
	"Inconclusive":         other,
	"Other":                other,
	"Not Valid":            other,
	"Invalid":              other,
}

func (r *result) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var ok bool
	if *r, ok = resultStrings[s]; !ok {
		return fmt.Errorf("invalid result %q", s)
	}
	return nil
}

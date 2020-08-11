// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestTest_Unmarshal(t *testing.T) {
	const in = `{"collectedDate":"6/23/2020","reportedDate":"6/25/2020","ageRange": "30 to 39","testType":"Molecular","result":"Negative","patientCity":"Las Piedras","patientId":"5161e02e-8aca-4c50-9a16-0007ec5f5e51","createdAt":"06/30/2020 13:49"}`

	var got test
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatal("Unmarshaling failed: ", err)
	}
	if diff := cmp.Diff(test{
		Collected:   jsonDate(makeTime("2006-01-02", "2020-06-23")),
		Reported:    jsonDate(makeTime("2006-01-02", "2020-06-25")),
		AgeRange:    age30To39,
		Type:        molecular,
		Result:      negative,
		PatientID:   "5161e02e-8aca-4c50-9a16-0007ec5f5e51",
		PatientCity: "Las Piedras",
		Created:     jsonTime(makeTime("2006-01-02 15:04", "2020-06-30 13:49")),
	}, got, cmpopts.IgnoreUnexported(jsonDate{}, jsonTime{})); diff != "" {
		t.Error("Didn't unmarshal test correctly:\n" + diff)
	}
}

func makeTime(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(fmt.Sprintf("Failed parsing %q: %v", s, err))
	}
	return t
}

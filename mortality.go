package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v [flags] <YYYYMMDD.csv> ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	state := flag.String("state", "California", `State (as it appears in CSV files)`)
	weekEnd := flag.String("week-end", "20200425", `Week end as "YYYYMMDD"`)
	flag.Parse()

	if y, m, d := parseDate(*weekEnd); y == "" || m == "" || d == "" {
		log.Fatalf("Bad week end date %q; should be YYYYMMDD", *weekEnd)
	}

	for _, p := range flag.Args() {
		if err := readFile(p, *state, *weekEnd); err != nil {
			log.Fatalf("Failed reading %v: %v", p, err)
		}
	}
}

var dateRegexp = regexp.MustCompile(`^(20\d\d)([01]\d)([0-3]\d)$`)

func parseDate(s string) (year, month, day string) {
	m := dateRegexp.FindStringSubmatch(s)
	if len(m) == 0 {
		return "", "", ""
	}
	return m[1], m[2], m[3]
}

func readFile(p, state, weekEnd string) error {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	fileDate := base[:len(base)-len(ext)]
	if y, m, d := parseDate(fileDate); y == "" || m == "" || d == "" {
		return fmt.Errorf("file not named YYYYMMDD.csv")
	}

	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// Find the positions of various columns.
	cols, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed reading header: %v", err)
	}
	var weekEndCol, stateCol, observedCol, typeCol, outcomeCol int
	for name, dst := range map[string]*int{
		"Week Ending Date": &weekEndCol,
		"State":            &stateCol,
		"Observed Number":  &observedCol,
		"Type":             &typeCol,
		"Outcome":          &outcomeCol,
	} {
		found := false
		for i, s := range cols {
			if s == name || s == "\ufeff"+name { // Sigh.
				*dst = i
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing column %q", name)
		}
	}

	// The CDC changed the date format from MM/DD/YYYY to YYYY-MM-DD.
	year, month, day := parseDate(weekEnd)
	weekEnd1 := fmt.Sprintf("%s/%s/%s", month, day, year)
	weekEnd2 := fmt.Sprintf("%s-%s-%s", year, month, day)

	for {
		vals, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if vals[stateCol] == state &&
			(vals[weekEndCol] == weekEnd1 || vals[weekEndCol] == weekEnd2) &&
			vals[typeCol] == "Unweighted" && // vs. "Predicted (weighted)"
			vals[outcomeCol] == "All causes" { // vs. "All causes, excluding COVID-19"
			fmt.Printf("%s %s\n", fileDate, vals[observedCol])
		}
	}
	return nil
}

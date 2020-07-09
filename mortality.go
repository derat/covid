// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const dateLayout = "20060102"

func main() {
	now := time.Now()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v [flags] <YYYYMMDD.csv> ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	action := flag.String("action", "plot", `Action to perform ("plot", "summarize")`)
	state := flag.String("state", "", `State as it appears in CSV files, e.g. "California", or empty for all`)
	start := flag.String("start", now.AddDate(0, -3, 0).Format(dateLayout), `Starting week-ending date`)
	end := flag.String("end", now.Format(dateLayout), `Ending week-ending date`)
	covid := flag.Bool("covid", false, "Show only deaths attributed to COVID-19")
	predicted := flag.Bool("predicted", false, "Use predicted deaths rather than observed")
	excess := flag.Bool("excess", false, "Show excess (vs. upper-bound threshold) deaths")
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if *covid && *excess {
		log.Fatal("Can't use -covid and -excess simultaneously")
	}

	if *state == "" {
		*state = "United States"
	}
	startDate, err := time.Parse(dateLayout, *start)
	if err != nil {
		log.Fatalf("Bad -start date %q: %v", *start, err)
	}
	endDate, err := time.Parse(dateLayout, *end)
	if err != nil {
		log.Fatalf("Bad -end date %q: %v", *end, err)
	}

	// Read the CSV files.
	ds := newDataSet(*state, startDate, endDate, *covid, *predicted, *excess)
	for _, p := range flag.Args() {
		if err := ds.readFile(p); err != nil {
			log.Fatalf("Failed reading %v: %v", p, err)
		}
	}

	switch *action {
	case "plot":
		// Write the data to a temp file.
		dp, err := writeData(ds)
		if err != nil {
			log.Fatal("Failed writing data file: ", err)
		}
		defer os.Remove(dp)

		// Write the gnuplot commands to a temp file.
		gp, err := writeGnuplot(ds, dp)
		if err != nil {
			log.Fatal("Failed writing gnuplot file: ", err)
		}
		defer os.Remove(gp)

		if err := exec.Command("gnuplot", "-p", gp).Run(); err != nil {
			log.Fatal("Failed running gnuplot: ", err)
		}
	case "summarize":
		if err := ds.summarize(os.Stdout); err != nil {
			log.Fatal("Failed writing summary: ", err)
		}
	default:
		log.Fatalf("Invalid action %q", *action)
	}
}

// writeData creates a temp file and writes ds's data to it.
// The file's path is returned.
func writeData(ds *dataSet) (string, error) {
	f, err := ioutil.TempFile("", "mortality.data.")
	if err != nil {
		return "", err
	}
	if err := ds.write(f); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	err = f.Close()
	return f.Name(), err
}

// writeGnuplot writes a gnuplot file for plotting data from dataPath,
// the path returned by an earlier writeData(ds) call.
func writeGnuplot(ds *dataSet, dataPath string) (string, error) {
	f, err := ioutil.TempFile("", "mortality.gnuplot.")
	if err != nil {
		return "", err
	}

	var titleParts []string
	addTitlePart := func(cond bool, a, b string) {
		if cond {
			titleParts = append(titleParts, a)
		} else if b != "" {
			titleParts = append(titleParts, b)
		}
	}
	addTitlePart(true, "CDC Weekly", "")
	addTitlePart(ds.predicted, "Predicted", "Observed")
	addTitlePart(ds.excess, "Excess", "")
	addTitlePart(ds.covid, "COVID-19", "All-Cause")
	addTitlePart(true, "Mortality for", "")
	addTitlePart(ds.state != "", ds.state, "United States")
	title := strings.Join(titleParts, " ")

	if err := template.Must(template.New("").Funcs(map[string]interface{}{
		"indexCol": func(i int) int { return i + 2 },
	}).Parse(`
set title "{{.Title}}\n\n" . \
  "{/*0.8 Source: https://data.cdc.gov/NCHS/Excess-Deaths-Associated-with-COVID-19/xkkf-xrst/\n}" . \
  "{/*0.8 Shows changes to CDC data over time.}"

set xlabel 'Data Update Date'
set xdata time
set timefmt '%Y%m%d'

set ylabel 'Deaths'
set yrange [*<0:*]
set grid xtics ytics

set key autotitle columnheader outside top right title 'Week Ending'

# https://stackoverflow.com/a/57239036
set linetype  1 lc rgb "dark-violet" lw 1 dt 1 pt 0
set linetype  2 lc rgb "#009e73"     lw 1 dt 1 pt 7
set linetype  3 lc rgb "#56b4e9"     lw 1 dt 1 pt 6 pi -1
set linetype  4 lc rgb "#e69f00"     lw 1 dt 1 pt 5 pi -1
set linetype  5 lc rgb "#f0e442"     lw 1 dt 1 pt 8
set linetype  6 lc rgb "#0072b2"     lw 1 dt 1 pt 3
set linetype  7 lc rgb "#e51e10"     lw 1 dt 1 pt 11
set linetype  8 lc rgb "black"       lw 1 dt 1
set linetype  9 lc rgb "dark-violet" lw 1 dt 3 pt 0
set linetype 10 lc rgb "#009e73"     lw 1 dt 3 pt 7
set linetype 11 lc rgb "#56b4e9"     lw 1 dt 3 pt 6 pi -1
set linetype 12 lc rgb "#e69f00"     lw 1 dt 3 pt 5 pi -1
set linetype 13 lc rgb "#f0e442"     lw 1 dt 3 pt 8
set linetype 14 lc rgb "#0072b2"     lw 1 dt 3 pt 3
set linetype 15 lc rgb "#e51e10"     lw 1 dt 3 pt 11
set linetype 16 lc rgb "black"       lw 1 dt 3
set linetype cycle 16

num_lines = {{.NumLines}}
plot for [i=2:num_lines+2] '{{.DataPath}}' using 1:i with lines
`)).Execute(f, struct {
		Title    string
		DataPath string
		NumLines int
	}{title, dataPath, len(ds.weekSeries)}); err != nil {
		f.Close()
		return "", err
	}
	return f.Name(), f.Close()
}

// timeseries contains the values of a variable at different points in time.
type timeseries map[string]int

// dataSet holds data for a state.
// It parses CSV files downloaded on different days and tracks how each week's
// reported mortality has changed over time.
type dataSet struct {
	state      string    // state name, e.g. "California"
	start, end time.Time // start and end dates
	covid      bool
	predicted  bool
	excess     bool
	fileDates  map[string]struct{}   // dates of parsed data as e.g. "20200425"
	weekSeries map[string]timeseries // keyed by week end as e.g. "20200425"
}

// newDataSet returns a new dataSet that saves mortality data for week-ending
// dates between start and end for the supplied state.
func newDataSet(state string, start, end time.Time, covid, predicted, excess bool) *dataSet {
	return &dataSet{
		state:      state,
		start:      start,
		end:        end,
		covid:      covid,
		predicted:  predicted,
		excess:     excess,
		fileDates:  make(map[string]struct{}),
		weekSeries: make(map[string]timeseries),
	}
}

// sortedWeekEnds returns the keys (e.g. "20200425") from ds.weekSeries, sorted in ascending order.
func (ds *dataSet) sortedWeekEnds() []string {
	wes := make([]string, 0, len(ds.weekSeries))
	for we := range ds.weekSeries {
		wes = append(wes, we)
	}
	sort.Strings(wes)
	return wes
}

// sortedFileDates returns the values (e.g. "20200425") from ds.fileDates, sorted in ascending order.
func (ds *dataSet) sortedFileDates() []string {
	fds := make([]string, 0, len(ds.fileDates))
	for fd := range ds.fileDates {
		fds = append(fds, fd)
	}
	sort.Strings(fds)
	return fds
}

// readFile parses the CSV file at p.
// The path's base filename must have the form 'YYYYMMDD.csv'
// (describing the day when the file was downloaded).
func (ds *dataSet) readFile(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	// Extract the date from the filename.
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	fileDate := base[:len(base)-len(ext)]
	if _, err := time.Parse(dateLayout, fileDate); err != nil {
		return fmt.Errorf("file not named YYYYMMDD.csv: %v", err)
	}
	ds.fileDates[fileDate] = struct{}{}

	r := csv.NewReader(f)

	// Find the positions of columns that we care about.
	cols, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed reading header: %v", err)
	}
	var weekEndCol, stateCol, observedCol, thresholdCol, typeCol, outcomeCol int
	for nameList, dst := range map[string]*int{
		"Week Ending Date":                &weekEndCol,
		"State":                           &stateCol,
		"Observed Number":                 &observedCol,
		"Upper Bound Threshold,Threshold": &thresholdCol, // column named changed
		"Type":                            &typeCol,
		"Outcome":                         &outcomeCol,
	} {
		names := strings.Split(nameList, ",")
		found := false
	ColumnLoop:
		for i, s := range cols {
			s = strings.TrimLeft(s, "\ufeff") // sigh
			for _, name := range names {
				if s == name {
					*dst = i
					found = true
					break ColumnLoop
				}
			}
		}
		if !found {
			return fmt.Errorf("missing column %q", nameList)
		}
	}

	// Week-ending dates for which we saw excluding-COVID numbers.
	gotExclCovid := make(map[string]struct{})

	for {
		vals, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if vals[stateCol] != ds.state {
			continue
		}

		if (ds.predicted && vals[typeCol] != "Predicted (weighted)") ||
			(!ds.predicted && vals[typeCol] != "Unweighted") {
			continue
		}

		// The CDC started with dates formatted as MM/DD/YYYY but later changed to YYYY-MM-DD.
		s := vals[weekEndCol]
		weekEnd, err := time.Parse("2006-01-02", s)
		if err != nil {
			if weekEnd, err = time.Parse("01/02/2006", s); err != nil {
				return fmt.Errorf("failed to parse week-ending date %q: %v", s, err)
			}
		}
		if weekEnd.Before(ds.start) || weekEnd.After(ds.end) {
			continue
		}

		s = vals[observedCol]
		if s == "" {
			continue
		}
		observed, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("failed to parse observed value %q: %v", s, err)
		}

		if ds.excess {
			s = vals[thresholdCol]
			if s == "" {
				continue
			}
			threshold, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("failed to parse threshold value %q: %v", s, err)
			}
			observed -= threshold
		}

		ws := weekEnd.Format(dateLayout)
		ts, ok := ds.weekSeries[ws]
		if !ok {
			ts = make(timeseries)
			ds.weekSeries[ws] = ts
		}

		if o := vals[outcomeCol]; o == "All causes" {
			ts[fileDate] += observed
		} else if ds.covid && o == "All causes, excluding COVID-19" {
			ts[fileDate] -= observed
			gotExclCovid[ws] = struct{}{}
		}
	}

	// For recent weeks, actual (as opposed to estimated) excluding-COVID numbers aren't reported.
	// Clear these data points to avoid incorrectly reporting all-cause deaths here.
	if ds.covid {
		for we, ts := range ds.weekSeries {
			if _, ok := gotExclCovid[we]; !ok {
				delete(ts, fileDate)
			}
		}
	}

	return nil
}

// write writes ds's data to w in gnuplot's format, i.e. lines with tab-separated values.
func (ds *dataSet) write(w io.Writer) error {
	var writeErr error
	write := func(s string) {
		if writeErr == nil {
			_, writeErr = io.WriteString(w, s)
		}
	}

	weekEnds := ds.sortedWeekEnds()
	fileDates := ds.sortedFileDates()

	// Put line names on the first line.
	weekStrings := make([]string, len(weekEnds))
	for i, we := range weekEnds {
		t, _ := time.Parse(dateLayout, we)
		weekStrings[i] = t.Format("01/02")
	}
	write("Date\t" + strings.Join(weekStrings, "\t") + "\n")

	// Each following line starts with the reporting date from a file
	// and the per-week data from the file.
	for _, fd := range fileDates {
		vals := make([]string, 0, 1+len(weekEnds))
		vals = append(vals, fd)
		for _, we := range weekEnds {
			if v, ok := ds.weekSeries[we][fd]; ok {
				vals = append(vals, strconv.Itoa(v))
			} else {
				vals = append(vals, "?")
			}
		}
		write(strings.Join(vals, "\t") + "\n")
	}

	return writeErr
}

// summarize summarizes ds's data to w as human-readable text.
func (ds *dataSet) summarize(w io.Writer) error {
	var writeErr error
	writef := func(format string, args ...interface{}) {
		if writeErr == nil {
			_, writeErr = fmt.Fprintf(w, format, args...)
		}
	}

	const dl = "2006-01-02"

	fileDates := ds.sortedFileDates()

	for _, we := range ds.sortedWeekEnds() {
		// Find the highest value.
		ts := ds.weekSeries[we]
		max := 0
		for _, v := range ts {
			if v > max {
				max = v
			}
		}
		if max == 0 {
			continue
		}

		wt, _ := time.Parse(dateLayout, we)
		writef("Week ending %s\n", wt.Format(dl))

		for _, fd := range fileDates {
			if v, ok := ts[fd]; ok {
				ft, _ := time.Parse(dateLayout, fd)
				days := int(math.Round(float64(ft.Sub(wt)) / float64(24*time.Hour)))
				pct := float64(v) / float64(max) * 100
				writef("%s  %5v  %5.1f%% of max after %dd)\n", ft.Format(dl), v, pct, days)
			}
		}
		writef("\n")
	}

	return writeErr
}

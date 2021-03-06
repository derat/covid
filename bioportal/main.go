// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/derat/covid/filewriter"
	"github.com/derat/covid/gnuplot"
)

const (
	// Positive results are reported more quickly than negative results.
	// Positivity is not plotted for days close to the current date.
	positivityDelay = 14 * 24 * time.Hour

	// Maximum value for positivity rates in heatmaps. Larger values are capped.
	positivityMaxRate = 0.2

	// Minimum number of tests to plot in positivity heatmaps.
	// Values based on fewer tests are dropped.
	positivityMinTests = 1

	// Minimum number of positive tests for start of age distribution plot.
	ageDistMinPosTests = 10
)

var (
	loc       *time.Location // PR time zone
	startDate time.Time      // earliest date to accept
)

func init() {
	var err error
	if loc, err = time.LoadLocation("America/Puerto_Rico"); err != nil {
		panic(err)
	}
	startDate = time.Date(2020, 3, 12, 0, 0, 0, 0, loc)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v <input> [out-dir]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if ln := len(flag.Args()); ln == 0 || ln > 2 {
		flag.Usage()
		os.Exit(2)
	}

	fn := flag.Arg(0)
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal("Failed opening input: ", err)
	}
	defer f.Close()

	var r io.Reader = f
	if filepath.Ext(fn) == ".gz" {
		gr, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal("Failed decompressing input: ", err)
		}
		defer gr.Close()
		r = gr
	}

	colStats, repStats, err := readTests(r)
	if err != nil {
		log.Fatal("Failed reading tests: ", err)
	}

	// If an output dir wasn't supplied, just print a summary.
	if len(flag.Args()) < 2 {
		for _, d := range sortedTimes(repStats) {
			fmt.Printf("%s: %s\n", d.Format("2006-01-02"), repStats[d])
		}
		return
	}

	outDir := flag.Arg(1)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal("Failed creating output dir: ", err)
	}

	avgColStats := averageStats(colStats, 7)
	avgRepStats := averageStats(repStats, 7)
	weekColStats := weeklyStats(colStats)
	weekRepStats := weeklyStats(repStats)

	// Find the max 90th-percentile delay so we can use the same scale on delay plots.
	maxDelay := 0
	for _, s := range weekRepStats {
		if v := s.delayPct(90); v > maxDelay {
			maxDelay = v
		}
		if v := s.posDelayPct(90); v > maxDelay {
			maxDelay = v
		}
		if v := s.negDelayPct(90); v > maxDelay {
			maxDelay = v
		}
	}

	// Returns a plot function that writes delay distribution data supplied by f.
	makeDelayDataFunc := func(f func(s *stats, pct float64) int) func(w *filewriter.FileWriter) {
		return func(w *filewriter.FileWriter) {
			w.Printf("Date\t10th\t25th\t50th\t75th\t90th\n")
			for _, week := range sortedTimes(weekRepStats) {
				s := weekRepStats[week]
				w.Printf("%s\t%d\t%d\t%d\t%d\t%d\n", week.Format("2006-01-02"),
					f(s, 10), f(s, 25), f(s, 50), f(s, 75), f(s, 90))
			}
		}
	}

	// Returns a plot function that writes age-stratified heatmap data supplied by f.
	makeAgeFunc := func(m statsMap, f func(s *stats, ar ageRange) interface{},
		maxAge ageRange, maxDate time.Time) func(w *filewriter.FileWriter) {
		return func(w *filewriter.FileWriter) {
			w.Printf("X\tDate\tAge\tValue\n")
			for i, week := range sortedTimes(m) {
				if !maxDate.IsZero() && week.AddDate(0, 0, 7).After(maxDate) {
					break
				}
				s := m[week]
				for ar := age0To9; ar <= maxAge; ar++ {
					w.Printf("%d\t%s\t%d\t%v\n", i, week.Format("01/02"), ar.min(), f(s, ar))
				}
			}
		}
	}

	now := time.Now()

	for _, plot := range []struct {
		out  string                         // output file, e.g. "my-plot.png"
		tmpl string                         // gnuplot template data
		data func(w *filewriter.FileWriter) // writes gnuplot data to w
		vars map[string]interface{}         // extra variables to pass to template
	}{
		{
			out:  "positives-age.png",
			tmpl: ageHeatTmpl,
			data: makeAgeFunc(weekRepStats, func(s *stats, ar ageRange) interface{} { return s.agePos[ar] },
				age100To109, time.Time{}),
			vars: map[string]interface{}{"Units": "positive COVID-19 tests"},
		},
		{
			out:  "positives-age-scaled.png",
			tmpl: ageHeatTmpl,
			data: makeAgeFunc(weekRepStats, func(s *stats, ar ageRange) interface{} {
				pop := unAgePop[ar]
				if pop == 0 {
					return 0
				}
				return int64(math.Round(100000 * float64(s.agePos[ar]) / float64(pop)))
			}, age80To89, time.Time{}),
			vars: map[string]interface{}{"Units": "positive COVID-19 tests per 100,000 people"},
		},
		{
			out:  "positivity-age.png",
			tmpl: ageHeatTmpl,
			data: makeAgeFunc(weekColStats, func(s *stats, ar ageRange) interface{} {
				pos := float64(s.agePos[ar])
				total := pos + float64(s.ageNeg[ar])
				if total < positivityMinTests {
					return 0
				}
				return math.Min(pos/total, positivityMaxRate)
			}, age100To109, now.Add(-positivityDelay)),
			vars: map[string]interface{}{"Units": "COVID-19 test positivity rate", "Collect": true},
		},
		{
			out:  "results-age-scaled.png",
			tmpl: ageHeatTmpl,
			data: makeAgeFunc(weekRepStats, func(s *stats, ar ageRange) interface{} {
				pop := unAgePop[ar]
				if pop == 0 {
					return 0
				}
				return int64(math.Round(100000 * float64(s.agePos[ar]+s.ageNeg[ar]) / float64(pop)))
			}, age80To89, time.Time{}),
			vars: map[string]interface{}{"Units": "total COVID-19 tests per 100,000 people"},
		},
		{
			out:  "test-types.png",
			tmpl: typesTmpl,
			data: func(w *filewriter.FileWriter) {
				w.Printf("Date\tMolecular\tSerological\tAntigen\tUnknown\n")
				for _, d := range sortedTimes(avgRepStats) {
					s := avgRepStats[d]
					w.Printf("%s\t%d\t%d\t%d\t%d\n", d.Format("2006-01-02"), s.total(), s.ab, s.ag, s.unk)
				}
			},
		},
		{
			out:  "positivity.png",
			tmpl: posRateTmpl,
			data: func(w *filewriter.FileWriter) {
				w.Printf("Date\tPositivity\n")
				for _, d := range sortedTimes(avgColStats) {
					if now.Sub(d) < positivityDelay {
						break
					}
					s := avgColStats[d]
					posPct := 100 * float64(s.pos) / float64(s.pos+s.neg)
					w.Printf("%s\t%0.1f\n", d.Format("2006-01-02"), posPct)
				}
			},
		},
		{
			out:  "result-delays.png",
			tmpl: delaysTmpl,
			data: makeDelayDataFunc(func(s *stats, pct float64) int { return s.delayPct(pct) }),
			vars: map[string]interface{}{"TestType": "total", "MaxDelay": maxDelay},
		},
		{
			out:  "positive-result-delays.png",
			tmpl: delaysTmpl,
			data: makeDelayDataFunc(func(s *stats, pct float64) int { return s.posDelayPct(pct) }),
			vars: map[string]interface{}{"TestType": "positive", "MaxDelay": maxDelay},
		},
		{
			out:  "negative-result-delays.png",
			tmpl: delaysTmpl,
			data: makeDelayDataFunc(func(s *stats, pct float64) int { return s.negDelayPct(pct) }),
			vars: map[string]interface{}{"TestType": "negative", "MaxDelay": maxDelay},
		},
		{
			out:  "age-dist.png",
			tmpl: ageDistTmpl,
			data: func(w *filewriter.FileWriter) {
				ars := []ageRange{age0To9, age10To19, age20To29, age30To39, age40To49, age50To59, age60To69, age70To79, age80To89, age90To99}
				w.Printf("Date")
				for _, ar := range ars {
					w.Printf("\t%d-%d", ar.min(), ar.max())
				}
				w.Printf("\n")

				started := false
				for _, d := range sortedTimes(avgColStats) {
					s := avgColStats[d]

					if !started {
						if s.pos < ageDistMinPosTests {
							continue
						}
						started = true
					}

					w.Printf(d.Format("2006-01-02"))
					var total, cumul int
					for _, ar := range ars {
						total += s.agePos[ar]
					}
					for _, ar := range ars {
						cumul += s.agePos[ar]
						w.Printf("\t%0.2f", float64(cumul)/float64(total))
					}
					w.Printf("\n")
				}
			},
		},
	} {
		dp := filepath.Join("/tmp", "bioportal."+plot.out+".dat")
		dw := filewriter.New(dp)
		plot.data(dw)
		if err := dw.Close(); err != nil {
			log.Fatalf("Failed writing data for %v: %v", plot.out, err)
		}
		td := templateData(dp, filepath.Join(outDir, plot.out), now, plot.vars)
		if err := gnuplot.ExecTemplate(plot.tmpl, td); err != nil {
			log.Fatalf("Failed plotting %v: %v", plot.out, err)
		}
		os.Remove(dp)
	}
}

// readTests reads a JSON array of test objects from r and returns daily stats
// aggregated by collection date and by reporting date.
func readTests(r io.Reader) (colStats, repStats statsMap, err error) {
	// Instead of unmarshaling all tests into slice all at once, strip off the
	// opening bracket so we can read them one at a time. See the "Stream"
	// example at https://golang.org/pkg/encoding/json/#Decoder.Decode.
	dec := json.NewDecoder(r)
	if t, err := dec.Token(); err != nil {
		return nil, nil, fmt.Errorf("failed reading opening bracket: %v", err)
	} else if d, ok := t.(json.Delim); !ok || d != '[' {
		return nil, nil, fmt.Errorf("data starts with %v instead of opening bracket", t)
	}

	now := time.Now()
	colStats = make(statsMap)
	repStats = make(statsMap)

	for dec.More() {
		var t test
		if err := dec.Decode(&t); err != nil {
			return nil, nil, fmt.Errorf("failed reading test: %v", err)
		}

		col := time.Time(t.Collected)
		colValid := !col.Before(startDate) && !col.After(now)
		rep := time.Time(t.Reported)
		repValid := !rep.Before(startDate) && !rep.After(now)

		delay := -1
		if colValid && repValid && !col.After(rep) {
			delay = int(math.Round(float64(rep.Sub(col)) / float64(24*time.Hour)))
		}

		if colValid {
			colStats.get(col).update(t.Type, t.Result, t.AgeRange, delay)
		}
		if repValid {
			repStats.get(rep).update(t.Type, t.Result, t.AgeRange, delay)
		}
	}

	if t, err := dec.Token(); err != nil {
		return nil, nil, fmt.Errorf("failed reading closing bracket: %v", err)
	} else if d, ok := t.(json.Delim); !ok || d != ']' {
		return nil, nil, fmt.Errorf("data ends with %v instead of closing bracket", t)
	}
	return colStats, repStats, nil
}

// sortedTimes returns sorted keys from m, which must be a map with time.Time keys.
// See https://stackoverflow.com/a/35366762.
func sortedTimes(m interface{}) []time.Time {
	var keys []time.Time
	for _, k := range reflect.ValueOf(m).MapKeys() {
		keys = append(keys, k.Interface().(time.Time))
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	return keys
}

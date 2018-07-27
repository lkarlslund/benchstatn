// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package benchstat

import (
	"fmt"
	"strings"

	"github.com/lkarlslund/benchstatn/stats"
)

// A Table is a table for display in the benchstat output.
type Table struct {
	Metric      string
	OldNewDelta bool // is this a delta table?
	Configs     []string
	Groups      []string
	Rows        []*Row
}

// A Row is a table row for display in the benchstat output.
type Row struct {
	Benchmark string     // benchmark name
	Group     string     // group name
	Scaler    Scaler     // formatter for stats means
	Metrics   []*Metrics // columns of statistics
	PctDeltas []float64  // unformatted percent change
	Deltas    []string   // formatted percent change
	Notes     []string   // additional information
	Changes   []int      // +1 better, -1 worse, 0 unchanged
}

// Tables returns tables comparing the benchmarks in the collection.
func (c *Collection) Tables() []*Table {
	deltaTest := c.DeltaTest
	if deltaTest == nil {
		deltaTest = UTest
	}
	alpha := c.Alpha
	if alpha == 0 {
		alpha = 0.05
	}

	// Update statistics.
	for _, m := range c.Metrics {
		m.computeStats()
	}

	var tables []*Table
	key := Key{}
	for _, key.Unit = range c.Units {
		table := new(Table)
		table.Configs = c.Configs
		table.Groups = c.Groups
		table.Metric = metricOf(key.Unit)
		table.OldNewDelta = len(c.Configs) == 2 || c.ShowDelta
		for _, key.Group = range c.Groups {
			for _, key.Benchmark = range c.Benchmarks[key.Group] {
				row := &Row{Benchmark: key.Benchmark}
				if len(c.Groups) > 1 {
					// Show group headers if there is more than one group.
					row.Group = key.Group
				}

				for _, key.Config = range c.Configs {
					m := c.Metrics[key]
					if m == nil {
						row.Metrics = append(row.Metrics, new(Metrics))
						continue
					}
					row.Metrics = append(row.Metrics, m)
					if row.Scaler == nil {
						row.Scaler = NewScaler(m.Mean, m.Unit)
					}
				}

				// If there are only two configs being compared, add stats.
				if table.OldNewDelta {
					if len(c.Configs) == 2 {
						// Classic, just compare old and new

						// Add a zero delta change to "old"
						row.PctDeltas = append(row.PctDeltas, 0)
						row.Deltas = append(row.Deltas, "")
						row.Notes = append(row.Notes, "")
						row.Changes = append(row.Changes, 0)

						k0 := key
						k0.Config = c.Configs[0]
						k1 := key
						k1.Config = c.Configs[1]
						old := c.Metrics[k0]
						new := c.Metrics[k1]

						pctdelta, delta, note, change := metricDelta(old, new, deltaTest, alpha)
						if table.Metric == "speed" {
							change = -change
						}
						// Add it
						row.PctDeltas = append(row.PctDeltas, pctdelta)
						row.Deltas = append(row.Deltas, delta)
						row.Notes = append(row.Notes, note)
						row.Changes = append(row.Changes, change)

					} else {
						// Find best case
						var baseline *Metrics
						for _, config := range c.Configs {
							k := key
							k.Config = config
							contender := c.Metrics[k]
							if baseline == nil {
								baseline = contender
							} else {
								_, _, _, change := metricDelta(baseline, contender, deltaTest, alpha)
								if table.Metric != "speed" {
									change = -change
								}
								if change < 0 && !c.DeltaWorst {
									baseline = contender
								}
							}
						}

						// Compare everything to this
						for _, config := range c.Configs {
							k := key
							k.Config = config
							thismetric := c.Metrics[k]

							if baseline == thismetric && thismetric != nil {
								// No changes aginst ourselves
								// Add a zero delta change to "old"
								row.PctDeltas = append(row.PctDeltas, 0)
								if thismetric != nil {
									if c.DeltaWorst {
										row.Deltas = append(row.Deltas, "Worst")
									} else {
										row.Deltas = append(row.Deltas, "Best")
									}
								} else {
									row.Deltas = append(row.Deltas, "")
								}
								row.Notes = append(row.Notes, "")
								row.Changes = append(row.Changes, 0)
							} else {
								pctdelta, delta, note, change := metricDelta(baseline, thismetric, deltaTest, alpha)
								if table.Metric != "speed" {
									change = -change
								}

								// Add it
								row.PctDeltas = append(row.PctDeltas, pctdelta)
								row.Deltas = append(row.Deltas, delta)
								row.Notes = append(row.Notes, note)
								row.Changes = append(row.Changes, change)
							}
						}

						// Compare to this one
					}
				}

				table.Rows = append(table.Rows, row)
			}
		}

		if len(table.Rows) > 0 {
			if c.Order != nil {
				Sort(table, c.Order)
			}
			if c.AddGeoMean {
				addGeomean(c, table, key.Unit, table.OldNewDelta)
			}
			tables = append(tables, table)
		}
	}

	return tables
}

var metricSuffix = map[string]string{
	"ns/op": "time/op",
	"ns/GC": "time/GC",
	"B/op":  "alloc/op",
	"MB/s":  "speed",
}

func metricDelta(old, new *Metrics, deltaTest DeltaTest, alpha float64) (pctdelta float64, delta string, note string, change int) {
	if old == nil || new == nil {
		return
	}

	pval, testerr := deltaTest(old, new)
	delta = "~"
	if testerr == stats.ErrZeroVariance {
		note = "(zero variance)"
	} else if testerr == stats.ErrSampleSize {
		note = "(too few samples)"
	} else if testerr == stats.ErrSamplesEqual {
		note = "(all equal)"
	} else if testerr != nil {
		note = fmt.Sprintf("(%s)", testerr)
	} else if pval < alpha {
		if new.Mean == old.Mean {
			delta = "0.00%"
		} else {
			pct := ((new.Mean / old.Mean) - 1.0) * 100.0
			pctdelta = pct
			delta = fmt.Sprintf("%+.2f%%", pct)
			if pct < 0 { // smaller is better, except speeds
				change = +1
			} else {
				change = -1
			}
		}
	}
	if note == "" && pval != -1 {
		note = fmt.Sprintf("(p=%0.3f n=%d+%d)", pval, len(old.RValues), len(new.RValues))
	}
	return
}

// metricOf returns the name of the metric with the given unit.
func metricOf(unit string) string {
	if s := metricSuffix[unit]; s != "" {
		return s
	}
	for s, suff := range metricSuffix {
		if dashs := "-" + s; strings.HasSuffix(unit, dashs) {
			prefix := strings.TrimSuffix(unit, dashs)
			return prefix + "-" + suff
		}
	}
	return unit
}

// addGeomean adds a "geomean" row to the table,
// showing the geometric mean of all the benchmarks.
func addGeomean(c *Collection, t *Table, unit string, delta bool) {
	row := &Row{Benchmark: "[Geo mean]"}
	key := Key{Unit: unit}
	geomeans := []float64{}
	maxCount := 0
	for _, key.Config = range c.Configs {
		var means []float64
		for _, key.Group = range c.Groups {
			for _, key.Benchmark = range c.Benchmarks[key.Group] {
				m := c.Metrics[key]
				// Omit 0 values from the geomean calculation,
				// as these either make the geomean undefined
				// or zero (depending on who you ask). This
				// typically comes up with things like
				// allocation counts, where it's fine to just
				// ignore the benchmark.
				if m != nil && m.Mean != 0 {
					means = append(means, m.Mean)
				}
			}
		}
		if len(means) > maxCount {
			maxCount = len(means)
		}
		if len(means) == 0 {
			row.Metrics = append(row.Metrics, new(Metrics))
			delta = false
		} else {
			geomean := stats.GeoMean(means)
			geomeans = append(geomeans, geomean)
			if row.Scaler == nil {
				row.Scaler = NewScaler(geomean, unit)
			}
			row.Metrics = append(row.Metrics, &Metrics{
				Unit: unit,
				Mean: geomean,
			})
		}
	}
	if maxCount <= 1 {
		// Only one benchmark contributed to this geomean.
		// Since the geomean is the same as the benchmark
		// result, don't bother outputting it.
		return
	}

	if delta {
		for i := range geomeans {
			if len(geomeans) != 2 || i != 1 {
				row.PctDeltas = append(row.PctDeltas, 0)
				row.Deltas = append(row.Deltas, "")
				row.Changes = append(row.Changes, 0)
				row.Notes = append(row.Notes, "")
			} else {
				pct := ((geomeans[1] / geomeans[0]) - 1.0) * 100.0
				row.PctDeltas = append(row.PctDeltas, pct)
				row.Deltas = append(row.Deltas, fmt.Sprintf("%+.2f%%", pct))
				row.Changes = append(row.Changes, 0)
				row.Notes = append(row.Notes, "")
			}
		}
	}
	t.Rows = append(t.Rows, row)
}

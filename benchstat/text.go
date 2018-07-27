// Copyright 2017 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package benchstat

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// FormatText appends a fixed-width text formatting of the tables to w.
func FormatText(w io.Writer, tables []*Table) {
	var textTables [][]*textRow
	for _, t := range tables {
		textTables = append(textTables, toText(t))
	}

	var max []int
	for _, table := range textTables {
		for _, row := range table {
			if len(row.cols) == 1 {
				// Header row
				continue
			}
			for len(max) < len(row.cols) {
				max = append(max, 0)
			}
			for i, s := range row.cols {
				n := utf8.RuneCountInString(s)
				if max[i] < n {
					max[i] = n
				}
			}
		}
	}

	for ti, table := range textTables {
		if ti > 0 {
			fmt.Fprintf(w, "\n")
		}

		// headings
		row := table[0]
		for i, s := range row.cols {
			switch i {
			case 0:
				fmt.Fprintf(w, "%-*s", max[i], s)
			default:
				fmt.Fprintf(w, "  %-*s", max[i], s)
			}
		}
		fmt.Fprintln(w, "")

		// data
		for _, row := range table[1:] {
			for i, s := range row.cols {
				switch {
				case len(row.cols) == 1:
					// Single statistics
					fmt.Fprint(w, s)
				case i == 0:
					// Test name
					fmt.Fprintf(w, "%-*s", max[i], s)
				default:
					// Is this a delta, or text?
					isnote := tables[ti].OldNewDelta && ((len(row.cols) > 5 && i%3 == 0) || i == len(row.cols)-1)
					if isnote {
						// Left-align notes
						fmt.Fprintf(w, "  %-*s", max[i], s)
						break
					}
					fmt.Fprintf(w, "  %*s", max[i], s)
				}
			}
			fmt.Fprintf(w, "\n")
		}
	}
}

// A textRow is a row of printed text columns.
type textRow struct {
	cols []string
}

func newTextRow(cols ...string) *textRow {
	return &textRow{cols: cols}
}

func (r *textRow) add(col string) {
	r.cols = append(r.cols, col)
}

func (r *textRow) trim() {
	for len(r.cols) > 0 && r.cols[len(r.cols)-1] == "" {
		r.cols = r.cols[:len(r.cols)-1]
	}
}

// toText converts the Table to a textual grid of cells,
// which can then be printed in fixed-width output.
func toText(t *Table) []*textRow {
	var textRows []*textRow
	switch len(t.Configs) {
	case 1:
		textRows = append(textRows, newTextRow("name", t.Metric))
	case 2:
		textRows = append(textRows, newTextRow("name", "old "+t.Metric, "new "+t.Metric, "delta"))
	default:
		row := newTextRow("name \\ " + t.Metric)
		for _, config := range t.Configs {
			row.cols = append(row.cols, config)
			if t.OldNewDelta {
				row.cols = append(row.cols, "delta", "note")
			}
		}
		textRows = append(textRows, row)
	}

	var group string

	for _, row := range t.Rows {
		if row.Group != group {
			group = row.Group
			textRows = append(textRows, newTextRow(group))
		}
		text := newTextRow(row.Benchmark)
		for i, m := range row.Metrics {
			text.cols = append(text.cols, m.Format(row.Scaler))
			if t.OldNewDelta && (len(row.Metrics) > 2 || i > 0) {
				delta := row.Deltas[i]
				if delta == "~" {
					delta = "~   "
				}
				text.cols = append(text.cols, delta)
				text.cols = append(text.cols, row.Notes[i])
			}
		}
		textRows = append(textRows, text)
	}
	for _, r := range textRows {
		r.trim()
	}
	return textRows
}

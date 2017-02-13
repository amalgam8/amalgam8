// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package terminal

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
)

// Table .
type Table interface {
	AddRow(row []string)
	AddMultipleRows(rows [][]string)
	SetHeader(header []string)
	SortByColumnIndex(col int)
	SortByColumnHeader(header string)
	PrintTable()
}

// table .
type table struct {
	header       []string
	body         [][]string
	sortByCol    int
	sortByHeader string
	outout       io.Writer
}

// NewTable .
func NewTable(w io.Writer) Table {
	return &table{
		outout:    w,
		sortByCol: -1,
	}
}

// AddRow .
func (t *table) AddRow(row []string) {
	t.body = append(t.body, row)
}

// AddMultipleRows .
func (t *table) AddMultipleRows(rows [][]string) {
	for _, row := range rows {
		t.AddRow(row)
	}
}

// SetHeader .
func (t *table) SetHeader(header []string) {
	t.header = header
}

// PrintTable formats and prints a table.
// +-----------+-----------+-----------+
// | Header 1  | Header 2  | Header 3  |
// +-----------+-----------+-----------+
// | cell 11   | cell 12   | cell 13   |
// | cell 21   | cell 22   | cell 23   |
// +-----------+-----------+-----------+.
func (t *table) PrintTable() {
	var tableBuf bytes.Buffer
	sort.Sort(table(*t))
	t.formatTable(&tableBuf, t.header, t.body)
	rows := strings.Split(tableBuf.String(), "\n")
	border := t.border(rows[0])
	tableBuf.Reset()
	fmt.Fprintf(&tableBuf, "%s\n", border)
	fmt.Fprintf(&tableBuf, "%s\n", rows[0])
	fmt.Fprintf(&tableBuf, "%s\n", border)
	for i := 1; i < len(rows)-1; i++ {
		fmt.Fprintf(&tableBuf, "%s\n", rows[i])
	}
	fmt.Fprintf(&tableBuf, "%s\n", border)
	fmt.Fprintf(t.outout, "%s\n", tableBuf.String())
}

// SortByColumnIndex .
func (t *table) SortByColumnIndex(col int) {
	t.sortByCol = col
}

// SortByColumnHeader .
func (t *table) SortByColumnHeader(header string) {
	t.sortByHeader = header
}

func (t *table) formatTable(table io.Writer, header []string, body [][]string) {
	w := tabwriter.NewWriter(table, 0, 0, 1, ' ', tabwriter.Debug|tabwriter.TabIndent)
	println()
	// Construct Header
	for _, cell := range header {
		fmt.Fprintf(w, "\t %s", cell)
	}
	fmt.Fprintln(w, "\t")
	// Construct Body
	for _, row := range body {
		for _, cell := range row {
			fmt.Fprintf(w, "\t %s", cell)
		}
		fmt.Fprintln(w, "\t")
	}
	w.Flush()
}

func (t *table) border(sample string) string {
	rule := regexp.MustCompile("[^|\033\033[0m]")
	sample = rule.ReplaceAllString(sample, "-")
	rule = regexp.MustCompile("[|]")
	sample = rule.ReplaceAllString(sample, "+")
	return strings.Replace(sample, "m", "-", -1)
}

func (t table) Len() int      { return len(t.body) }
func (t table) Swap(i, j int) { t.body[i], t.body[j] = t.body[j], t.body[i] }
func (t table) Less(i, j int) bool {
	// Validate header
	if t.sortByHeader != "" {
		for i, v := range t.header {
			if v == t.sortByHeader {
				t.sortByCol = i
			}
		}
	}

	// Validate column number
	if t.sortByCol < 0 || t.sortByCol >= len(t.body[0]) {
		t.sortByCol = -1
		return false
	}

	return t.body[i][t.sortByCol] < t.body[j][t.sortByCol]
}

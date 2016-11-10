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
	"strings"
	"text/tabwriter"
)

// PrintTable formats and prints a table.
// +-----------+-----------+-----------+
// | Header 1  | Header 2  | Header 3  |
// +-----------+-----------+-----------+
// | cell 11   | cell 12   | cell 13   |
// | cell 21   | cell 22   | cell 23   |
// +-----------+-----------+-----------+.
func (t *term) PrintTable(header []string, body [][]string) {
	var table bytes.Buffer
	t.formatTable(&table, header, body)
	rows := strings.Split(table.String(), "\n")
	border := t.border(rows[0])
	table.Reset()
	fmt.Fprintf(&table, "%s\n", border)
	fmt.Fprintf(&table, "%s\n", rows[0])
	fmt.Fprintf(&table, "%s\n", border)
	for i := 1; i < len(rows)-1; i++ {
		fmt.Fprintf(&table, "%s\n", rows[i])
	}
	fmt.Fprintf(&table, "%s\n", border)
	fmt.Fprintf(t.Output, "%s\n", table.String())
}

func (t *term) formatTable(table io.Writer, header []string, body [][]string) {
	w := tabwriter.NewWriter(table, 0, 0, 1, ' ', tabwriter.Debug|tabwriter.TabIndent)
	println()
	// Construct Header
	for _, cell := range header {
		// TODO: Fix format for windows
		// if !strings.Contains(cell, "\033[0m") {
		// 	cell = fmt.Sprint(t.FontColor(Black, Normal, cell))
		// }
		fmt.Fprintf(w, "\t %s", cell)
	}
	fmt.Fprintln(w, "\t")
	// Construct Body
	for _, row := range body {
		for _, cell := range row {
			// TODO: Fix format for windows
			// if !strings.Contains(cell, "\033[0m") {
			// 	cell = fmt.Sprint(t.FontColor(Black, Normal, cell))
			// }
			fmt.Fprintf(w, "\t %s", cell)
		}
		fmt.Fprintln(w, "\t")
	}
	w.Flush()
}

func (t *term) border(sample string) string {
	rule := regexp.MustCompile("[^|\033\033[0m]")
	sample = rule.ReplaceAllString(sample, "-")
	rule = regexp.MustCompile("[|]")
	return rule.ReplaceAllString(sample, "+")
}

// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docs

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// newMarkdownTable returns a configured md table with specified headings
func newMarkdownTable(headings []string, buf *bytes.Buffer) tablewriter.Table {
	table := tablewriter.NewTable(buf,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithRowAutoWrap(0),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.Off},
				Alignment:  tw.CellAlignment{Global: tw.AlignNone},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignNone},
			},
		}),
	)
	table.Header(headings)
	return *table
}

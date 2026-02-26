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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertBetweenIdx(t *testing.T) {
	type args struct {
		original []string
		start    int
		stop     int
		toInsert []string
	}
	tests := []struct {
		name string
		args args
		want []string
		err  string
	}{
		{
			name: "simple",
			args: args{
				original: []string{"a", "b", "c", "d", "e", "f"},
				start:    1,
				stop:     2,
				toInsert: []string{"x", "y"},
			},
			want: []string{"a", "x", "y", "d", "e", "f"},
		},
		{
			name: "small original",
			args: args{
				original: []string{"a", "b", "c"},
				start:    0,
				stop:     1,
				toInsert: []string{"x", "y", "z"},
			},
			want: []string{"x", "y", "z", "c"},
		},
		{
			name: "negative start",
			args: args{
				original: []string{"a", "b", "c"},
				start:    -1,
				stop:     1,
				toInsert: []string{"x", "y", "z"},
			},
			err: "unable insert, invalid start: -1 or stop: 1",
		},
		{
			name: "oob start",
			args: args{
				original: []string{"a", "b", "c"},
				start:    3,
				stop:     4,
				toInsert: []string{"x", "y", "z"},
			},
			err: "unable insert, invalid start: 3 or stop: 4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := insertBetweenIdx(tt.args.original, tt.args.start, tt.args.stop, tt.args.toInsert)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.want, got)
			}

		})
	}
}

func TestFindInsertPoint(t *testing.T) {
	type args struct {
		doc            string
		docMarkerStart string
		docMarkerEnd   string
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
		err       string
	}{
		{
			name: "simple",
			args: args{
				doc: `lorem
ipsom
markerstart
dolor
markerend`,
				docMarkerStart: "markerstart",
				docMarkerEnd:   "markerend"},
			wantStart: 3,
			wantEnd:   5,
		},
		{
			name: "missing start marker",
			args: args{
				doc: `lorem
ipsom
dolor
markerend`,
				docMarkerStart: "markerstart",
				docMarkerEnd:   "markerend"},
			err: "unable to find start marker: markerstart",
		},
		{
			name: "missing end marker",
			args: args{
				doc: `lorem
ipsom
markerstart
dolor`,
				docMarkerStart: "markerstart",
				docMarkerEnd:   "markerend"},
			err: "unable to find end marker: markerend",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := findInsertPoint(strings.Split(tt.args.doc, "\n"), tt.args.docMarkerStart, tt.args.docMarkerEnd)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.wantStart, gotStart)
				require.Equal(tt.wantStart, gotEnd)
			}
		})
	}
}

func TestInsertIntoReadme(t *testing.T) {
	type args struct {
		title     string
		current   string
		generated string
	}
	tests := []struct {
		name string
		args args
		want string
		err  string
	}{
		{
			name: "simple",
			args: args{
				title: "# sample",
				current: `<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->

<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->
<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->

<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->`,
				generated: `This is new
generated content that should replace old content`,
			},
			want: `<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->
# sample
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->
<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
This is new
generated content that should replace old content
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->`,
		},
		{
			name: "with existing",
			args: args{
				title: "# sample",
				current: `<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->
# Old title
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->

Custom content added by user

<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
This is
old generated
content
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->`,
				generated: `This is new
generated content that should replace old content`,
			},
			want: `<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->
# sample
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->

Custom content added by user

<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
This is new
generated content that should replace old content
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->`,
		},
		{
			name: "no title hooks",
			args: args{
				title: "# sample",
				current: `Custom content added by user

<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
This is
old generated
content
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
more custom info
`,
				generated: `This is new
generated content that should replace old content`,
			},
			want: `Custom content added by user

<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
This is new
generated content that should replace old content
<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->
more custom info
`,
		},
		{
			name: "no body hooks",
			args: args{
				title: "# sample",
				current: `Custom content added by user

This is
old generated
content
more custom info`,
				generated: `This is new
generated content that should replace old content`,
			},
			want: `Custom content added by user

This is
old generated
content
more custom info`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGenerated, err := InsertIntoReadme(tt.args.title, tt.args.current, tt.args.generated)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.want, gotGenerated)
			}
		})
	}
}

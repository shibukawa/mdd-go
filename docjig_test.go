package mdd

import (
	"embed"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelMatch(t *testing.T) {
	type Doc struct {
		Name string
	}

	type args struct {
		primaryLabel string
		inputLabel   string
	}
	tests := []struct {
		name       string
		args       args
		wantSuffix string
		wantMatch  bool
	}{
		{
			name: "match with original",
			args: args{
				primaryLabel: "Sample",
				inputLabel:   "Sample",
			},
			wantSuffix: "",
			wantMatch:  true,
		},
		{
			name: "match with original (case insensitive)",
			args: args{
				primaryLabel: "Sample",
				inputLabel:   "sample",
			},
			wantSuffix: "",
			wantMatch:  true,
		},
		{
			name: "Get Suffix",
			args: args{
				primaryLabel: "Sample",
				inputLabel:   "Sample: Suffix",
			},
			wantSuffix: "Suffix",
			wantMatch:  true,
		},
		{
			name: "simple alias",
			args: args{
				primaryLabel: "Sample",
				inputLabel:   "Example",
			},
			wantSuffix: "",
			wantMatch:  true,
		},
		{
			name: "alias with lang",
			args: args{
				primaryLabel: "Sample",
				inputLabel:   "サンプル: Suffix",
			},
			wantSuffix: "Suffix",
			wantMatch:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jig := NewDocJig[Doc]()
			jig.Alias("Sample", "Example").Lang("ja", "例", "サンプル")
			got, ok := jig.matchLabel(tc.args.primaryLabel, tc.args.inputLabel)
			assert.Equal(t, tc.wantSuffix, got)
			assert.Equal(t, tc.wantMatch, ok)
		})
	}
}

func TestGlob(t *testing.T) {
	type Doc struct {
		Title     string
		CodeBlock string
	}

	jig := NewDocJig[Doc]()
	root := jig.Root()
	root.Label("Title")
	root.CodeFence("CodeBlock")

	type args struct {
		patterns []string
	}
	tests := []struct {
		name      string
		args      args
		want      map[string]*Doc
		wantError string
	}{
		{
			name: "find single",
			args: args{
				patterns: []string{"testdata/sample.md"},
			},
			want: map[string]*Doc{
				"testdata/sample.md": {
					Title:     "File in root folder",
					CodeBlock: "parent file content",
				},
			},
		},
		{
			name: "glob match",
			args: args{
				patterns: []string{"testdata/*"},
			},
			want: map[string]*Doc{
				"testdata/sample.md": {
					Title:     "File in root folder",
					CodeBlock: "parent file content",
				},
			},
		},
		{
			name: "glob match including child folder",
			args: args{
				patterns: []string{"testdata/**/*"},
			},
			want: map[string]*Doc{
				"testdata/subfolder/child.md": {
					Title:     "Child folder file",
					CodeBlock: "child file content",
				},
			},
		},
		{
			name: "glob match including child folder",
			args: args{
				patterns: []string{"testdata/**/*", "testdata/*"},
			},
			want: map[string]*Doc{
				"testdata/sample.md": {
					Title:     "File in root folder",
					CodeBlock: "parent file content",
				},
				"testdata/subfolder/child.md": {
					Title:     "Child folder file",
					CodeBlock: "child file content",
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := jig.ParseGlob(tc.args.patterns...)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

//go:embed testdata/*
var testFixtures embed.FS

func TestFS(t *testing.T) {
	type Doc struct {
		Title     string
		CodeBlock string
	}

	jig := NewDocJig[Doc]()
	root := jig.Root()
	root.Label("Title")
	root.CodeFence("CodeBlock")

	type args struct {
		patterns []string
	}
	tests := []struct {
		name      string
		args      args
		want      map[string]*Doc
		wantError string
	}{
		{
			name: "find single",
			args: args{
				patterns: []string{"testdata/sample.md"},
			},
			want: map[string]*Doc{
				"testdata/sample.md": {
					Title:     "File in root folder",
					CodeBlock: "parent file content",
				},
			},
		},
		{
			name: "glob match including child folder",
			args: args{
				patterns: []string{"testdata/**/*"},
			},
			want: map[string]*Doc{
				"testdata/subfolder/child.md": {
					Title:     "Child folder file",
					CodeBlock: "child file content",
				},
			},
		},
		{
			name: "glob match including child folder",
			args: args{
				patterns: []string{"testdata/**/*", "testdata/*"},
			},
			want: map[string]*Doc{
				"testdata/sample.md": {
					Title:     "File in root folder",
					CodeBlock: "parent file content",
				},
				"testdata/subfolder/child.md": {
					Title:     "Child folder file",
					CodeBlock: "child file content",
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := jig.ParseFS(testFixtures, tc.args.patterns...)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

type PostProcessTestDoc struct {
	PostProcessCalled bool
}

func (d *PostProcessTestDoc) PostProcess() {
	d.PostProcessCalled = true
}

func TestPostProcessOK(t *testing.T) {
	jig := NewDocJig[PostProcessTestDoc]()
	got, err := jig.ParseString(`# Test`)
	assert.NoError(t, err)
	assert.True(t, got.PostProcessCalled)
}

var PostProcessError = errors.New("dummy error")

type PostProcessTestDocNG struct {
}

func (d *PostProcessTestDocNG) PostProcess() error {
	return PostProcessError
}

func TestPostProcessNG(t *testing.T) {
	jig := NewDocJig[PostProcessTestDocNG]()
	got, err := jig.ParseString(`# Test`)
	assert.Error(t, err, PostProcessError.Error())
	assert.Nil(t, got)
}

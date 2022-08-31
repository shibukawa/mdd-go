package mdd

import (
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

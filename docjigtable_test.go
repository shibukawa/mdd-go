package mdd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	type CustomCell struct {
		Name  string
		Value string
	}
	type DefinedType int
	const (
		TRUE  DefinedType = 0
		FALSE DefinedType = 1
	)
	type Row struct {
		StringCell  string
		IntCell     int
		DefinedType DefinedType
		BoolCell    bool
		CustomCell  CustomCell
		CustomCellP *CustomCell
	}

	type Doc struct {
		TableContent []Row
	}

	type args struct {
		create func(t *testing.T) *DocJig[Doc]
		src    string
	}
	tests := []struct {
		name    string
		args    args
		want    *Doc
		wantErr string
	}{
		{
			name: "simple table",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("IntCell")
					table.Field("BoolCell")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | IntCell | BoolCell |
				|------------|---------|----------|
				| hello      | 5       | true     |
				| world!!    | 7       | false    |
				`),
			},
			want: &Doc{
				TableContent: []Row{
					{
						StringCell: "hello",
						IntCell:    5,
						BoolCell:   true,
					},
					{
						StringCell: "world!!",
						IntCell:    7,
						BoolCell:   false,
					},
				},
			},
		},
		{
			name: "table cell translation",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					jig.Alias("StringCell").Lang("ja", "文字列")
					jig.Alias("IntCell").Lang("ja", "数値")
					jig.Alias("BoolCell").Lang("ja", "ブール")
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("IntCell")
					table.Field("BoolCell")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| 文字列      | 数値     | ブール |
				|------------|---------|----------|
				| hello      | 5       | true     |
				| world!!    | 7       | false    |
				`),
			},
			want: &Doc{
				TableContent: []Row{
					{
						StringCell: "hello",
						IntCell:    5,
						BoolCell:   true,
					},
					{
						StringCell: "world!!",
						IntCell:    7,
						BoolCell:   false,
					},
				},
			},
		},
		{
			name: "required flag (no field)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("IntCell")
					table.Field("BoolCell").Required()
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | IntCell |
				|------------|---------|
				| hello      | 5       |
				| world!!    | 7       |
				`),
			},
			wantErr: "required column(BoolCell) are missing (inside 'Root Heading' section)",
		},
		{
			name: "custom cell (defined type)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("DefinedType").As(func(v string, d *Doc) (any, error) {
						maps := map[string]DefinedType{
							"true":  TRUE,
							"false": FALSE,
						}
						return maps[v], nil
					})
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | DefinedType |
				|------------|-------------|
				| hello      | true        |
				| world!!    | false       |
				`),
			},
			want: &Doc{
				TableContent: []Row{
					{
						StringCell:  "hello",
						DefinedType: TRUE,
					},
					{
						StringCell:  "world!!",
						DefinedType: FALSE,
					},
				},
			},
		},
		{
			name: "custom cell (value)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("CustomCell", "KeyValue").As(func(v string, d *Doc) (any, error) {
						key, value, ok := strings.Cut(v, ":")
						if !ok {
							return nil, fmt.Errorf("cell content '%s' is invalid", v)
						}
						return CustomCell{Name: key, Value: value}, nil
					})
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | KeyValue |
				|------------|----------|
				| hello      | length:5 |
				| world!!    | length:7 |
				`),
			},
			want: &Doc{
				TableContent: []Row{
					{
						StringCell: "hello",
						CustomCell: CustomCell{
							Name:  "length",
							Value: "5",
						},
					},
					{
						StringCell: "world!!",
						CustomCell: CustomCell{
							Name:  "length",
							Value: "7",
						},
					},
				},
			},
		},
		{
			name: "custom cell (pointer)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("CustomCellP", "KeyValue").As(func(v string, d *Doc) (any, error) {
						key, value, ok := strings.Cut(v, ":")
						if !ok {
							return nil, fmt.Errorf("cell content '%s' is invalid", v)
						}
						return &CustomCell{Name: key, Value: value}, nil
					})
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | KeyValue |
				|------------|----------|
				| hello      | length:5 |
				| world!!    | length:7 |
				`),
			},
			want: &Doc{
				TableContent: []Row{
					{
						StringCell: "hello",
						CustomCellP: &CustomCell{
							Name:  "length",
							Value: "5",
						},
					},
					{
						StringCell: "world!!",
						CustomCellP: &CustomCell{
							Name:  "length",
							Value: "7",
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jig := tc.args.create(t)
			got, err := jig.ParseString(tc.args.src)
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestTableStoresAnyCells(t *testing.T) {
	type Doc struct {
		TableContent []map[string]string
	}

	type args struct {
		create func(t *testing.T) *DocJig[Doc]
		src    string
	}
	tests := []struct {
		name    string
		args    args
		want    *Doc
		wantErr string
	}{
		{
			name: "simple table",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Table("TableContent").AsMap()
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				| StringCell | IntCell | BoolCell |
				|------------|---------|----------|
				| hello      | 5       | true     |
				| world!!    | 7       | false    |
				`),
			},
			want: &Doc{
				TableContent: []map[string]string{
					{
						"StringCell": "hello",
						"IntCell":    "5",
						"BoolCell":   "true",
					},
					{
						"StringCell": "world!!",
						"IntCell":    "7",
						"BoolCell":   "false",
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jig := tc.args.create(t)
			got, err := jig.ParseString(tc.args.src)
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestRepeatableTable(t *testing.T) {
	type SingleTable struct {
		Name  string
		Cells []map[string]string
	}

	type Doc struct {
		TableContents []SingleTable
	}

	type args struct {
		create func(t *testing.T) *DocJig[Doc]
		src    string
	}
	tests := []struct {
		name    string
		args    args
		want    *Doc
		wantErr string
	}{
		{
			name: "simple table",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					tables := root.Children("TableContents", "Table")
					tables.Label("Name")
					tables.Table("Cells").AsMap()
					return jig
				},
				src: TrimIndent(t, `
				# Test Fixtures

				## Table: Normal User

				| Id | Name  | Email             |
				|----|-------|-------------------|
				| 1  | Alice | alice@example.com |
				| 2  | Bob   | bob@example.com   |

				## Table: User Org

				| Id | Name  |
				|----|-------|
				| 1  | Legal |
				| 2  | R&D   |
				`),
			},
			want: &Doc{
				TableContents: []SingleTable{
					{
						Name: "Normal User",
						Cells: []map[string]string{
							{
								"Id":    "1",
								"Name":  "Alice",
								"Email": "alice@example.com",
							},
							{
								"Id":    "2",
								"Name":  "Bob",
								"Email": "bob@example.com",
							},
						},
					},
					{
						Name: "User Org",
						Cells: []map[string]string{
							{
								"Id":   "1",
								"Name": "Legal",
							},
							{
								"Id":   "2",
								"Name": "R&D",
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jig := tc.args.create(t)
			got, err := jig.ParseString(tc.args.src)
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestTable_Generate(t *testing.T) {
	type Row struct {
		StringCell string
		IntCell    int
		BoolCell   bool
	}

	type Doc struct {
		TableContent []Row
	}

	type args struct {
		create func(t *testing.T) *DocJig[Doc]
		opt    GenerateOption
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr string
	}{
		{
			name: "table",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell").Samples("a", "b")
					table.Field("IntCell").Samples(1, 2)
					table.Field("BoolCell").Samples(true, false)
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Title]

				| StringCell | IntCell | BoolCell |
				|------------|---------|----------|
				| a          | 1       | true     |
				| b          | 2       | false    |
				`),
		},
		{
			name: "no sample",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					table := root.Table("TableContent")
					table.Field("StringCell")
					table.Field("IntCell")
					table.Field("BoolCell")
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Title]

				| StringCell | IntCell | BoolCell |
				|------------|---------|----------|
				| ...        | ...     | ...      |
				| ...        | ...     | ...      |
				`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jig := tc.args.create(t)
			var buffer bytes.Buffer
			err := jig.GenerateTemplate(&buffer, tc.args.opt)
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, strings.TrimRight(buffer.String(), "\n"))
			}
		})
	}
}

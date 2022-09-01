package mdd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeFence(t *testing.T) {
	type Doc struct {
		Code string
		Lang string
		Info string
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
			name: "single code fence: just code",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.CodeFence("Code")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				~~~sql
				SELECT email, first_name, last_name FROM persons;
				~~~
				`),
			},

			want: &Doc{
				Code: "SELECT email, first_name, last_name FROM persons;",
			},
		},
		{
			name: "single code fence: match language code (not match)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.CodeFence("Code", "sql")
					return jig
				},
				src: TrimIndent(t, `
					# Root Heading

					~~~yaml
					hello: world
					~~~
					`),
			},
			want: &Doc{
				Code: "",
			},
		},
		{
			name: "single code fence: store lang and info",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.CodeFence("Code").Language("Lang").Info("Info")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				~~~sql :test
				SELECT email, first_name, last_name FROM persons;
				~~~
				`),
			},

			want: &Doc{
				Code: "SELECT email, first_name, last_name FROM persons;",
				Lang: "sql",
				Info: "test",
			},
		},
		{
			name: "error: match more than two code fence",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.CodeFence("Code")
					return jig
				},
				src: TrimIndent(t, `
					# Root Heading

					~~~json
					{"hello": "world"}
					~~~

					~~~sql
					SELECT email, first_name, last_name FROM persons;
					~~~

					~~~yaml
					hello: world
					~~~
					`),
			},
			wantErr: "field 'Code' for code fence is already filled (inside 'Root Heading' section)",
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

func TestRepeatableCodeFence(t *testing.T) {
	type Query struct {
		Name string
		Code string
		Lang string
		Info string
	}

	type Doc struct {
		Queries []Query
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
			name: "single code fence: store lang and info",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					samples := jig.Root().Child(".", "Samples").Children("Queries")
					samples.Label("Name")
					samples.CodeFence("Code").Language("Lang").Info("Info")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Samples

				### Query Name and Email

				~~~sql :info
				SELECT email, first_name, last_name FROM persons;
				~~~

				### Query Name

				~~~sql :info2
				SELECT first_name, last_name FROM persons;
				~~~
				`),
			},

			want: &Doc{
				Queries: []Query{
					{
						Name: "Query Name and Email",
						Code: `SELECT email, first_name, last_name FROM persons;`,
						Lang: "sql",
						Info: "info",
					},
					{
						Name: "Query Name",
						Code: `SELECT first_name, last_name FROM persons;`,
						Lang: "sql",
						Info: "info2",
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

func TestCodeFence_Generate(t *testing.T) {
	type Doc struct {
		Name string
		Code string
		Lang string
		Info string
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
			name: "code fence",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.CodeFence("Code", "sql").Info("Info").Language("Lang").SampleCode("select * from users;").SampleInfo(":test")
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Name]

				~~~sql:test
				select * from users;
				~~~
				`, "~~~", "```"),
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

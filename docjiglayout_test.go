package mdd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayout_ChildIsValue(t *testing.T) {
	type Level6 struct {
		Name string
	}

	type Level5 struct {
		Name   string
		Level6 Level6
	}

	type Level4 struct {
		Name   string
		Level5 Level5
	}

	type Level3 struct {
		Name   string
		Level4 Level4
	}

	type Level2 struct {
		Name   string
		Level3 Level3
	}

	type Doc struct {
		Name   string
		NameP  *string
		Level2 Level2
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
			name: "root heading",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					return jig
				},
				src: TrimIndent(t, `
						# Root Heading
						`),
			},
			want: &Doc{
				Name: "Root Heading",
			},
		},
		{
			name: "root heading (field is pointer)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("NameP")
					return jig
				},
				src: TrimIndent(t, `
						# Root Heading
						`),
			},
			want: &Doc{
				NameP: &[]string{"Root Heading"}[0],
			},
		},
		{
			name: "root heading: invalid field name",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("InvalidName")
					return jig
				},
				src: TrimIndent(t, `
						# Root Heading
						`),
			},
			wantErr: "mdd.Doc doesn't have field 'InvalidName' for heading title (inside 'Root Heading' section)",
		},
		{
			name: "nested heading: don't assign header label if no label",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading")
					return jig
				},
				src: TrimIndent(t, `
						# Root Heading

						## Level2 Heading: Child Heading
						`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: Level2{
					Name: "",
				},
			},
		},
		{
			name: "nested heading: store label",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading").Label("Name")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading
				`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: Level2{
					Name: "Child Heading",
				},
			},
		},
		{
			name: "nested heading: invalid field",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading").Label("InvalidName")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading
				`),
			},
			wantErr: "mdd.Level2 doesn't have field 'InvalidName' for heading title (inside 'Level2 Heading: Child Heading' section)",
		},
		{
			name: "nested heading: max levels",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					level2 := root.Child("Level2", "Level2 Heading").Label("Name")
					level3 := level2.Child("Level3", "Level3 Heading").Label("Name")
					level4 := level3.Child("Level4", "Level4 Heading").Label("Name")
					level5 := level4.Child("Level5", "Level5 Heading").Label("Name")
					level5.Child("Level6", "Level6 Heading").Label("Name")

					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading

				### Level3 Heading: Child Heading

				#### Level4 Heading: Child Heading

				##### Level5 Heading: Child Heading

				###### Level6 Heading: Child Heading
				`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: Level2{
					Name: "Child Heading",
					Level3: Level3{
						Name: "Child Heading",
						Level4: Level4{
							Name: "Child Heading",
							Level5: Level5{
								Name: "Child Heading",
								Level6: Level6{
									Name: "Child Heading",
								},
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

func TestLayout_ChildIsPointer(t *testing.T) {
	type Level6 struct {
		Name string
	}

	type Level5 struct {
		Name   string
		Level6 *Level6
	}

	type Level4 struct {
		Name   string
		Level5 *Level5
	}

	type Level3 struct {
		Name   string
		Level4 *Level4
	}

	type Level2 struct {
		Name   string
		Level3 *Level3
	}

	type Doc struct {
		Name   string
		Level2 *Level2
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
			name: "nested heading: don't assign header label if no label",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading") /* Label("Name") */
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading
				`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: &Level2{
					Name: "",
				},
			},
		},
		{
			name: "nested heading: store label",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading").Label("Name")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading
				`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: &Level2{
					Name: "Child Heading",
				},
			},
		},
		{
			name: "nested heading: invalid field",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level2 Heading").Label("InvalidName")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading
				`),
			},
			wantErr: "mdd.Level2 doesn't have field 'InvalidName' for heading title (inside 'Level2 Heading: Child Heading' section)",
		},
		{
			name: "nested heading: max levels",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					level2 := root.Child("Level2", "Level2 Heading").Label("Name")
					level3 := level2.Child("Level3", "Level3 Heading").Label("Name")
					level4 := level3.Child("Level4", "Level4 Heading").Label("Name")
					level5 := level4.Child("Level5", "Level5 Heading").Label("Name")
					level5.Child("Level6", "Level6 Heading").Label("Name")

					return jig
				},
				src: TrimIndent(t, `
				# Root Heading

				## Level2 Heading: Child Heading

				### Level3 Heading: Child Heading

				#### Level4 Heading: Child Heading

				##### Level5 Heading: Child Heading

				###### Level6 Heading: Child Heading
				`),
			},
			want: &Doc{
				Name: "Root Heading",
				Level2: &Level2{
					Name: "Child Heading",
					Level3: &Level3{
						Name: "Child Heading",
						Level4: &Level4{
							Name: "Child Heading",
							Level5: &Level5{
								Name: "Child Heading",
								Level6: &Level6{
									Name: "Child Heading",
								},
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

func TestLayout_HeadingOption(t *testing.T) {

	type Doc struct {
		Name      string
		IntOpt    int
		BoolOpt   bool
		StringOpt string
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
			name: "Options",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Option("IntOpt")
					root.Option("BoolOpt")
					root.Option("StringOpt")
					return jig
				},
				src: TrimIndent(t, `
				# Root Heading (IntOpt=100, BoolOpt, StringOpt=test)
				`),
			},
			want: &Doc{
				Name:      "Root Heading",
				IntOpt:    100,
				BoolOpt:   true,
				StringOpt: "test",
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

func TestLayout_Generate(t *testing.T) {
	type Level2 struct {
		Name      string
		StringOpt string
		BoolOpt   bool
	}

	type Doc struct {
		Name   string
		Level2 Level2
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
			name: "only header (default)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2").Label("Name")
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Title]
				
				## [Level2]
				`),
		},
		{
			name: "no label header with pattern",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name")
					root.Child("Level2", "Level 2")
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Title]
				
				## Level 2
				`),
		},
		{
			name: "only header (sample)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name").Sample("Doc Title")
					root.Child("Level2", "Level2").Sample("Child Title") // ignored if Label() is not called
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Doc Title]
				
				## Level2
				`),
		},
		{
			name: "only header (sample content)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					root.Label("Name").Sample("Doc Title")
					root.SampleContent("This is sample content.")
					level2 := root.Child("Level2", "Level2").Label("Name").Sample("Child Title")
					level2.SampleContent("This is sample content in child layout.")
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Doc Title]

				This is sample content.
				
				## Level2: [Child Title]

				This is sample content in child layout.
				`),
		},
		{
			name: "only header (option)",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					root := jig.Root()
					label := root.Label("Name").Sample("Doc Title")
					label.Option("StringOpt").Sample("Test")
					label.Option("BoolOpt").Sample(true)
					return jig
				},
			},
			want: TrimIndent(t, `
				# [Doc Title] (StringOpt=[Test], BoolOpt)
				`),
		},
		{
			name: "Translation",
			args: args{
				create: func(t *testing.T) *DocJig[Doc] {
					jig := NewDocJig[Doc]()
					jig.Alias("Doc Title").Lang("ja", "ドキュメントタイトル")
					jig.Alias("Child Title").Lang("ja", "子タイトル")
					jig.Alias("Level2").Lang("ja", "レベル2")

					root := jig.Root()
					root.Label("Name").Sample("Doc Title")
					root.Child("Level2", "Level2").Label("Name").Sample("Child Title")
					return jig
				},
				opt: GenerateOption{
					Language: "ja",
				},
			},
			want: TrimIndent(t, `
				# [ドキュメントタイトル]

				## レベル2: [子タイトル]
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

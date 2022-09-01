package mdd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
	"github.com/russross/blackfriday/v2"
)

type alias struct {
	lang  string
	label string
}

// DocJig is a entry struct of this package.
//
// To create markdown parser, you should instantiate
// this struct by using [NewDocJig] function and specify
// contents with its and [Layout]'s methods ([Layout] is retreated by [DocJig.Root] method):
//
//	  package yourparser
//
//	  type LiterateBatch struct {
//		     Name string
//	      Code string
//	  }
//
//	  var jig = NewDocJig[LiterateBatch]()
//
//	  func init() {
//		     root := jig.Root()            // Top heading
//	      root.Name("Name")             // Assign heading title to Name field
//	      root.CodeFence("Code", "sh")  // Assign code fence content (type sh) to Code field
//	  }
//
//	  // Add your package's entry functions you like (XXXParseYYY funcs)
//	  func Parse(r io.Reader) (*LiterateBatch, error) {
//	      return jig.Parse(r)
//	  }
//
//	  func ParseString(s string) (*LiterateBatch, error) {
//	      return jig.ParseString(s)
//	  }
//
//	  func ParseFile(fp filepath) (*LiterateBatch, error) {
//	      return jig.ParseFile(fp)
//	  }
//
//	  func MustParse(r io.Reader) *LiterateBatch {
//	      return jig.Parse(r)
//	  }
//
//	  func MustParseString(s string) *LiterateBatch {
//	      return jig.ParseString(s)
//	  }
//
//	  func MustParseFile(fp filepath) *LiterateBatch {
//	      return jig.ParseFile(fp)
//	  }
type DocJig[T any] struct {
	root        *Layout[T]
	DefaultLang string
	aliases     map[string][]*alias
}

// NewDocJig is entry point function of this library
func NewDocJig[T any]() *DocJig[T] {
	j := &DocJig[T]{
		DefaultLang: "en",
		aliases:     make(map[string][]*alias),
	}

	j.root = &Layout[T]{
		j:     j,
		Level: 1,
	}
	return j
}

// Alias specify label alias or translation
//
// Basically Translation and standard alias is not different when parsing:
//
//	jig.Alias("Table", "表")
//	jig.Alias("Table").Lang("ja", "表")
//
// Language is used for [DocJig.GenerateTemplate].
func (j *DocJig[T]) Alias(primaryLabel string, aliases ...string) *Alias[T] {
	lowLabel := strings.ToLower(primaryLabel)
	if _, ok := j.aliases[lowLabel]; !ok {
		j.aliases[lowLabel] = append(j.aliases[lowLabel], &alias{
			lang:  j.DefaultLang,
			label: primaryLabel,
		})
	}
	for _, a := range aliases {
		j.aliases[lowLabel] = append(j.aliases[lowLabel], &alias{
			lang:  j.DefaultLang,
			label: a,
		})
	}
	return &Alias[T]{
		parent:       j,
		primaryLabel: lowLabel,
	}
}

func (j *DocJig[T]) matchLabel(pattern, actualLabel string) (suffix string, ok bool) {
	vl := strings.ToLower(actualLabel)
	pl := strings.ToLower(pattern)

	// alias is not registered
	if strings.HasPrefix(vl, pl) {
		actualLabel = actualLabel[len(pattern):]
		suffix = strings.TrimLeft(actualLabel, " :\t")
		ok = true
		return
	}

	for _, a := range j.aliases[pl] {
		if strings.HasPrefix(vl, strings.ToLower(a.label)) {
			actualLabel = actualLabel[len(a.label):]
			suffix = strings.TrimLeft(actualLabel, " :\t")
			ok = true
			return
		}
	}
	return actualLabel, false
}

func (j *DocJig[T]) findTranslation(pattern, lang string) string {
	if lang == "" {
		lang = j.DefaultLang
	}
	pl := strings.ToLower(pattern)
	for _, a := range j.aliases[pl] {
		if a.lang == lang {
			return a.label
		}
	}
	return pattern
}

func (j *DocJig[T]) translateToPrimaryKey(variant string, searchTargets []string) (index int, key string, ok bool) {
	for i, target := range searchTargets {
		for _, a := range j.aliases[target] {
			if variant == strings.ToLower(a.label) {
				index = i
				ok = true
				key = j.aliases[target][0].label
				return
			}
		}
	}
	return -1, "", false
}

// Alias is used to absorb orthographical variants or translation
//
// This object is created by [DocJig.Alias] method.
type Alias[T any] struct {
	parent       *DocJig[T]
	primaryLabel string
}

// Lang specifies word in other language
func (i *Alias[T]) Lang(lang string, aliases ...string) *Alias[T] {
	for _, a := range aliases {
		i.parent.aliases[i.primaryLabel] = append(i.parent.aliases[i.primaryLabel], &alias{
			lang:  lang,
			label: a,
		})
	}

	return i
}

// Root returns top [Layout] of markdown document
//
// [Layout] represents document block that has single heading and contents
// before same level heading.
//
// mmd-go only support markdown file that has single level 1 heading
func (j *DocJig[T]) Root(pattern ...string) *Layout[T] {
	return j.root
}

// GenerateOption is option to modify [DocJig.GenerateTemplate]'s result.
type GenerateOption struct {
	Language string
}

// GenerateTemplate generates markdown template
func (j *DocJig[T]) GenerateTemplate(w io.Writer, opt ...GenerateOption) error {
	var o GenerateOption
	if len(opt) > 0 {
		o = opt[0]
	}
	return j.root.generateTemplate(w, o.Language)
}

// Parse method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func Parse(r io.Reader) (*YourDocument, error) {
//	    return jig.Parse(r)
//	}
func (j *DocJig[T]) Parse(r io.Reader) (*T, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return j.ParseString(string(src))
}

// Parse method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func ParseString(r io.Reader) (*YourDocument, error) {
//	    return jig.ParseString(r)
//	}
func (j *DocJig[T]) ParseString(src string) (*T, error) {
	var result T

	rootResult := reflect.ValueOf(&result).Elem()

	// var current reflect.Value

	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions))
	root := parser.Parse([]byte(src))

	node := root.FirstChild
	stack := make([]*Layout[T], 7)
	stack[1] = j.root
	targets := make([]reflect.Value, 7)
	targets[1] = rootResult
	labels := make([]string, 7)
	currentLevel := 1

	for node != nil {
		switch node.Type {
		case blackfriday.Heading:
			var layout *Layout[T]
			var target reflect.Value
			label := plainTextRenderer(node.FirstChild)
			suffix := label
			currentLevel = node.Level
			ok := true
			labelMatched := false
			if currentLevel == 1 {
				layout = j.root
				target = rootResult
				suffix, labelMatched = j.matchLabel(layout.labelPattern, label)
			} else {
				parent := stack[currentLevel-1]
				var err error
				layout, target, suffix, ok, err = parent.findMatchedChild(label, targets[currentLevel-1])
				if ok {
					labelMatched = true
				}
				if err != nil {
					return nil, err
				}
			}
			if ok { // level1 or layer matched
				stack[currentLevel] = layout
				targets[currentLevel] = target
				noOptLabel, err := layout.processOption(suffix, target)
				if err != nil {
					return nil, err
				}
				if labelMatched {
					labels[currentLevel] = noOptLabel
					err := assignValue(target, layout.labelFieldName, noOptLabel, "heading title", label)
					if err != nil {
						return nil, err
					}
				}
			}
		case blackfriday.CodeBlock:
			layout := stack[currentLevel]
			target := targets[currentLevel]
			label := labels[currentLevel]

			lang, info := parseCodeBlockType(node.CodeBlockData.Info)

			cf, ok := layout.findMatchedCodeFence(lang)
			if ok {
				err := assignValue(target, cf.fieldName, strings.Trim(string(node.Literal), "\n"), "code fence", label)
				if err != nil {
					return nil, err
				}
				err = assignValue(target, cf.languageFieldName, lang, "code fence's lang", label)
				if err != nil {
					return nil, err
				}
				err = assignValue(target, cf.infoFieldName, info, "code fence's info", label)
				if err != nil {
					return nil, err
				}
			}
		case blackfriday.Table:
			layout := stack[currentLevel]
			target := targets[currentLevel]
			label := labels[currentLevel]
			cells, key2column := parseTable(node)
			if layout.table != nil {
				err := layout.table.assignCells(target, cells, key2column, label, &result)
				if err != nil {
					return nil, err
				}
			}
		}
		node = node.Next
	}

	return &result, nil
}

func assignValue(target reflect.Value, fieldName string, value any, context, label string) error {
	if fieldName == "" {
		return nil
	}
	field := getFieldByName(target, fieldName)

	if field.IsValid() {
		if !field.IsZero() {
			return fmt.Errorf("field '%s' for %s is already filled (inside '%s' section)", fieldName, context, label)
		}
		return runtimescan.FuzzyAssign(field.Addr().Interface(), value)
	} else {
		return fmt.Errorf("%s doesn't have field '%s' for %s (inside '%s' section)", target.Type(), fieldName, context, label)
	}
}

func getFieldByName(target reflect.Value, fieldName string) reflect.Value {
	if target.Kind() == reflect.Pointer {
		return target.Elem().FieldByName(fieldName)
	} else {
		return target.FieldByName(fieldName)
	}
}

// parseCodeBlockType parses code block first line
//
// It returns ("sql", ""):
//
//	```sql
//	```
//
// It returns ("csv", "users"):
//
//	```csv :users
//	```
func parseCodeBlockType(info []byte) (mode string, targetname string) {
	blocks := bytes.SplitN(info, []byte{':'}, 2)
	switch len(blocks) {
	case 0:
		return "", ""
	case 1:
		mode = strings.TrimSpace(string(blocks[0]))
	case 2:
		mode = strings.TrimSpace(string(blocks[0]))
		targetname = strings.TrimSpace(string(blocks[1]))
	}
	return
}

// Parse method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func ParseFile(filepath string) (*YourDocument, error) {
//	    return jig.ParseFile(filepath)
//	}
func (t *DocJig[T]) ParseFile(filepath string) (*T, error) {
	o, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer o.Close()
	return t.Parse(o)
}

// Parse method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func MustParse(r io.Reader) *YourDocument {
//	    return jig.MustParse(r)
//	}
func (t *DocJig[T]) MustParse(r io.Reader) *T {
	result, err := t.Parse(r)
	if err != nil {
		panic(err)
	}
	return result
}

// MustParseString method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func MustParseString(src string) *YourDocument {
//	    return jig.MustParseString(src)
//	}
func (t *DocJig[T]) MustParseString(src string) *T {
	result, err := t.ParseString(src)
	if err != nil {
		panic(err)
	}
	return result
}

// MustParseFile method is an entry point of your DocJig instance for
// your package user.
//
// You should wrap and create public function in you package like this:
//
//	func MustParseFile(filepath string) *YourDocument {
//	    return jig.MustParseFile(filepath)
//	}
func (t *DocJig[T]) MustParseFile(filepath string) *T {
	result, err := t.ParseFile(filepath)
	if err != nil {
		panic(err)
	}
	return result
}

// parseTable parses table and convert to slice of map
func parseTable(node *blackfriday.Node) (cells []map[string]string, key2column map[string]int) {
	headMode := false
	var column int
	var row int
	keyMap := make(map[int]string)
	key2column = make(map[string]int)

	var currentRow map[string]string

	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.TableCell:
			text := plainTextRenderer(node)
			if headMode {
				keyMap[column] = strings.ToLower(text)
				key2column[text] = column
			} else {
				if key, ok := keyMap[column]; ok {
					currentRow[key] = text
				}
			}
			column++
			return blackfriday.SkipChildren
		case blackfriday.TableHead:
			if entering {
				headMode = true
			} else {
				headMode = false
			}
		case blackfriday.TableRow:
			column = 0
			if entering {
				if headMode {
				} else {
					currentRow = make(map[string]string)
				}
			} else {
				row++
				if headMode {
					// todo error check
				} else {
					cells = append(cells, currentRow)
				}
			}
		}
		return blackfriday.GoToNext
	})
	return
}

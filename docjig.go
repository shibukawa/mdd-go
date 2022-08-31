package mdd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
	"github.com/russross/blackfriday/v2"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type alias struct {
	lang  string
	label string
}

type DocJig[T any] struct {
	root        *Layout[T]
	DefaultLang string
	aliases     map[string][]*alias
}

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

type Alias[T any] struct {
	parent       *DocJig[T]
	primaryLabel string
}

func (i *Alias[T]) Lang(lang string, aliases ...string) *Alias[T] {
	for _, a := range aliases {
		i.parent.aliases[i.primaryLabel] = append(i.parent.aliases[i.primaryLabel], &alias{
			lang:  lang,
			label: a,
		})
	}

	return i
}

func (j *DocJig[T]) Root(pattern ...string) *Layout[T] {
	return j.root
}

func (j *DocJig[T]) GenerateTemplate(w io.Writer) {
	// todo
}

func (j *DocJig[T]) Parse(r io.Reader) (*T, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return j.ParseString(string(src))
}

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

func (t *DocJig[T]) ParseFile(filepath string) (*T, error) {
	o, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer o.Close()
	return t.Parse(o)
}

func (t *DocJig[T]) MustParse(r io.Reader) *T {
	result, err := t.Parse(r)
	if err != nil {
		panic(err)
	}
	return result
}

func (t *DocJig[T]) MustParseString(src string) *T {
	result, err := t.ParseString(src)
	if err != nil {
		panic(err)
	}
	return result
}

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

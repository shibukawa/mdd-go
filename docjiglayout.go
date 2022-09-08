package mdd

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
)

// Layout[T] represents document structure
//
// Heading title makes layers:
//
//	# Root layer
//
//	## Child layer
//
//	## GrandChild layer
type Layout[T any] struct {
	j                 *DocJig[T]
	Level             int
	labelFieldName    string
	labelPattern      string
	samples           []string
	sampleContents    []string
	labelID           string
	instanceFieldName string
	children          []*Layout[T]
	codeFences        []*CodeFence[T]
	table             *Table[T]
	repeat            bool
	options           []*Option[T]
}

func (l *Layout[T]) Sample(sample string, samples ...string) *Layout[T] {
	l.samples = append([]string{sample}, samples...)
	return l
}

func (l *Layout[T]) SampleContent(sample string, samples ...string) *Layout[T] {
	l.sampleContents = append([]string{sample}, samples...)
	return l
}

func (l *Layout[T]) ID(labelID string) *Layout[T] {
	l.labelID = labelID
	return l
}

func (l *Layout[T]) Label(fieldName string, pattern ...string) *Layout[T] {
	l.labelFieldName = fieldName
	if len(pattern) > 0 {
		l.labelPattern = pattern[0]
	}
	return l
}

func (l *Layout[T]) Child(instanceFieldName string, pattern ...string) *Layout[T] {
	if l.Level == 6 {
		panic("Level should be under 7")
	}
	child := &Layout[T]{
		j:                 l.j,
		Level:             l.Level + 1,
		instanceFieldName: instanceFieldName,
	}
	l.children = append(l.children, child)
	if len(pattern) > 0 {
		child.labelPattern = pattern[0]
	}
	return child
}

func (l *Layout[T]) Children(instanceFieldName string, pattern ...string) *Layout[T] {
	children := l.Child(instanceFieldName, pattern...)
	children.repeat = true
	return children
}

func (l *Layout[T]) CodeFence(fieldName string, targetLanguages ...string) *CodeFence[T] {
	cf := &CodeFence[T]{
		fieldName:       fieldName,
		targetLanguages: targetLanguages,
		repeat:          l.repeat,
	}
	l.codeFences = append(l.codeFences, cf)
	return cf
}

func (l *Layout[T]) Table(fieldName string) *Table[T] {
	if l.table != nil {
		panic("this layout has table already")
	}
	l.table = &Table[T]{
		j:         l.j,
		fieldName: fieldName,
	}
	return l.table
}

func (l *Layout[T]) Option(fieldName string, pattern ...string) *Option[T] {
	result := &Option[T]{
		j:         l.j,
		l:         l,
		fieldName: fieldName,
	}
	if len(pattern) > 0 {
		result.pattern = pattern[0]
	} else {
		result.pattern = fieldName
	}
	l.options = append(l.options, result)
	return result
}

func (l *Layout[T]) findMatchedChild(label string, parentValue reflect.Value) (child *Layout[T], childValue reflect.Value, suffix string, ok bool, err error) {
	for _, c := range l.children {
		suffix, ok = l.j.matchLabel(c.labelPattern, label)
		if ok {
			if c.instanceFieldName == "." {
				childValue = parentValue
			} else {
				childValue = parentValue.FieldByName(c.instanceFieldName)
			}
			if !childValue.IsValid() {
				return nil, reflect.Value{}, "", false, fmt.Errorf("%s should have field %s but not", parentValue.Type(), c.instanceFieldName)
			}
			if !c.repeat {
				if childValue.Kind() == reflect.Pointer {
					if childValue.IsNil() {
						newInstance := reflect.New(childValue.Type().Elem())
						childValue.Set(newInstance)
					}
					childValue = childValue.Elem()
				}
			} else if c.instanceFieldName != "." {
				// add slice
				slice := childValue
				rowType := slice.Type().Elem() // todo: should support pointer type
				row := reflect.New(rowType).Elem()
				slice = reflect.Append(slice, row)
				childValue.Set(slice)

				childValue = slice.Index(slice.Len() - 1).Addr()
			}
			child = c
			return
		}
	}
	return nil, reflect.Value{}, "", false, nil
}

func (l *Layout[T]) findMatchedCodeFence(lang string) (cf *CodeFence[T], ok bool) {
	for _, c := range l.codeFences {
		if c.matchLanguage(lang) {
			return c, true
		}
	}
	return nil, false
}

var matchOpt = regexp.MustCompile(`(.*)\s*\((.*)\)`)

func (l *Layout[T]) processOption(label string, target reflect.Value) (string, error) {
	result := matchOpt.FindStringSubmatch(label)
	if len(result) < 2 {
		return label, nil
	}
	for _, opt := range strings.Split(result[2], ",") {
		opt := strings.TrimSpace(opt)
		var key string
		var value any
		var ok bool
		if key, value, ok = strings.Cut(opt, "="); ok {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value.(string))
		} else {
			key = opt
			value = true
		}
		for _, o := range l.options {
			if o.pattern == key {
				f := target.FieldByName(key)
				if !f.IsValid() {
					return "", fmt.Errorf("%s should have field %s but not", target.Type(), o.fieldName)
				}
				err := runtimescan.FuzzyAssign(f.Addr().Interface(), value)
				if err != nil {
					return "", err
				}
				break
			}
		}
	}
	return strings.TrimSpace(result[1]), nil
}

type Option[T any] struct {
	j         *DocJig[T]
	l         *Layout[T]
	fieldName string
	pattern   string
	sample    any
}

func (o *Option[T]) Sample(s any) {
	o.sample = s
}

func (l Layout[T]) generateTemplate(w io.Writer, lang string) error {
	length := 1
	if l.repeat {
		length = 2
	}

	i18n := func(src string) string {
		return l.j.findTranslation(src, lang)
	}

	for i := 0; i < length; i++ {
		fmt.Fprintf(w, "%s %s\n\n", strings.Repeat("#", l.Level), l.templateLabel(i, lang))

		if len(l.sampleContents) > 0 {
			var c string
			if i < len(l.sampleContents) {
				c = i18n(l.sampleContents[i])
			} else {
				c = i18n(l.sampleContents[len(l.sampleContents)-1])
			}
			fmt.Fprintf(w, "%s\n\n", c)
		}

		for _, cf := range l.codeFences {
			cf.generateTemplate(w)
		}

		if l.table != nil {
			l.table.generateTemplate(w, lang)
		}

		for _, c := range l.children {
			c.generateTemplate(w, lang)
		}
	}
	return nil
}

func (l Layout[T]) templateLabel(i int, lang string) string {
	var result string
	i18n := func(src string) string {
		return l.j.findTranslation(src, lang)
	}
	if l.labelPattern != "" {
		if l.labelFieldName != "" {
			if i < len(l.samples) {
				result = i18n(l.labelPattern) + ": [" + i18n(l.samples[i]) + "]"
			} else {
				result = i18n(l.labelPattern) + ": [" + i18n("Lorem Ipsum") + "]"
			}
		} else {
			result = i18n(l.labelPattern)
		}
	} else if len(l.samples) > 0 {
		if i < len(l.samples) {
			result = "[" + i18n(l.samples[i]) + "]"
		} else {
			result = "[...]"
		}
	} else if l.instanceFieldName != "" {
		if l.labelFieldName != "" {
			result = "[" + i18n(l.instanceFieldName) + "]"
		} else {
			result = i18n(l.instanceFieldName)
		}
	} else {
		result = "[" + i18n("Title") + "]"
	}
	var opts []string
	for _, o := range l.options {
		if o.sample != nil {
			if o.sample == true {
				opts = append(opts, i18n(o.fieldName))
			} else {
				opts = append(opts, fmt.Sprintf("%s=[%v]", i18n(o.fieldName), o.sample))
			}
		}
	}
	if len(opts) > 0 {
		result += " (" + strings.Join(opts, ", ") + ")"
	}

	return result
}

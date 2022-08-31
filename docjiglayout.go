package mdd

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
)

type Layout[T any] struct {
	j                 *DocJig[T]
	Level             int
	labelFieldName    string
	labelPattern      string
	instanceFieldName string
	children          []*Layout[T]
	codeFences        []*CodeFence[T]
	table             *Table[T]
	repeat            bool
	options           []*Option[T]
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
			} else {
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
}
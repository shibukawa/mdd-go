package mdd

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
	"github.com/shibukawa/formatdata-go"
)

type Table[T any] struct {
	j         *DocJig[T]
	fieldName string
	fields    []*StructField[T]
	asMap     bool
}

func (t *Table[T]) Field(fieldName string, key ...string) *StructField[T] {
	var k string
	var origK string
	if len(key) > 0 {
		k = strings.ToLower(key[0])
		origK = key[0]
	} else {
		k = strings.ToLower(fieldName)
		origK = fieldName
	}
	f := &StructField[T]{
		fieldName: fieldName,
		key:       k,
		origKey:   origK,
	}
	t.fields = append(t.fields, f)
	return f
}

func (t *Table[T]) AsMap() {
	t.asMap = true
}

func (t Table[T]) assignCells(target reflect.Value, cells []map[string]string, key2column map[string]int, label string, doc *T) error {
	if t.asMap {
		return t.assignCellsAsMap(target, cells, key2column, label, doc)
	} else {
		return t.assignCellsAsStruct(target, cells, key2column, label, doc)
	}
}

func (t Table[T]) assignCellsAsStruct(target reflect.Value, cells []map[string]string, key2column map[string]int, label string, doc *T) error {
	slice := target.FieldByName(t.fieldName)
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("field '%s' of %s is not slice type (inside '%s' section)", t.fieldName, target.Type(), label)
	}
	rowType := slice.Type().Elem() // todo: should support pointer type

	var fieldKeys []string
	for _, f := range t.fields {
		fieldKeys = append(fieldKeys, f.key)
	}

	keyMap := make([]string, len(t.fields))
	usedKeys := make(map[string]bool)
	for k := range key2column {
		index, k2, ok := t.j.translateToPrimaryKey(k, fieldKeys)
		if ok {
			keyMap[index] = k
			usedKeys[strings.ToLower(k2)] = true
		} else {
			lk := strings.ToLower(k)
			for i, f := range t.fields {
				if lk == f.key {
					keyMap[i] = lk
					usedKeys[lk] = true
					break
				}
			}
		}
	}

	var missingField []string
	for _, f := range t.fields {
		if f.required {
			if _, ok := usedKeys[f.key]; !ok {
				missingField = append(missingField, f.origKey)
			}
		}
	}

	if len(missingField) > 0 {
		return fmt.Errorf("required column(%s) are missing (inside '%s' section)", strings.Join(missingField, ", "), label)
	}

	for _, rv := range cells {
		row := reflect.New(rowType).Elem()
		for fi, f := range t.fields {
			var cv any = rv[keyMap[fi]]
			ct := row.FieldByName(f.fieldName)
			if !ct.IsValid() {
				return fmt.Errorf("%s doesn't have field '%s' (inside '%s' section)", rowType, f.fieldName, label)
			}
			if f.convert != nil {
				newV, err := f.convert(rv[keyMap[fi]], doc)
				if err != nil {
					return fmt.Errorf("can't convert value '%s' at field '%s' (inside '%s' section): %w", rv[keyMap[fi]], f.origKey, label, err)
				}
				runtimescan.FuzzyAssign(ct, newV)
			} else {
				runtimescan.FuzzyAssign(ct.Addr().Interface(), cv)
			}
		}
		slice = reflect.Append(slice, row)
	}
	sliceTarget := target.FieldByName(t.fieldName)
	sliceTarget.Set(slice)
	return nil
}

func (t Table[T]) assignCellsAsMap(target reflect.Value, cells []map[string]string, key2column map[string]int, label string, doc *T) error {
	var slice reflect.Value
	if target.Kind() == reflect.Pointer { // for repeat
		slice = target.Elem().FieldByName(t.fieldName)
	} else {
		slice = target.FieldByName(t.fieldName)
	}
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("field '%s' of %s is not slice type (inside '%s' section)", t.fieldName, target.Type(), label)
	}

	usedKeys := make([]string, len(key2column))
	origKeyMap := make([]string, len(key2column))
	for k, v := range key2column {
		usedKeys[v] = strings.ToLower(k)
		origKeyMap[v] = k
	}
	keyMap := make([]string, len(key2column))
	for i, shortK := range usedKeys {
		index, key, ok := t.j.translateToPrimaryKey(shortK, usedKeys)
		if ok {
			keyMap[index] = shortK
			origKeyMap[index] = key
		} else {
			keyMap[i] = shortK
		}
	}

	var result []map[string]string
	for _, rv := range cells {
		row := make(map[string]string)
		for ki, key := range keyMap {
			row[origKeyMap[ki]] = rv[key]
		}
		result = append(result, row)
	}
	slice = reflect.ValueOf(result)
	if target.Kind() == reflect.Pointer { // for repeat
		sliceTarget := target.Elem().FieldByName(t.fieldName)
		sliceTarget.Set(slice)
	} else {
		sliceTarget := target.FieldByName(t.fieldName)
		sliceTarget.Set(slice)
	}
	return nil
}

func (t Table[T]) generateTemplate(w io.Writer, lang string) {
	maxRows := 2
	headers := make([]any, len(t.fields))
	for i, f := range t.fields {
		headers[i] = t.j.findTranslation(f.fieldName, lang)
		if len(f.samples) > maxRows {
			maxRows = len(f.samples)
		}
	}
	cells := make([][]any, maxRows+1)
	cells[0] = headers
	for i := 0; i < maxRows; i++ {
		row := make([]any, len(t.fields))
		for j, f := range t.fields {
			if i < len(f.samples) {
				row[j] = f.samples[i]
			} else {
				row[j] = "..."
			}
		}
		cells[i+1] = row
	}
	formatdata.FormatDataTo(cells, w, formatdata.Opt{
		OutputFormat: formatdata.Markdown,
	})
	io.WriteString(w, "\n")
}

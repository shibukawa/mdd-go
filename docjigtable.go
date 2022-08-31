package mdd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/future-architect/tagscanner/runtimescan"
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

func (t *Table[T]) PreserveAllCells(fieldName string) *Table[T] {
	return t
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

	var usedKeys []string
	for _, f := range t.fields {
		usedKeys = append(usedKeys, f.key)
	}

	keyMap := make([]string, len(t.fields))
	for k := range key2column {
		index, _, ok := t.j.translateToPrimaryKey(k, usedKeys)
		if ok {
			keyMap[index] = k
		} else {
			lk := strings.ToLower(k)
			for i, f := range t.fields {
				if lk == f.key {
					keyMap[i] = lk
					break
				}
			}
		}
	}

	var missingField []string
	for _, f := range t.fields {
		if f.required {
			found := false
			for _, exist := range keyMap {
				if f.key == exist {
					found = true
					break
				}
			}
			if !found {
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
				vv := reflect.ValueOf(newV)
				if vv.Type() == ct.Type() && vv.Kind() == reflect.Struct && vv.Type() == ct.Type() {
					ct.Set(vv)
				} else if vv.Type() == ct.Type() && vv.Kind() == reflect.Pointer && vv.Type().Elem().Kind() == reflect.Struct {
					ct.Set(vv)
				} else {
					ctp := ct.Addr().Interface()
					runtimescan.FuzzyAssign(&ctp, &newV)
				}
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

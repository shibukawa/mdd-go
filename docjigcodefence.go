package mdd

import (
	"fmt"
	"io"
)

type CodeFence[T any] struct {
	fieldName         string
	targetLanguages   []string
	languageFieldName string
	infoFieldName     string
	repeat            bool
	sampleCode        string
	sampleInfo        string
}

func (cf *CodeFence[T]) Language(fieldName string) *CodeFence[T] {
	cf.languageFieldName = fieldName
	return cf
}

func (cf *CodeFence[T]) Info(fieldName string) *CodeFence[T] {
	cf.infoFieldName = fieldName
	return cf
}

func (cf *CodeFence[T]) SampleCode(code string) *CodeFence[T] {
	cf.sampleCode = code
	return cf
}

func (cf *CodeFence[T]) SampleInfo(info string) *CodeFence[T] {
	cf.sampleInfo = info
	return cf
}

func (cf CodeFence[T]) matchLanguage(lang string) bool {
	if len(cf.targetLanguages) == 0 {
		return true
	}
	for _, l := range cf.targetLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

func (cf CodeFence[T]) generateTemplate(w io.Writer) {
	var lang string
	if len(cf.targetLanguages) > 0 {
		lang = cf.targetLanguages[0]
	}
	var code string
	if cf.sampleCode != "" {
		code = cf.sampleCode + "\n"
	}
	fmt.Fprintf(w, "```%s%s\n%s```\n\n", lang, cf.sampleInfo, code)
}

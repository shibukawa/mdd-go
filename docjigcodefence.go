package mdd

type CodeFence[T any] struct {
	fieldName         string
	targetLanguages   []string
	languageFieldName string
	infoFieldName     string
	repeat            bool
}

func (cf *CodeFence[T]) Language(fieldName string) *CodeFence[T] {
	cf.languageFieldName = fieldName
	return cf
}

func (cf *CodeFence[T]) Info(fieldName string) *CodeFence[T] {
	cf.infoFieldName = fieldName
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

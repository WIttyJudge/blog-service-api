package validator

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	validator.Validate
	Trans ut.Translator
}

func New() *Validator {
	english := en.New()
	uni := ut.New(english, english)

	trans, _ := uni.GetTranslator("en")
	v := validator.New()

	en_translations.RegisterDefaultTranslations(v, trans)

	return &Validator{
		Validate: *v,
		Trans:    trans,
	}
}

func (v *Validator) ValidationErrorsToSlice(err error) []string {
	errors := err.(validator.ValidationErrors)
	slice := make([]string, 0, len(errors))

	for _, e := range errors {
		slice = append(slice, e.Translate(v.Trans))
	}

	return slice
}

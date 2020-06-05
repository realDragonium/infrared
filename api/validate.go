package api

import (
	"errors"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"reflect"
	"strings"
)

var translator ut.Translator

func formatValidationErrors(errs error) map[string]interface{} {
	fieldErrs := errs.(validator.ValidationErrors)
	validationErrMap := map[string]interface{}{}

	for _, fieldErr := range fieldErrs {
		namespace := strings.Split(fieldErr.Namespace(), ".")
		if len(namespace) > 1 {
			namespace = namespace[1:]
		}

		outer := validationErrMap

		for n, name := range namespace {
			if len(namespace)-1 == n {
				outer[name] = fieldErr.Translate(translator)
				break
			}

			inner, ok := outer[name].(map[string]interface{})
			if !ok {
				inner = map[string]interface{}{}
				outer[name] = inner
			}

			outer = inner
		}
	}

	return validationErrMap
}

func setupValidator() error {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("could not find the validator")
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	translator = uni.GetFallback()
	if err := entranslations.RegisterDefaultTranslations(validate, translator); err != nil {
		return err
	}

	return nil
}

package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errs []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   e.Field(),
			Message: msgForTag(e.Tag(), e.Param()),
		})
	}
	return errs
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short (min " + param + ")"
	case "max":
		return "Value is too long (max " + param + ")"
	case "oneof":
		return "Must be one of: " + param
	case "uuid":
		return "Must be a valid UUID"
	case "e164":
		return "Must be a valid phone number (e.g. +6281234567890)"
	}
	return "Invalid value"
}

package validator

import (
	"github.com/Lomank123/go-service-event/pkg/client/enum"
	"github.com/go-playground/validator/v10"
)

func ValidateEventType(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	for _, t := range enum.AllEventTypes {
		if v == string(t) {
			return true
		}
	}
	return false
}

func ValidateEventStatus(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	for _, s := range enum.AllEventStatuses {
		if v == string(s) {
			return true
		}
	}
	return false
}

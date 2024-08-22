/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package validation

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"

	"github.com/emicklei/go-restful"
)

type BadRequestError struct {
	Reason string `json:"reason"`
}

func IsValidWithinRange(validateType reflect.Type, valueToValidate int, fieldName string, resp *restful.Response) bool {
	field, found := validateType.FieldByName(fieldName)
	if found {
		minimum, _ := strconv.Atoi(field.Tag.Get("minimum"))
		maximum, _ := strconv.Atoi(field.Tag.Get("maximum"))
		if valueToValidate > maximum || valueToValidate < minimum {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, BadRequestError{
				Reason: fieldName + " should be in the range of " + field.Tag.Get("minimum") + " to " + field.Tag.Get("maximum"),
			})
			return false
		}
	}
	return true

}

func IsValidLength(validateType reflect.Type, valueToValidate string, fieldName string, resp *restful.Response) bool {
	field, found := validateType.FieldByName(fieldName)
	if found {
		maximum, _ := strconv.Atoi(field.Tag.Get("maximum"))
		if len(valueToValidate) > int(maximum) {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, BadRequestError{
				Reason: fieldName + " length should be less than " + field.Tag.Get("maximum"),
			})
			return false
		}
	}
	return true
}

func IsValidString(valueToValidate string, resp *restful.Response) bool {
	validRegex := regexp.MustCompile("[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*")
	if !validRegex.MatchString(valueToValidate) {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, BadRequestError{
			Reason: "Allowed characters: lowercase letters (a-z), numbers (0-9), and hyphens (-)",
		})
		return false
	}
	return true

}

/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package util

import (
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strconv"

	"github.com/emicklei/go-restful"
)

type BadRequestError struct {
	Reason string `json:"reason"`
}

func IsValidLength(validateType reflect.Type, valueToValidate string, fieldName string, resp *restful.Response) bool {
	field, found := validateType.FieldByName(fieldName)
	if found {
		maximum, _ := strconv.Atoi(field.Tag.Get("maximum"))
		if len(valueToValidate) > int(maximum) {
			resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
				Reason: fieldName + " length should be less than " + field.Tag.Get("maximum"),
			})
			return false
		}
	}
	return true
}

// Check if all label pairs are valid. If quantity of labels is reduced after being converted to map, there are duplicated keys.
func IsValidLabels(validateType reflect.Type, size int, labels map[string]string, resp *restful.Response) bool {
	if size > len(labels) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: validateType.Name() + " Key should be unique",
		})
		return false
	} else {
		for key, value := range labels { // Empty array is also valid.
			if !IsValidLabelEntry(validateType, key, value, resp) { // check validness
				return false
			}
		}
	}
	return true
}

func IsValidLabelEntry(validateType reflect.Type, key string, value string, resp *restful.Response) bool {
	if !IsValidLength(validateType, key, "Key", resp) {
		return false
	}

	if !IsValidLabelKeyString(key, resp) {
		return false
	}

	if !IsValidLength(validateType, value, "Value", resp) {
		return false
	}

	if !IsValidLabelValueString(value, resp) {
		return false
	}

	return true
}

func IsValidLabelKeyString(valueToValidate string, resp *restful.Response) bool {
	validRegex := regexp.MustCompile("^[A-Za-z0-9-_./]+$")
	if !validRegex.MatchString(valueToValidate) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Invalid key. Valid characters: A-Z, a-z, 0-9, -(hyphen), _(underscore), .(dot), and /(slash)",
		})
		return false
	}
	return true
}

func IsValidLabelValueString(valueToValidate string, resp *restful.Response) bool {
	validRegex := regexp.MustCompile("^[A-Za-z0-9-_.]*$") // Also match "", so use '*'
	if !validRegex.MatchString(valueToValidate) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Invalid value. Valid characters: A-Z, a-z, 0-9, -(hyphen), _(underscore), and .(dot)",
		})
		return false
	}
	return true
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

func IsValidCIDR(cidr string, resp *restful.Response) bool {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, BadRequestError{
			Reason: "Invalid CIDR address: " + cidr,
		})
		return false
	}
	return true
}

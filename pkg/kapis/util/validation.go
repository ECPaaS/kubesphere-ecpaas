/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com
*/

package util

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

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
	validNameRegex := regexp.MustCompile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")

	parts := strings.Split(valueToValidate, "/")
	var name string
	
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		validPrefixRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
		var prefix string
		prefix, name = parts[0], parts[1]
		if len(prefix) > 253 {
			resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
				Reason: "Invalid key. The prefix in label key must be no more then 253 charactors",
			})
			return false
		} else if !validPrefixRegex.MatchString(prefix) {
			resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
				Reason: "Invalid key. The prefix in label key must consist of lower case alphanumeric characters,"+
				" '-' or '.', and must start and end with an alphanumeric character",
			})
			return false
		}
	default:
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Invalid key. A valid label key must consist of alphanumeric characters, '-', '_' or '.',"+
			" and must start and end with an alphanumeric character (e.g. 'MyName', or 'my.name', or '123-abc')"+
			" with an optional DNS subdomain prefix and '/' (e.g. 'example.com/MyName')",
		})
		return false
	}

	if !validNameRegex.MatchString(name) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Invalid key. A valid label key must consist of alphanumeric characters, '-', '_' or '.',"+
			" and must start and end with an alphanumeric character (e.g. 'MyName', or 'my.name', or '123-abc')",
		})
		return false
	}
	return true
}

func IsValidLabelValueString(valueToValidate string, resp *restful.Response) bool {
	validRegex := regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")
	if !validRegex.MatchString(valueToValidate) {
		resp.WriteHeaderAndEntity(http.StatusForbidden, BadRequestError{
			Reason: "Invalid value. A valid label value must be an empty string or consist of alphanumeric characters, '-', '_' or '.',"+
			" and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345')",
		})
		return false
	}
	return true
}

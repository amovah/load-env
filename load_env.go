package load_nev

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var tagSetParamCount = map[string]int{
	"required": 1,
	"name":     2,
	"min":      2,
	"max":      2,
	"default":  2,
}

var numberTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
}

var allowTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
	reflect.String,
	reflect.Bool,
}

func contains(slice []reflect.Kind, item reflect.Kind) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type envTag struct {
	name         string
	required     bool
	min          float64
	max          float64
	defaultValue string
	hasMap       map[string]bool
}

// LoadEnv loads environment variables into a struct.
// The struct must have a tag of "env" and the tag must be in the format of:
//  env:"NAME"
// 	env:"NAME,required"
// 	env:"NAME,required,name=NAME_VAR"
func LoadEnv(target interface{}) error {
	reflectOf := reflect.TypeOf(target)
	if reflectOf.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	if reflectOf.Elem().Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	for i := 0; i < reflectOf.Elem().NumField(); i++ {
		field := reflectOf.Elem().Field(i)
		if !contains(allowTypes, field.Type.Kind()) {
			return errors.New(fmt.Sprintf("field %s with type %s is not allowed", field.Name, field.Type.Kind()))
		}

		options := field.Tag.Get("env")
		tag, err := extractOptionsFromString(options, field)
		if err != nil {
			return err
		}

		valueFromEnv := os.Getenv(tag.name)
		if tag.required && valueFromEnv == "" {
			return errors.New(fmt.Sprintf("env %s is required", tag.name))
		}

		if valueFromEnv == "" {
			valueFromEnv = tag.defaultValue
		}

		err = setValueToStructFromEnvTag(target, field, valueFromEnv, tag, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func setValueToStructFromEnvTag(target interface{}, field reflect.StructField, valueFromEnv string, tag envTag, fieldNumber int) error {
	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(valueFromEnv, 10, 64)
		if err != nil {
			return errors.New(fmt.Sprintf("env %s must be a number", tag.name))
		}

		if tag.min != 0 && parsed < int64(tag.min) {
			return errors.New(fmt.Sprintf("env %s must be greater than %f", tag.name, tag.min))
		}
		if tag.max != 0 && parsed > int64(tag.max) {
			return errors.New(fmt.Sprintf("env %s must be less than %f", tag.name, tag.max))
		}

		reflect.ValueOf(target).Elem().Field(fieldNumber).SetInt(parsed)
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(valueFromEnv, 10, 64)
		if err != nil {
			return errors.New(fmt.Sprintf("env %s must be a number", tag.name))
		}

		if tag.min != 0 && parsed < uint64(tag.min) {
			return errors.New(fmt.Sprintf("env %s must be greater than %f", tag.name, tag.min))
		}

		if tag.max != 0 && parsed > uint64(tag.max) {
			return errors.New(fmt.Sprintf("env %s must be less than %f", tag.name, tag.max))
		}

		reflect.ValueOf(target).Elem().Field(fieldNumber).SetUint(parsed)
		break
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(valueFromEnv, 10)
		if err != nil {
			return errors.New(fmt.Sprintf("env %s must be a number", tag.name))
		}
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetFloat(parsed)
		break
	case reflect.String:
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetString(valueFromEnv)
		break
	case reflect.Bool:
		parsed, err := strconv.ParseBool(valueFromEnv)
		if err != nil {
			return errors.New(fmt.Sprintf("env %s must be a boolean", tag.name))
		}
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetBool(parsed)
		break
	}

	return nil
}

func extractOptionsFromString(options string, field reflect.StructField) (envTag, error) {
	var tag = envTag{hasMap: map[string]bool{}}

	for _, option := range strings.Split(options, ",") {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}

		splittedOption := strings.Split(option, "=")
		optionKey := splittedOption[0]
		var optionValue string
		if tagSetParamCount[optionKey] == 2 {
			if len(splittedOption) < 2 {
				return tag, errors.New(fmt.Sprintf("option %s must has a parameter", optionKey))
			}
			optionValue = splittedOption[1]
		}

		if optionKey == "name" && optionValue == "" {
			return tag, errors.New("option name must has a parameter")
		}

		if optionKey == "min" || optionKey == "max" {
			if !contains(numberTypes, field.Type.Kind()) {
				return tag, errors.New(fmt.Sprintf("option %s must be used with a number type", optionKey))
			}
		}

		switch optionKey {
		case "required":
			tag.required = true
		case "name":
			tag.name = optionValue
		case "min":
			parsed, err := strconv.ParseFloat(optionValue, 64)
			if err != nil {
				return tag, errors.New(fmt.Sprintf("option %s must be a number", optionKey))
			}
			tag.min = parsed
		case "max":
			parsed, err := strconv.ParseFloat(optionValue, 64)
			if err != nil {
				return tag, errors.New(fmt.Sprintf("option %s must be a number", optionKey))
			}
			tag.max = parsed
		case "default":
			tag.defaultValue = optionValue
		}
		tag.hasMap[optionKey] = true
	}

	return tag, nil
}

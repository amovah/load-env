package load_env

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
	reflect.Struct,
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
	defaultValue string
	hasMap       map[string]bool
}

// LoadEnv loads environment variables into a struct.
// The struct must have a tag of "env" and the tag must be in the format of:
// env:"name=NAME"
// env:"name=NAME,required"
// env:"name=NAME,default=DEFAULT"
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
			return fmt.Errorf("field %s with type %s is not allowed", field.Name, field.Type.Kind())
		}

		if field.Type.Kind() == reflect.Struct {
			innerStruct := reflect.New(field.Type)
			err := LoadEnv(innerStruct.Interface())
			if err != nil {
				return err
			}

			reflect.ValueOf(target).Elem().Field(i).Set(innerStruct.Elem())
			continue
		}

		options := field.Tag.Get("env")
		tag, err := extractOptionsFromString(options, field)
		if err != nil {
			return err
		}

		valueFromEnv := os.Getenv(tag.name)
		if tag.required && valueFromEnv == "" {
			return fmt.Errorf("env %s is required", tag.name)
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
			return fmt.Errorf("env %s must be a number", tag.name)
		}

		reflect.ValueOf(target).Elem().Field(fieldNumber).SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(valueFromEnv, 10, 64)
		if err != nil {
			return fmt.Errorf("env %s must be a number", tag.name)
		}

		reflect.ValueOf(target).Elem().Field(fieldNumber).SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(valueFromEnv, 64)
		if err != nil {
			return fmt.Errorf("env %s must be a number", tag.name)
		}
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetFloat(parsed)
	case reflect.String:
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetString(valueFromEnv)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(valueFromEnv)
		if err != nil {
			return fmt.Errorf("env %s must be a boolean", tag.name)
		}
		reflect.ValueOf(target).Elem().Field(fieldNumber).SetBool(parsed)
	default:
		return fmt.Errorf("unsupported kind: %s", field.Type.Kind())
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
				return tag, fmt.Errorf("option %s must has a parameter", optionKey)
			}
			optionValue = splittedOption[1]
		}

		if optionKey == "name" && optionValue == "" {
			return tag, errors.New("option name must has a parameter")
		}

		switch optionKey {
		case "required":
			tag.required = true
		case "name":
			tag.name = optionValue
		case "default":
			tag.defaultValue = optionValue
		}
		tag.hasMap[optionKey] = true
	}

	return tag, nil
}

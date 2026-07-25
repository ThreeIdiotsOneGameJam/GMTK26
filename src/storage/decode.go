package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// decode unmarshals storage JSON into destination.
//
// On a JSON type mismatch (e.g. a field's type changed between versions), it
// recovers by loading compatible fields and leaving incompatible ones at their
// zero value. repaired is true when that fallback path was used.
func decode(data []byte, destination any) (repaired bool, err error) {
	if err := json.Unmarshal(data, destination); err == nil {
		return false, nil
	} else if !isJSONTypeError(err) {
		return false, fmt.Errorf("%w: %v", ErrInvalidData, err)
	}

	if err := decodeStructFields(data, destination); err != nil {
		return false, err
	}
	return true, nil
}

func isJSONTypeError(err error) bool {
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &typeErr)
}

func decodeStructFields(data []byte, destination any) error {
	rv := reflect.ValueOf(destination)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%w: destination must be a non-nil pointer", ErrInvalidData)
	}

	structValue, ok := structValueOf(rv.Elem())
	if !ok {
		return fmt.Errorf("%w: cannot recover non-struct value after type error", ErrInvalidData)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidData, err)
	}

	// Failed Unmarshal may have left destination partially filled.
	structValue.Set(reflect.Zero(structValue.Type()))

	fields := exportedJSONFields(structValue)
	for key, rawValue := range raw {
		field, ok := fields[key]
		if !ok {
			continue
		}
		if err := json.Unmarshal(rawValue, field.Addr().Interface()); err != nil {
			continue
		}
	}

	return nil
}

func structValueOf(v reflect.Value) (reflect.Value, bool) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}, false
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct || !v.CanSet() {
		return reflect.Value{}, false
	}
	return v, true
}

func exportedJSONFields(structValue reflect.Value) map[string]reflect.Value {
	structType := structValue.Type()
	fields := make(map[string]reflect.Value, structType.NumField())

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		if fieldType.PkgPath != "" { // unexported
			continue
		}

		name, ok := jsonFieldName(fieldType)
		if !ok {
			continue
		}

		fieldValue := structValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}
		fields[name] = fieldValue
	}

	return fields
}

func jsonFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name, true
	}
	return name, true
}

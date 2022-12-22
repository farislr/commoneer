package internal

import (
	"reflect"
	"strings"
)

func ModifyOrKeepField(existingQuery string, model interface{}) (query string) {
	if !strings.Contains(existingQuery, "*") {
		return existingQuery
	}

	fields := GetColumns(model)

	query = strings.ReplaceAll(existingQuery, "*", strings.Join(fields, ", "))

	return query
}

func GetColumns(model interface{}) []string {
	value := reflect.Indirect(reflect.ValueOf(model))

	if value.Kind() == reflect.Slice {
		value = reflect.Indirect(reflect.New(value.Type().Elem()))
	}

	rType := value.Type()
	var fields []string
	for i := 0; i < value.NumField(); i++ {
		if v, ok := rType.Field(i).Tag.Lookup("column"); ok {
			fields = append(fields, v)
		}
	}

	return fields
}

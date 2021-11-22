package shared

import (
	"reflect"
	"regexp"
)

func KeyOf(v interface{}) (key string) {
	t := reflect.TypeOf(v)

	switch t.Kind() {
	case reflect.Ptr:
		key = t.Elem().Name()
	case reflect.Struct:
		key = t.Name()
	case reflect.String:
		key = v.(string)
	default:
		key = "Unknown"
	}

	key = regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(key, "")

	return
}


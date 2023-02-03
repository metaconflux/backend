// Original: https://gist.github.com/hvoecking/10772475
package template

import (
	"fmt"
	"reflect"

	"github.com/metaconflux/backend/internal/utils"
)

func Template(in interface{}, out interface{}, params map[string]interface{}) error {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(in)

	copy := reflect.New(original.Type()).Elem()
	err := templateRecursive(copy, original, nil, params)
	if err != nil {
		return err
	}

	reflect.ValueOf(out).Elem().Set(copy.Elem())

	// Remove the reflection wrapper
	return nil
}

func templateRecursive(copy, original reflect.Value, tags *reflect.StructTag, params map[string]interface{}) error {
	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return fmt.Errorf("Nil pointer")
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		templateRecursive(copy.Elem(), originalValue, nil, params)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		templateRecursive(copyValue, originalValue, nil, params)
		copy.Set(copyValue)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			tagTmp := original.Type().Field(i).Tag
			templateRecursive(copy.Field(i), original.Field(i), &tagTmp, params)
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			templateRecursive(copy.Index(i), original.Index(i), nil, params)
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			templateRecursive(copyValue, originalValue, nil, params)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string translate it (yay finally we're doing what we came for)
	case reflect.String:
		value := original.Interface().(string)
		if tags != nil {
			tagVal, ok := tags.Lookup("template")
			if ok && tagVal != "false" {
				var err error
				value, err = utils.Template(value, params)
				if err != nil {
					return err
				}
			}
		}
		copy.SetString(value)

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}

	return nil

}

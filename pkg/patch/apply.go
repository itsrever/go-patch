package patch

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
)

// Apply updates the target struct in-place with non-zero values from the patch struct.
// Only fields with the same name and type get updated. Fields in the patch struct can be
// pointers to the target's type.
//
// Returns true if any value has been changed.
func Apply(target interface{}, patch map[string]interface{}) (changed bool, err error) {
	var dst = structs.New(target)

	for key, value := range patch {
		var name = key
		var dstField, ok = findField(dst, name)
		if !ok {
			continue // skip non-existing fields
		}
		dstKind := dstField.Kind()
		dstValue := dstField.Value()
		srcValue := reflect.ValueOf(value)
		srcValue = reflect.Indirect(srcValue)

		// recursive for structs and pointers to structs
		if dstKind == reflect.Struct ||
			(dstKind == reflect.Pointer &&
				reflect.Indirect(reflect.ValueOf(dstValue)).Kind() == reflect.Struct) {
			valueAsStruct, ok := value.(map[string]interface{})
			if !ok {
				err = fmt.Errorf("%v is not a struct", name)
				return
			}
			iChanged, iErr := Apply(dstValue, valueAsStruct)
			if iErr != nil {
				err = iErr
				return
			}
			changed = changed || iChanged
			continue
		}

		if !reflect.DeepEqual(value, dstValue) {
			changed = true
		}

		// handling of setting arrays/slices
		if dstKind == reflect.Slice {
			dstElemType := reflect.TypeOf(dstValue).Elem()
			castedArray := reflect.MakeSlice(reflect.TypeOf(dstValue), srcValue.Len(), srcValue.Len())
			valueAsArray, ok := value.([]interface{})
			if !ok {
				err = fmt.Errorf("%v is not an array", name)
				return
			}
			for i, srcElemValue := range valueAsArray {
				reflectSrcElemValue := reflect.ValueOf(srcElemValue)
				if !reflectSrcElemValue.CanConvert(dstElemType) {
					err = fmt.Errorf("can't convert %v to dst type", name)
					break
				}
				castedArray.Index(i).Set(reflectSrcElemValue.Convert(dstElemType))
			}
			if err != nil {
				return
			}
			err = dstField.Set(castedArray.Interface())
			if err != nil {
				return
			}
			continue

		}

		// other values
		if !srcValue.CanConvert(reflect.TypeOf(dstValue)) {
			err = fmt.Errorf("can't convert %v to dst type", name)
			return
		}

		srcValue = srcValue.Convert(reflect.TypeOf(dstField.Value()))
		err = dstField.Set(srcValue.Interface())
		if err != nil {
			return
		}

	}
	return
}

func findField(dst *structs.Struct, name string) (*structs.Field, bool) {
	for _, field := range dst.Fields() {
		tag := field.Tag("json")
		if tag == "" {
			tag = field.Name()
		} else {
			tag, _, _ = strings.Cut(tag, ",")
		}
		if tag == name {
			if field.IsExported() {
				return field, true
			}
		}
	}
	return nil, false
}

package utils

import (
	"database/sql"
	"errors"
	"reflect"
)

// Copy copy things
func Copy(toValue interface{}, fromValue interface{}) (err error) {
	var (
		isSlice bool
		amount  = 1
		from    = Indirect(reflect.ValueOf(fromValue))
		to      = Indirect(reflect.ValueOf(toValue))
	)

	if !to.CanAddr() {
		return errors.New("copy to value is unaddressable")
	}

	// Return is from value is invalid
	if !from.IsValid() {
		return
	}

	fromType := IndirectType(from.Type())
	toType := IndirectType(to.Type())

	// Just set it if possible to assign
	// And need to do copy anyway if the type is struct
	done := cleanSet(fromType, from, to, toType)
	if done {
		return
	}

	isSlice, amount = sliceInfo(to, isSlice, from, amount)

	for i := 0; i < amount; i++ {
		var dest, source reflect.Value

		if isSlice {
			// source
			source = setSource(from, source, i)
			// dest
			dest = Indirect(reflect.New(toType).Elem())
		} else {
			source = Indirect(from)
			dest = Indirect(to)
		}

		// check source
		if source.IsValid() {
			fromTypeFields := deepFields(fromType)
			//fmt.Printf("%#v", fromTypeFields)
			// Copy from field to field or method
			err2 := copyFieldToFieldOrMethod(fromTypeFields, source, dest)
			if err2 != nil {
				return err2
			}

			// Copy from method to field
			copyMethodToField(toType, source, dest)
		}
		sliceWorker(isSlice, dest, to)
	}
	return
}

func setSource(from reflect.Value, source reflect.Value, i int) reflect.Value {
	if from.Kind() == reflect.Slice {
		source = Indirect(from.Index(i))
	} else {
		source = Indirect(from)
	}
	return source
}

func sliceInfo(to reflect.Value, isSlice bool, from reflect.Value, amount int) (bool, int) {
	if to.Kind() == reflect.Slice {
		isSlice = true
		if from.Kind() == reflect.Slice {
			amount = from.Len()
		}
	}
	return isSlice, amount
}

func cleanSet(fromType reflect.Type, from reflect.Value, to reflect.Value, toType reflect.Type) bool {
	if fromType.Kind() != reflect.Struct && from.Type().AssignableTo(to.Type()) {
		to.Set(from)
		return true
	}

	if fromType.Kind() != reflect.Struct || toType.Kind() != reflect.Struct {
		return true
	}
	return false
}

func copyFieldToFieldOrMethod(fromTypeFields []reflect.StructField, source reflect.Value, dest reflect.Value) error {
	for _, field := range fromTypeFields {
		name := field.Name

		if fromField := source.FieldByName(name); fromField.IsValid() {
			// has field
			if toField := dest.FieldByName(name); toField.IsValid() {
				err2 := setWithField(toField, fromField)
				if err2 != nil {
					return err2
				}
			} else {
				// try to set to method
				setToMethod(dest, name, fromField)
			}
		}
	}
	return nil
}

func setWithField(toField reflect.Value, fromField reflect.Value) error {
	if toField.CanSet() {
		if !set(toField, fromField) {
			if err := Copy(toField.Addr().Interface(), fromField.Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func setToMethod(dest reflect.Value, name string, fromField reflect.Value) {
	var toMethod reflect.Value
	if dest.CanAddr() {
		toMethod = dest.Addr().MethodByName(name)
	} else {
		toMethod = dest.MethodByName(name)
	}

	if toMethod.IsValid() && toMethod.Type().NumIn() == 1 && fromField.Type().AssignableTo(toMethod.Type().In(0)) {
		toMethod.Call([]reflect.Value{fromField})
	}
}

func copyMethodToField(toType reflect.Type, source reflect.Value, dest reflect.Value) {
	for _, field := range deepFields(toType) {
		name := field.Name

		var fromMethod reflect.Value
		if source.CanAddr() {
			fromMethod = source.Addr().MethodByName(name)
		} else {
			fromMethod = source.MethodByName(name)
		}

		if fromMethod.IsValid() && fromMethod.Type().NumIn() == 0 && fromMethod.Type().NumOut() == 1 {
			if toField := dest.FieldByName(name); toField.IsValid() && toField.CanSet() {
				values := fromMethod.Call([]reflect.Value{})
				if len(values) >= 1 {
					set(toField, values[0])
				}
			}
		}
	}
}

func sliceWorker(isSlice bool, dest reflect.Value, to reflect.Value) {
	if isSlice {
		if dest.Addr().Type().AssignableTo(to.Type().Elem()) {
			to.Set(reflect.Append(to, dest.Addr()))
		} else if dest.Type().AssignableTo(to.Type().Elem()) {
			to.Set(reflect.Append(to, dest))
		}
	}
}

func deepFields(reflectType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	if reflectType = IndirectType(reflectType); reflectType.Kind() == reflect.Struct {
		for i := 0; i < reflectType.NumField(); i++ {
			v := reflectType.Field(i)
			if v.Anonymous {
				fields = append(fields, deepFields(v.Type)...)
			} else {
				fields = append(fields, v)
			}
		}
	}

	return fields
}

func set(to, from reflect.Value) bool {
	if from.IsValid() {
		if to.Kind() == reflect.Ptr {
			//set `to` to nil if from is nil
			if from.Kind() == reflect.Ptr && from.IsNil() {
				to.Set(reflect.Zero(to.Type()))
				return true
			} else if to.IsNil() {
				to.Set(reflect.New(to.Type().Elem()))
			}
			to = to.Elem()
		}

		b, done := setStep2(to, from)
		if done {
			return b
		}
	}
	return true
}

func setStep2(to reflect.Value, from reflect.Value) (bool, bool) {
	if from.Type().ConvertibleTo(to.Type()) {
		to.Set(from.Convert(to.Type()))
	} else if scanner, ok := to.Addr().Interface().(sql.Scanner); ok {
		err := scanner.Scan(from.Interface())
		if err != nil {
			return false, true
		}
	} else if from.Kind() == reflect.Ptr {
		return set(to, from.Elem()), true
	} else {
		return false, true
	}
	return false, false
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type From struct {
	ID      string
	Balance float64
}

type To struct {
	ID      int
	Balance string
}

func mapStructFields(ctx context.Context, from interface{}, to interface{}) {
	fromVal := reflect.ValueOf(from).Elem()
	toVal := reflect.ValueOf(to).Elem()
	fromType := fromVal.Type()

	for i := 0; i < fromVal.NumField(); i++ {
		fieldName := fromType.Field(i).Name
		fromField := fromVal.FieldByName(fieldName)
		toField := toVal.FieldByName(fieldName)

		if toField.IsValid() && toField.CanSet() {
			mapField(ctx, fromField, toField)
		}
	}
}

func mapField(ctx context.Context, fromField, toField reflect.Value) {
	// First we will Handle pointers
	if fromField.Kind() == reflect.Ptr && !fromField.IsNil() {
		fromField = fromField.Elem()
		// We use .Elem so we can get actual value that is not nil
	}
	if toField.Kind() == reflect.Ptr && !toField.IsNil() {
		toField = toField.Elem()
	}
	if !fromField.IsValid() || !toField.IsValid() {
		// If either feild invalid
		return
	}

	fromKind := fromField.Kind()
	toKind := toField.Kind()

	if fromKind == reflect.Interface && !fromField.IsNil() {
		realValue := reflect.ValueOf(fromField.Interface())

		if realValue.Type().AssignableTo(toField.Type()) {
			toField.Set(realValue)
			return
		}

		if toField.Kind() == reflect.String {
			switch realValue.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				toField.SetString(strconv.FormatInt(realValue.Int(), 10))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				toField.SetString(strconv.FormatUint(realValue.Uint(), 10))
			case reflect.Float32, reflect.Float64:
				toField.SetString(strconv.FormatFloat(realValue.Float(), 'f', -1, 64))
			case reflect.String:
				toField.SetString(realValue.String())
			}
			return
		}
	}

	// Handling sql.Null types

	switch v := fromField.Interface().(type) {
	case sql.NullInt64:
		if v.Valid {
			fromField = reflect.ValueOf(v.Int64)
			fromKind = reflect.Int64
		} else {
			fromField = reflect.ValueOf(int64(0))
		}
	case sql.NullInt32:
		if v.Valid {
			fromField = reflect.ValueOf(v.Int32)
			fromKind = reflect.Int32
		} else {
			fromField = reflect.ValueOf(int32(0))
		}
	case sql.NullInt16:
		if v.Valid {
			fromField = reflect.ValueOf(v.Int16)
			fromKind = reflect.Int16
		} else {
			fromField = reflect.ValueOf(int16(0))
		}
	case sql.NullString:
		if v.Valid {
			fromField = reflect.ValueOf(v.String)
			fromKind = reflect.String
		} else {
			fromField = reflect.ValueOf(" ")
		}
	case sql.NullFloat64:
		if v.Valid {
			fromField = reflect.ValueOf(v.Float64)
			fromKind = reflect.Float64
		} else {
			fromField = reflect.ValueOf(float64(0))
		}
	}

	// HANDLING Strings

	if toKind == reflect.String {
		switch fromKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			toField.SetString(strconv.FormatInt(fromField.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			toField.SetString(strconv.FormatUint(fromField.Uint(), 10))
		case reflect.Float32, reflect.Float64:
			toField.SetString(strconv.FormatFloat(fromField.Float(), 'f', -1, 64))
		case reflect.String:
			toField.SetString(fromField.String())
		}
		return
	}

	if fromKind == reflect.String {
		clean := strings.ReplaceAll(strings.TrimSpace(fromField.String()), ",", "")

		switch toKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val, err := strconv.ParseInt(clean, 10, toField.Type().Bits()); err == nil {
				toField.SetInt(val)
			}
			return
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if val, err := strconv.ParseUint(clean, 10, toField.Type().Bits()); err == nil {
				toField.SetUint(val)
			}
			return
		case reflect.Float32, reflect.Float64:
			if val, err := strconv.ParseFloat(clean, toField.Type().Bits()); err == nil {
				toField.SetFloat(val)
			}
			return
		case reflect.String:
			toField.SetString(clean)
			return
		}
	}

	// Numeric Conversions

	if (fromKind == reflect.Int || fromKind == reflect.Int8 || fromKind == reflect.Int16 || fromKind == reflect.Int32 || fromKind == reflect.Int64) &&
		(toKind == reflect.Float32 || toKind == reflect.Float64) {
		toField.SetFloat(float64(fromField.Int()))
		return
	}

	if (fromKind == reflect.Float32 || fromKind == reflect.Float64) &&
		(toKind == reflect.Int || toKind == reflect.Int8 || toKind == reflect.Int16 || toKind == reflect.Int32 || toKind == reflect.Int64) {
		toField.SetInt(int64(fromField.Float()))
		return
	}
	if (fromKind == reflect.Uint || fromKind == reflect.Uint8 || fromKind == reflect.Uint16 || fromKind == reflect.Uint32 || fromKind == reflect.Uint64) &&
		(toKind == reflect.Float32 || toKind == reflect.Float64) {
		toField.SetFloat(float64(fromField.Uint()))
		return
	}

	if (fromKind == reflect.Float32 || fromKind == reflect.Float64) &&
		(toKind == reflect.Uint || toKind == reflect.Uint8 || toKind == reflect.Uint16 || toKind == reflect.Uint32 || toKind == reflect.Uint64) {
		toField.SetUint(uint64(fromField.Float()))
		return
	}
  // If same
	if fromField.Type().AssignableTo(toField.Type()) {
		toField.Set(fromField)
	}
}

func main() {
	ctx := context.Background()
	from := From{
		ID:      "1234",
		Balance: 123.345,
	}
	to := To{}
	mapStructFields(ctx, &from, &to)
	fmt.Println("Mapped Struct :", to)
}

package utils

import (
	"database/sql"
	"reflect"
)

func ScanStructByDBTags(rows *sql.Row, dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()
	fields := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("db")
		if tag != "" && tag != "-" {
			fields[i] = v.Field(i).Addr().Interface()
		}
	}

	return rows.Scan(fields...)
}

func ScanStructByDBTagsForRows(rows *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()
	fields := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("db")
		if tag != "" && tag != "-" {
			fields[i] = v.Field(i).Addr().Interface()
		}
	}

	return rows.Scan(fields...)
}

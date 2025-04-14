package utils

import (
	"database/sql"
	"fmt"
	"reflect"
)

func ScanStructByDBTags(rows *sql.Rows, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("db tags not found: %w", err)
	}

	values := make([]any, len(columns))

	tagToField := make(map[string]int)
	for i := range make([]int, t.NumField()) {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" {
			tagToField[tag] = i
		}
	}

	for i, colName := range columns {
		fieldIndex, ok := tagToField[colName]
		if !ok {
			values[i] = new(any)
			continue
		}

		values[i] = v.Field(fieldIndex).Addr().Interface()
	}

	return rows.Scan(values...)
}

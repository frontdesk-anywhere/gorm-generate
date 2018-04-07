package gormgen

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

const tableSchemasQuery = `SELECT
    TABLE_NAME,
    COLUMN_NAME,
    IS_NULLABLE,
    DATA_TYPE,
    CHARACTER_MAXIMUM_LENGTH,
    NUMERIC_PRECISION,
    NUMERIC_SCALE,
    COLUMN_TYPE,
    COLUMN_KEY,
    EXTRA,
    COLUMN_DEFAULT
FROM COLUMNS
WHERE TABLE_SCHEMA = ?
ORDER BY TABLE_NAME, ORDINAL_POSITION
`

const primaryKeysQuery = `SELECT k.COLUMN_NAME, k.TABLE_NAME
FROM INFORMATION_SCHEMA.table_constraints t
LEFT JOIN INFORMATION_SCHEMA.key_column_usage k
USING(constraint_name,table_schema,table_name)
WHERE t.constraint_type='PRIMARY KEY'
	AND t.table_schema = ?;
`

// TableSchema contains the schemas of its columns in order.
type TableSchema []*ColumnSchema

type ColumnSchema struct {
	TableName              string
	ColumnName             string
	IsNullable             bool
	IsPrimaryKey           bool
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string
	Extra                  sql.NullString
	ColumnDefault          sql.NullString
}

// ReadDbSchema gathers the info about database structure using information_schema
// returns columns, if there are errors fatal gets thrown
func ReadDbSchema(dsn string) (map[string]TableSchema, error) {
	// Connect to the database. Switch to the INFORMATION_SCHEMA database to read table schema information.
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	u.Path = "INFORMATION_SCHEMA"

	conn, err := ConnectToDsn(u.String())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Query the database for all table primary keys
	keys, err := conn.Raw(primaryKeysQuery, dbName).Rows()
	primaryKeys := make(map[string]bool)
	if err != nil {
		return nil, err
	}
	for keys.Next() {
		colName := ""
		tableName := ""
		err = keys.Scan(
			&colName,
			&tableName,
		)
		if err != nil {
			return nil, err
		}

		primaryKeys[fmt.Sprintf("%v %v", tableName, colName)] = true
	}

	// Query the database for all table schemas.
	rows, err := conn.Raw(tableSchemasQuery, dbName).Rows()
	if err != nil {
		return nil, err
	}

	// Parse the returned schemas.
	tables := make(map[string]TableSchema)
	for rows.Next() {
		isNullable := ""
		cs := ColumnSchema{}
		err = rows.Scan(
			&cs.TableName,
			&cs.ColumnName,
			&isNullable,
			&cs.DataType,
			&cs.CharacterMaximumLength,
			&cs.NumericPrecision,
			&cs.NumericScale,
			&cs.ColumnType,
			&cs.ColumnKey,
			&cs.Extra,
			&cs.ColumnDefault,
		)
		if err != nil {
			return nil, err
		}

		cs.IsNullable = isNullable == "YES"

		if _, ok := primaryKeys[fmt.Sprintf("%v %v", cs.TableName, cs.ColumnName)]; ok {
			cs.IsPrimaryKey = true
		}

		_, ok := tables[cs.TableName]
		if !ok {
			tables[cs.TableName] = make(TableSchema, 0)
		}
		tables[cs.TableName] = append(tables[cs.TableName], &cs)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

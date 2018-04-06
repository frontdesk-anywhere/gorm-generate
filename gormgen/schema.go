package gormgen

import (
	"database/sql"
	"net/url"
	"strings"
)

type TableSchema map[string]*ColumnSchema

type ColumnSchema struct {
	TableName              string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string
	Extra                  string
	ColumnDefault          string
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

	// Query the database for all table schemas.
	rows, err := conn.Raw(tableSchemasQuery, dbName).Rows()
	if err != nil {
		return nil, err
	}

	// Parse the returned schemas.
	tables := make(map[string]TableSchema)
	for rows.Next() {
		cs := ColumnSchema{}
		err := rows.Scan(
			&cs.TableName,
			&cs.ColumnName,
			&cs.IsNullable,
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
		_, ok := tables[cs.TableName]
		if !ok {
			tables[cs.TableName] = make(TableSchema)
		}
		tables[cs.TableName][cs.ColumnName] = &cs
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

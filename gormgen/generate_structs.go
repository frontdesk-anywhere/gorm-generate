package gormgen

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

const gormStructsTemplate = `package {{.StructsPackage}}

import ({{range .Imports}}
	"{{.}}"{{end}}
)
{{range $tableName,$tableSchema := .DbSchema}}
type {{$tableName}} struct { {{range $columnName,$columnSchema := $tableSchema}}
	{{$columnSchema.GoColumnName}} {{$columnSchema.GoColumnType}} ` + "`db:\"{{$columnSchema.DbColumnName}}\"{{$columnSchema.GormTag}}`" + `{{end}}
}
{{end}}`

const gormRegistryTemplate = `package {{.StructsPackage}}

// AllModels returns a list of empty GORM DB models available in the {{.DbName}} database.
func AllModels() []interface{} {
	return []interface{} { {{range $tableName,$tableSchema := .DbSchema}}
		&{{$tableName}} { },{{end}}
	}
}
`

// StructsContext struct
type StructsContext struct {
	StructsPackage string
	Imports        map[string]string
	DbSchema       map[string]TableContext
	DbName         string
}

// TableContext struct
type TableContext []*ColumnContext

// ColumnContext struct
type ColumnContext struct {
	DbColumnName string
	GoColumnName string
	GoColumnType string
	GormTag      string
}

// CreateTemplateContext creates context for template
func (g *Generator) CreateTemplateContext(dbSchema map[string]TableSchema) (*StructsContext, error) {
	var dbContext = make(map[string]TableContext)
	var importMap = make(map[string]string)

	dnsParsed, err := url.Parse(g.DbDsn)
	if err != nil {
		return nil, err
	}
	dbName := strings.Replace(dnsParsed.Path, "/", "", -1)

	for _, tableSchema := range dbSchema {
		for _, columnSchema := range tableSchema {
			goType, requiredImport, err := goType(columnSchema)
			if err != nil {
				return nil, err
			}
			if requiredImport != "" {
				importMap[requiredImport] = requiredImport
			}

			goTableName := formatName(columnSchema.TableName)
			goColumnName := formatName(columnSchema.ColumnName)
			colContext := ColumnContext{
				DbColumnName: columnSchema.ColumnName,
				GoColumnName: goColumnName,
				GoColumnType: goType,
				GormTag:      gormTag(columnSchema),
			}

			_, ok := dbContext[goTableName]
			if !ok {
				dbContext[goTableName] = make(TableContext, 0)
			}
			dbContext[goTableName] = append(dbContext[goTableName], &colContext)
		}
	}
	absOutputPath, err := filepath.Abs(g.OutputPath)
	if err != nil {
		return nil, err
	}
	return &StructsContext{
		StructsPackage: path.Base(absOutputPath),
		Imports:        importMap,
		DbSchema:       dbContext,
		DbName:         dbName,
	}, nil
}

// GenerateGormStructs generates gorm structs
func (g *Generator) GenerateGormStructs() error {
	// Read the schemas of the tables in the provided database.
	dbSchema, err := ReadDbSchema(g.DbDsn)
	if err != nil {
		return err
	}

	// Build the template context.
	ctx, err := g.CreateTemplateContext(dbSchema)
	if err != nil {
		return err
	}

	// Render the templates.
	g.GenerateTemplateWithContext(gormStructsTemplate, g.StructsFile, ctx)
	if err != nil {
		return err
	}
	return g.GenerateTemplateWithContext(gormRegistryTemplate, g.StructsRegistryFile, ctx)
}

// formatName takes db column/table name and converts it to GoLang naming convention
// returns name in GoLang style
func formatName(name string) string {
	parts := strings.Split(name, "_")
	newName := ""
	for _, p := range parts {
		if len(p) < 1 {
			continue
		}
		newName = newName + strings.Replace(p, string(p[0]), strings.ToUpper(string(p[0])), 1)
	}
	return newName
}

// goType takes database column schema and converts it to golang type. Also returns the required imported
// (if any). Returns an error if type cannot be converted.
func goType(col *ColumnSchema) (string, string, error) {
	switch col.DataType {
	case "char", "varchar", "enum", "set", "text", "longtext", "mediumtext", "tinytext":
		if col.IsNullable {
			return "sql.NullString", "database/sql", nil
		} else {
			return "string", "", nil
		}
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		return "[]byte", "", nil
	case "date", "time", "datetime", "timestamp":
		return "time.Time", "time", nil
	case "bit", "tinyint", "smallint", "int", "mediumint", "bigint":
		if col.IsNullable {
			return "sql.NullInt64", "database/sql", nil
		} else {
			return "int64", "", nil
		}
	case "float", "decimal", "double":
		if col.IsNullable {
			return "sql.NullFloat64", "database/sql", nil
		} else {
			return "float64", "", nil
		}
	default:
		err := fmt.Errorf(
			"unrecognized type in column %s.%s found: %s",
			col.TableName,
			col.ColumnName,
			col.DataType,
		)
		return "", "", err
	}
}

// gormTag takes a database column schema and converts it into a gorm tag.
func gormTag(schema *ColumnSchema) string {
	tagParts := []string{}
	if schema.IsPrimaryKey {
		tagParts = append(tagParts, "primary_key")
	} else {
		tagParts = append(tagParts, "type:"+schema.DataType)
	}
	if !schema.IsNullable {
		tagParts = append(tagParts, "not null")
	}
	if schema.Extra.Valid && strings.Contains(schema.Extra.String, "auto_increment") {
		tagParts = append(tagParts, "AUTO_INCREMENT")
	}

	tag := ""
	if len(tagParts) > 0 {
		tag = strings.Join(tagParts, ";")
		tag = fmt.Sprintf(` gorm:"%s"`, tag)
	}
	return tag
}

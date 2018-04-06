package gormgen

import (
	"bytes"
	"go/format"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jinzhu/gorm"
)

// Connect creates an open connection to a database via gorm.
// If an error occurs, the returned connection will be `nil`
func ConnectToDsn(dsnStr string) (*gorm.DB, error) {
	dsn, err := url.Parse(dsnStr)
	if err != nil {
		return nil, err
	}

	// Delete the scheme and rebuild the DSN for GORM.
	dbScheme := dsn.Scheme
	dsn.Scheme = ""
	db, err := gorm.Open(dbScheme, strings.TrimPrefix(dsn.String(), "//"))
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Generator struct {
	OutputPath          string
	DbDsn               string
	StructsFile         string
	StructsRegistryFile string
}

// GenerateTemplate is used for file generation based on template and path
func (g *Generator) GenerateTemplate(tmpl string, filePath string) error {
	return g.GenerateTemplateWithContext(tmpl, filePath, nil)
}

// GenerateTemplateWithContext is used for file generation based on template, path, and context
func (g *Generator) GenerateTemplateWithContext(tmpl string, fileName string, context interface{}) error {
	filePath, err := filepath.Abs(path.Join(g.OutputPath, fileName))
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	generatedTemplate, err := template.New("template-" + path.Base(filePath)).Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = generatedTemplate.Execute(&buf, context)
	if err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = file.Write(formatted)
	return err
}

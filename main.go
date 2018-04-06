package main

import (
	"flag"
	"log"

	"github.com/frontdesk-anywhere/gorm-generate/gormgen"
)

func parseFlags() *gormgen.Generator {
	config := gormgen.Generator{}
	flag.StringVar(
		&config.DbDsn,
		"dsn",
		"",
		"A data source name specified as a gorm URI. The uri scheme is used to distinguish driver types. Required.",
	)
	flag.StringVar(
		&config.OutputPath,
		"output",
		".",
		"The path to generate code into.",
	)
	flag.StringVar(
		&config.StructsFile,
		"structs-file",
		"generated_structs.go",
		"The path to write generate structs to relative to -output.",
	)
	flag.StringVar(
		&config.StructsRegistryFile,
		"structs-registry-file",
		"generated_structs_registry.go",
		"The path to write the generate structs registry to relative to -output.",
	)
	flag.Parse()
	return &config
}

func main() {
	config := parseFlags()

	targets := flag.Args()
	if len(targets) == 0 {
		targets = []string{"all"}
	}

	for _, target := range targets {
		switch target {
		case "structs":
			err := config.GenerateGormStructs()
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}
}

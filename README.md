# `gorm-generate`

`gorm-generate` generates GORM structs from a live database.

## Usage

To run gorm generate:

```
go get -u github.com/frontdesk-anywhere/gorm-generate
go install github.com/frontdesk-anywhere/gorm-generate

cd ${dir_where_you_want_to_generate_structs}

gorm-generate \
    -dsn="mysql://root:root@tcp(127.0.0.1:3306)?your_database_name_here/?charset=utf8"

# Check the results.
cat generated_structs.go
cat generated_structs_registry.go
```

To use `gorm-generate` with `go generate`, add the following to the top of a go file (not with the
same name as the file you will be generating) in the package you wish to generate code into:

```
//go:generate gorm-generate -dsn="..."
```

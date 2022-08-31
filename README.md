# mdd-go: Markdown Driven (Literate) Development framework for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/shibukawa/mdd-go.svg)](https://pkg.go.dev/github.com/shibukawa/mdd-go)

Go library that maps markdown file content into struct.

## Features

* Mapping heading hierarchy to struct composition
* Parse heading text to map to struct field
  * Specify optional parameters in heading text
* Assign code block fence content to struct field
* Parse table and map to struct field
  * As a slice of struct
  * As a `map[string]string`
* Define aliases (l10n) about heading titles

## Simple Usage

Sample Document to parse:

~~~md
# Query User

You can mapping heading titles, code block fences, tables to Go's struct field.

This part is ignored. You can add detail description, design history, poem, and so on.

```sql
select user_id, name, email from users where name=/*name*/'bob';
```

## CRUD Matrix

Heading structure is important.

| Table | C | R | U | D | Description            |
|-------|---|---|---|---|------------------------|
| users |   | X |   |   | name is indexed        |
~~~

Sample Code:

```go
package main


import (
	"fmt"
    "log"
    "os"

	"github.com/shibukawa/mdd-go"
)

// Define structure
type SQLDoc struct {
	Name       string
	SQL        string
	CRUDMatrix []CRUDMatrix
}

type CRUDMatrix struct {
	Table       string
	C           bool
	R           bool
	U           bool
	D           bool
	Description string
}

var sqlJig = mdd.NewDocJig[SQLDoc]()

func init() {
    // Define markdown structure
	root := sqlJig.Root()
	root.Label("Name")
	root.CodeFence("SQL", "sql")

	crud := root.Child(".", "CRUD Matrix").Table("CRUDMatrix")
	crud.Field("Table").Required()
	crud.Field("C").Required()
	crud.Field("R").Required()
	crud.Field("U").Required()
	crud.Field("D").Required()
	crud.Field("Description")
}

func main() {
    md, err := os.Open("sample.md")
    if err != nil {
        log.Fatal(err)
    }
    defer md.Close()
	doc, _ := sqlJig.Parse(md)
	fmt.Printf("err: %v\n", err)
	fmt.Printf("Name: %v\n", doc.Name)
	fmt.Printf("SQL: %v\n", doc.SQL)
	fmt.Printf("CRUD: %v\n", doc.CRUDMatrix[0])
}
```

## License

Apache 2

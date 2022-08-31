package mdd_test

import (
	"fmt"

	"github.com/shibukawa/mdd-go"
)

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

var sqlSampleDoc = `
# Query User

You can mapping heading titles, code block fences, tables to Go's struct field.

This part is ignored. You can add detail description, design history, poem, and so on.

~~~sql
select user_id, name, email from users where name=/*name*/'bob';
~~~

## CRUD Matrix

Heading structure is important.

| Table | C | R | U | D | Description            |
|-------|---|---|---|---|------------------------|
| users |   | X |   |   | name is indexed        |
`

func Example() {
	// Usually these code is in init() function
	sqlJig := mdd.NewDocJig[SQLDoc]()
	root := sqlJig.Root()
	// Store root heading title to Name field
	root.Label("Name")
	// Store code fence block (sql) to SQL field
	root.CodeFence("SQL", "sql")

	// Store table to struct slice named CRUDMatrix
	// "." specifies struct hierarchy.
	// If it is name, mdd-go dig to the child element with the name
	// "." keeps same depth of struct structure
	crud := root.Child(".", "CRUD Matrix").Table("CRUDMatrix")
	crud.Field("Table").Required()
	crud.Field("C").Required()
	crud.Field("R").Required()
	crud.Field("U").Required()
	crud.Field("D").Required()
	crud.Field("Description")

	doc, err := sqlJig.ParseString(sqlSampleDoc)
	fmt.Printf("err: %v\n", err)
	fmt.Printf("Name: %v\n", doc.Name)
	fmt.Printf("SQL: %v\n", doc.SQL)
	fmt.Printf("CRUD: %v\n", doc.CRUDMatrix[0])

	// Output:
	// err: <nil>
	// Name: Query User
	// SQL: select user_id, name, email from users where name=/*name*/'bob';
	// CRUD: {users false true false false name is indexed}
}

type RootLayer struct {
	Name string
	Level2
}

type Level2 struct {
	Name string
	Level3
}

type Level3 struct {
	Name         string
	BoolOption   bool
	StringOption string
	IntOption    int
}

var nestedSampleAndLabel = `
# Root

## Level2 Heading: Child

### Level3 Heading: Grandchild (BoolOption, StringOption=option, IntOption=100)

`

func ExampleLayer() {
	// Usually these code is in init() function
	nestedJig := mdd.NewDocJig[RootLayer]()
	root := nestedJig.Root()

	// This code store whole title string to Name field
	root.Label("Name")

	// This layer has "pattern prefix" as a second param
	// mdd-go searching the heading that has this pattern.
	// Name field store a part of heading label that trims pattern part.
	level2 := root.Child("Level2", "Level2 Heading").Label("Name")

	// The layer title can have option that is in paren ().
	// Each Option() has destination field name.
	level3 := level2.Child("Level3", "Level3 Heading").Label("Name")
	level3.Option("StringOption")
	level3.Option("BoolOption")
	level3.Option("IntOption")

	doc := nestedJig.MustParseString(nestedSampleAndLabel)

	fmt.Printf("level1.Name: %s\n", doc.Name)
	fmt.Printf("level2.Name: %s\n", doc.Level2.Name)
	fmt.Printf("level3.Name: %s\n", doc.Level2.Level3.Name)
	fmt.Printf("      .StringOption: %s\n", doc.Level3.StringOption)
	fmt.Printf("      .BoolOption: %v\n", doc.Level3.BoolOption)
	fmt.Printf("      .IntOption: %d\n", doc.Level3.IntOption)
	// Output:
	// level1.Name: Root
	// level2.Name: Child
	// level3.Name: Grandchild
	//       .StringOption: option
	//       .BoolOption: true
	//       .IntOption: 100
}

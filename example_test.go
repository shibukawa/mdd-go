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

	doc, _ := sqlJig.ParseString(sqlSampleDoc)
	fmt.Printf("Name: %v\n", doc.Name)
	fmt.Printf("SQL: %v\n", doc.SQL)
	fmt.Printf("CRUD: %v\n", doc.CRUDMatrix[0])

	// Output:
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

func ExampleLayout() {
	// Usually these code is in init() function
	nestedJig := mdd.NewDocJig[RootLayer]()
	root := nestedJig.Root()

	// This code store whole title string to Name field
	root.Label("Name")

	// This layer has "pattern prefix" as a second param
	// mdd-go searching the heading that has this pattern.
	// Name field store a part of heading label that trims pattern part.
	level2 := root.Child("Level2", "Level2 Heading")
	level2.Label("Name")

	// The layer title can have option that is in paren ().
	// Each Option() has destination field name.
	level3 := level2.Child("Level3", "Level3 Heading").Label("Name")
	level3.Option("StringOption")
	level3.Option("BoolOption")
	level3.Option("IntOption")

	var nestedSampleAndLabel = `
# Root

## Level2 Heading: Child

### Level3 Heading: Grandchild (BoolOption, StringOption=option, IntOption=100)
`

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

type Shell struct {
	Name string
	Code string
}

type LiterateShell struct {
	Shells []Shell
}

func ExampleCodeFence() {
	// Usually these code is in init() function
	shellJig := mdd.NewDocJig[LiterateShell]()
	root := shellJig.Root()

	// Store multiple layers into Shells slice
	shells := root.Children("Shells")
	shells.Label("Name")
	shells.CodeFence("Code")

	var literateShellSrc = `
# Recovery Batch

## Dump

Dump database and copy the file to S3.

~~~sh
$ pg_dumpall -Fc -f mydb.dump mydb
$ aws s3 cp mydb.dump s3://backup-bucket/mydb.dump
~~~

## Restore

Retreat database dump from S3 and restore DB.

~~~
$ aws s3 cp s3://backup-bucket/mydb.dump mydb.dump
$ pg_restore -d mydb mydb.dump
~~~
`

	parsedShells := shellJig.MustParseString(literateShellSrc)

	for _, s := range parsedShells.Shells {
		fmt.Printf("Name: %s\n", s.Name)
		fmt.Println("Code:")
		fmt.Println(s.Code)
		fmt.Println("")
	}
	// Output:
	// Name: Dump
	// Code:
	// $ pg_dumpall -Fc -f mydb.dump mydb
	// $ aws s3 cp mydb.dump s3://backup-bucket/mydb.dump
	//
	// Name: Restore
	// Code:
	// $ aws s3 cp s3://backup-bucket/mydb.dump mydb.dump
	// $ pg_restore -d mydb mydb.dump
}

type User struct {
	ID    int
	Name  string
	Email string
}

type UserList struct {
	Users []User
}

func ExampleTable() {
	// Usually these code is in init() function
	userJig := mdd.NewDocJig[UserList]()
	root := userJig.Root()
	// Don't create instance for section title.
	// So table rows under *Users sections are
	// merged.
	usersSection := root.Children(".", "")

	// Define table column mapping to struct field
	userTable := usersSection.Table("Users")
	userTable.Field("ID")
	userTable.Field("Name")
	userTable.Field("Email")

	var userDataDocument = `
# Define Users

## Regular Users

| ID | Name  | Email             |
|----|-------|-------------------|
| 1  | Alan  | alan@example.com  |
| 2  | Ellie | ellie@example.com |

## Security Expert Users

| ID | Name  | Email             |
|----|-------|-------------------|
| 3  | Alice | alice@example.com |
| 4  | Bob   | bob@example.com   |

## Admin Users

You can shuffle column order:

| ID | Email             | Name   |
|----|-------------------|--------|
| 5  | denis@example.com | Dennis |
`

	doc := userJig.MustParseString(userDataDocument)

	for _, u := range doc.Users {
		fmt.Printf("%s (ID=%d, Email=%s)\n", u.Name, u.ID, u.Email)
	}
	// Output:
	// Alan (ID=1, Email=alan@example.com)
	// Ellie (ID=2, Email=ellie@example.com)
	// Alice (ID=3, Email=alice@example.com)
	// Bob (ID=4, Email=bob@example.com)
	// Dennis (ID=5, Email=denis@example.com)
}

func ExampleAlias() {
	// Usually these code is in init() function
	sqlJig := mdd.NewDocJig[SQLDoc]()
	sqlJig.Alias("CRUD Matrix", "CRUDTable").Lang("ja", "CRUDマトリックス", "CRUD表")
	sqlJig.Alias("Table").Lang("ja", "テーブル", "表")
	sqlJig.Alias("Description", "Desc", "Detail").Lang("ja", "説明", "詳細")

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

	var sqlSampleDoc = `
# Japanese Sample

You can absorb label variations including other language.

~~~sql
select user_id, name, email from users where name=/*name*/'bob';
~~~

## CRUDマトリックス

Heading structure is important.

| テーブル | C | R | U | D | 詳細                       |
|----------|---|---|---|---|----------------------------|
| users    |   | X |   |   | nameにはインデックスを貼る |
`

	doc, _ := sqlJig.ParseString(sqlSampleDoc)
	fmt.Printf("Name: %v\n", doc.Name)
	fmt.Printf("SQL: %v\n", doc.SQL)
	fmt.Printf("CRUD: %v\n", doc.CRUDMatrix[0])

	// Output:
	// Name: Japanese Sample
	// SQL: select user_id, name, email from users where name=/*name*/'bob';
	// CRUD: {users false true false false nameにはインデックスを貼る}
}

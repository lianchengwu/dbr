dbr (fork of gocraft/dbr) provides additions to Go's database/sql for super fast performance and convenience.

[![Build Status](https://travis-ci.org/mailru/dbr.svg?branch=master)](https://travis-ci.org/mailru/dbr)
[![Go Report Card](https://goreportcard.com/badge/github.com/lianchengwu/dbr)](https://goreportcard.com/report/github.com/lianchengwu/dbr)
[![Coverage Status](https://coveralls.io/repos/github/mailru/dbr/badge.svg?branch=develop)](https://coveralls.io/github/mailru/dbr?branch=develop)

## Getting Started

```go
// create a connection (e.g. "postgres", "mysql", or "sqlite3")
conn, _ := dbr.Open("postgres", "...")

// create a session for each business unit of execution (e.g. a web request or goworkers job)
sess := conn.NewSession(nil)

// get a record
var suggestion Suggestion
sess.Select("id", "title").From("suggestions").Where("id = ?", 1).Load(&suggestion)

// JSON-ready, with dbr.Null* types serialized like you want
json.Marshal(&suggestion)
```

## Feature highlights

### Use a Sweet Query Builder or use Plain SQL

mailru/dbr supports both.

Sweet Query Builder:
```go
stmt := dbr.Select("title", "body").
	From("suggestions").
	OrderBy("id").
	Limit(10)
```

Plain SQL:

```go
builder := dbr.SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
```

### Amazing instrumentation with session

All queries in mailru/dbr are made in the context of a session. This is because when instrumenting your app, it's important to understand which business action the query took place in.

Writing instrumented code is a first-class concern for mailru/dbr. We instrument each query to emit to a EventReceiver interface.

### Faster performance than using database/sql directly
Every time you call database/sql's db.Query("SELECT ...") method, under the hood, the mysql driver will create a prepared statement, execute it, and then throw it away. This has a big performance cost.

mailru/dbr doesn't use prepared statements. We ported mysql's query escape functionality directly into our package, which means we interpolate all of those question marks with their arguments before they get to MySQL. The result of this is that it's way faster, and just as secure.

Check out these [benchmarks](https://github.com/tyler-smith/golang-sql-benchmark).

### IN queries that aren't horrible
Traditionally, database/sql uses prepared statements, which means each argument in an IN clause needs its own question mark. mailru/dbr, on the other hand, handles interpolation itself so that you can easily use a single question mark paired with a dynamically sized slice.
```go
ids := []int64{1, 2, 3, 4, 5}
builder.Where("id IN ?", ids) // `id` IN ?
```
map object can be used for IN queries as well.
Note: interpolation map is slower than slice and it is preferable to use slice when it is possible.
```go
ids := map[int64]string{1: "one", 2: "two"}
builder.Where("id IN ?", ids)  // `id` IN ?
```

### JSON Friendly
Every try to JSON-encode a sql.NullString? You get:
```json
{
	"str1": {
		"Valid": true,
		"String": "Hi!"
	},
	"str2": {
		"Valid": false,
		"String": ""
  }
}
```

Not quite what you want. mailru/dbr has dbr.NullString (and the rest of the Null* types) that encode correctly, giving you:

```json
{
	"str1": "Hi!",
	"str2": null
}
```

### Inserting multiple records

```go
sess.InsertInto("suggestions").Columns("title", "body").
  Record(suggestion1).
  Record(suggestion2)
```

### Updating records on conflict

```go
stmt := sess.InsertInto("suggestions").Columns("title", "body").Record(suggestion1)
stmt.OnConflict("suggestions_pkey").Action("body", dbr.Proposed("body"))
```


### Updating records

```go
sess.Update("suggestions").
	Set("title", "Gopher").
	Set("body", "I love go.").
	Where("id = ?", 1)
```

### Transactions

```go
tx, err := sess.Begin()
if err != nil {
  return err
}
defer tx.RollbackUnlessCommitted()

// do stuff...

return tx.Commit()
```

### Load database values to variables

Querying is the heart of mailru/dbr.

* Load(&any): load everything!
* LoadStruct(&oneStruct): load struct
* LoadStructs(&manyStructs): load a slice of structs
* LoadValue(&oneValue): load basic type
* LoadValues(&manyValues): load a slice of basic types

```go
// columns are mapped by tag then by field
type Suggestion struct {
	ID int64  // id, will be autoloaded by last insert id
	Title string // title
	Url string `db:"-"` // ignored
	secret string // ignored
	Body dbr.NullString `db:"content"` // content
	User User
}

// By default dbr converts CamelCase property names to snake_case column_names
// You can override this with struct tags, just like with JSON tags
// This is especially helpful while migrating from legacy systems
type Suggestion struct {
	Id        int64
	Title     dbr.NullString `db:"subject"` // subjects are called titles now
	CreatedAt dbr.NullTime
}

var suggestions []Suggestion
sess.Select("*").From("suggestions").Load(&suggestions)
```

### Join multiple tables

dbr supports many join types:

```go
sess.Select("*").From("suggestions").
  Join("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  RightJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  FullJoin("subdomains", "suggestions.subdomain_id = subdomains.id")
```

You can join on multiple tables:

```go
sess.Select("*").From("suggestions").
  Join("subdomains", "suggestions.subdomain_id = subdomains.id").
  Join("accounts", "subdomains.accounts_id = accounts.id")
```

### Quoting/escaping identifiers (e.g. table and column names)

```go
dbr.I("suggestions.id") // `suggestions`.`id`
```

### Subquery

```go
sess.Select("count(id)").From(
  dbr.Select("*").From("suggestions").As("count"),
)
```

### Union

```go
dbr.Union(
  dbr.Select("*"),
  dbr.Select("*"),
)

dbr.UnionAll(
  dbr.Select("*"),
  dbr.Select("*"),
)
```

Union can be used in subquery.

### Alias/AS

* SelectStmt

```go
dbr.Select("*").From("suggestions").As("count")
```

* Identity

```go
dbr.I("suggestions").As("s")
```

* Union

```go
dbr.Union(
  dbr.Select("*"),
  dbr.Select("*"),
).As("u1")

dbr.UnionAll(
  dbr.Select("*"),
  dbr.Select("*"),
).As("u2")
```

### Building arbitrary condition

One common reason to use this is to prevent string concatenation in a loop.

* And
* Or
* Eq
* Neq
* Gt
* Gte
* Lt
* Lte

```go
dbr.And(
  dbr.Or(
    dbr.Gt("created_at", "2015-09-10"),
    dbr.Lte("created_at", "2015-09-11"),
  ),
  dbr.Eq("title", "hello world"),
)
```

### Built with extensibility

The core of dbr is interpolation, which can expand `?` with arbitrary SQL. If you need a feature that is not currently supported,
you can build it on your own (or use `dbr.Expr`).

To do that, the value that you wish to be expaned with `?` needs to implement `dbr.Builder`.

```go
type Builder interface {
	Build(Dialect, Buffer) error
}
```

## Driver support

* MySQL
* PostgreSQL
* SQLite3
* ClickHouse

These packages were developed by the [engineering team](https://eng.uservoice.com) at [UserVoice](https://www.uservoice.com) and currently power much of its infrastructure and tech stack.

## Thanks & Authors
Inspiration from these excellent libraries:
* [sqlx](https://github.com/jmoiron/sqlx) - various useful tools and utils for interacting with database/sql.
* [Squirrel](https://github.com/lann/squirrel) - simple fluent query builder.

Authors:
* Jonathan Novak -- [https://github.com/cypriss](https://github.com/cypriss)
* Tyler Smith -- [https://github.com/tyler-smith](https://github.com/tyler-smith)
* Tai-Lin Chu -- [https://github.com/taylorchu](https://github.com/taylorchu)
* Sponsored by [UserVoice](https://eng.uservoice.com)

Contributors:
* Paul Bergeron -- [https://github.com/dinedal](https://github.com/dinedal) - SQLite dialect
* Bulat Gaifullin -- [https://github.com/bgaifullin](https://github.com/bgaifullin) - ClickHouse dialect

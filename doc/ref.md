# gopsql reference

The library is used for serializing your own structs into SQL queries for PostgreSQL and then executed against a server.

## Interfaces

The library contains the following interface:

### Saveable
```go
type Saveable interface {
    GetID() int
    SetID(int)
}
```
This interface needs to be implemented by all types you wish to save.

## Functions

The library provides the following functions.

### Setup

```go
func Setup(dbHost string, dbUser string, dbPass string, dbName string, sslMode string, types []Saveable) error
```

This is the function that needs to be called first, it sets up the connection to the database and initializes the database tables based on the types provided.

#### Params

##### dbHost

The hostname or IP address of the database server.

##### dbUser

The username for accessing the database.

##### dbPass

The password associated with the username.

##### dbName

The name of the database on the server.

##### sslMode

How should SSL be used when connecting to the server?

Available options are:

- disable - No SSL
- require - Always SSL (skip verification)
- verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
- verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)

##### types

A list of types that implement the Saveable interface. These will be used to create the database tables for your models.

The tables are created with "CREATE TABLE IF NOT EXISTS". Which means that it does not overwrite existing tables. This library does not implement any support for schema migrations and you need to take care of those by yourself for now.



### InsertNew

```go
func InsertNew(obj Saveable) *sql.Row
```

This functions inserts a new row into a table.

#### Params

##### obj

A object of a type that implements Saveable. The structure of the struct is used to generate the SQL query and execute it using the values inside the struct.

The function returns a `*sql.Row`, you have to handle the result from the database server yourself.


### UpdateExisting

```go
func UpdateExisting(obj Saveable) error
```

This function updates an existing row in the database.

#### Params

##### obj

A object of a type that implements Saveable. The Saveable.GetID() function has to return the primary key (or other unique key) with which the row can be found.

The function overwrites everything in the row with the values from the provided object (except for the ID). So make sure that you retrieve the current state first.


### GetAll

```go
func GetAll(obj interface{}) (*sql.Rows, error)
```

The function returns all rows from the table that matches the provided type.

You need to handle the returned *sql.Rows yourself to map them to your type.


### GetFiltered

```go
func GetFiltered(obj interface{}, filter ...string) (*sql.Rows, error)
```

Returns a subset of rows from a table, based on the provided filter.

#### Example usage

```go
gopsql.GetFiltered(myStruct{}, "city", "new york", "AND", "age" "30")
```

#### Params

##### obj

A type that has a corresponding table in the database.

##### filter

2 or more strings that build up the filters used in the query. Currently only supports exact matches ("=", not "<" or ">").


### Delete

```go
func Delete(obj Saveable) error
```

Deletes a row from the database. The provided objects Saveable.GetID() needs to return a unique identifier (like primary key) which can be used to identify which row to delete.

#### Params

##### obj

The object containing an ID to delete.

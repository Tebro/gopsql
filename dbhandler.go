package gopsql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Tebro/logger"
	// Importing pq to support postgres
	_ "github.com/lib/pq"
)

const createTableQueryTemplate = `
	CREATE TABLE IF NOT EXISTS %s (
		%s
	);
`

func getCreateTableQuery(obj interface{}) string {
	objType := reflect.TypeOf(obj)

	columnName := objType.Name()

	var fieldLines []string
	for i := 0; i < objType.NumField(); i++ {
		objField := objType.Field(i)
		if _, ok := objField.Tag.Lookup("column_skip"); !ok {
			fieldLines = append(fieldLines, fmt.Sprintf("%s %s", objField.Name, objField.Tag.Get("column_type")))
		}
	}
	return fmt.Sprintf(createTableQueryTemplate, columnName, strings.Join(fieldLines, ","))
}

const selectAllQueryTemplate = "SELECT %s FROM %s;"

func getSelectAllQuery(obj interface{}) (string, bool) {
	objType := reflect.TypeOf(obj)

	var fieldNames []string
	orderBy := false
	var orderByField string
	for i := 0; i < objType.NumField(); i++ {
		if _, ok := objType.Field(i).Tag.Lookup("column_skip"); !ok {
			fieldNames = append(fieldNames, objType.Field(i).Name)
		}
		if _, ok := objType.Field(i).Tag.Lookup("column_order_by"); ok {
			orderBy = true
			orderByField = objType.Field(i).Name
		}
	}

	var lastPart string
	if orderBy {
		lastPart = fmt.Sprintf("%s ORDER BY %s", objType.Name(), orderByField)
	} else {
		lastPart = objType.Name()
	}

	return fmt.Sprintf(selectAllQueryTemplate, strings.Join(fieldNames, ","), lastPart), orderBy
}

func getSelectFilteredQuery(obj interface{}, filterString string) string {
	q, orderBy := getSelectAllQuery(obj)
	// Remove the semicolon
	var retval string

	if orderBy {
		parts := strings.Split(q, " ORDER BY ")
		retval = fmt.Sprintf("%s WHERE %s ORDER BY %s", parts[0], filterString, parts[1])

	} else {
		q = strings.TrimRight(q, ";")
		retval = fmt.Sprintf("%s WHERE %s;", q, filterString)
	}

	logger.Debug(fmt.Sprintf("Built filtered SELECT query: '%s'", retval))

	return retval
}

func parseFilterString(filters ...string) (string, []string, error) {
	if len(filters) < 2 {
		return "", nil, errors.New("Not enough parameters provided")
	}
	var filterValues []string

	if len(filters) == 2 {
		filterValues = append(filterValues, filters[1])
		return fmt.Sprintf("%s=$1", filters[0]), filterValues, nil
	}
	var parts []string
	if (len(filters)+1)%3 == 0 {
		valueIndex := 1
		for i := 0; i < len(filters); i += 3 {
			filterValues = append(filterValues, filters[i+1])
			parts = append(parts, fmt.Sprintf("%s=$%d", filters[i], valueIndex))
			valueIndex++
			if i+2 < len(filters) {
				parts = append(parts, filters[i+2])
			}
		}
	}
	return strings.Join(parts, " "), filterValues, nil
}

// GetAll gets all the rows for the type
func GetAll(obj interface{}) (*sql.Rows, error) {
	q, _ := getSelectAllQuery(obj)

	return db.Query(q)
}

// GetFiltered does a query with a WHERE section GetFiltered(sometype, field, value, "AND", field, value)
func GetFiltered(obj interface{}, filter ...string) (*sql.Rows, error) {
	parsedFilter, filterValues, err := parseFilterString(filter...)
	if err != nil {
		return nil, err
	}
	q := getSelectFilteredQuery(obj, parsedFilter)

	filterValueInterfaces := make([]interface{}, len(filterValues))
	for i, v := range filterValues {
		filterValueInterfaces[i] = v
	}

	return db.Query(q, filterValueInterfaces...)
}

// Saveable interface is so that Save function always has a way of getting and setting the ID
type Saveable interface {
	GetID() int
	SetID(int)
}

// UpdateExisting updates a record in the DB based on ID
func UpdateExisting(obj Saveable) error {
	q, values := getUpdateQuery(obj)

	valueInterfaces := make([]interface{}, len(values))
	for i, v := range values {
		valueInterfaces[i] = v
	}

	_, err := db.Exec(q, valueInterfaces...)
	return err
}

// InsertNew inserts a new entry and returns the Row for someone with more knowledge to retrieve ID
func InsertNew(obj Saveable) *sql.Row {
	q, values := getInsertQuery(obj)

	valueInterfaces := make([]interface{}, len(values))
	for i, v := range values {
		valueInterfaces[i] = v
	}

	logger.Debug(fmt.Sprintf("Going to insert with: '%v' and %v", q, valueInterfaces))

	return db.QueryRow(q, valueInterfaces...)
}

const updateQueryTemplate = "UPDATE %s SET %s WHERE %s;"

func getUpdateQuery(obj Saveable) (string, []string) {
	objVal := reflect.ValueOf(obj)
	columnName := objVal.Type().Name()

	var updateValues []string
	var values []string
	valueIndex := 1
	for i := 0; i < objVal.NumField(); i++ {
		field := objVal.Type().Field(i)
		_, found1 := field.Tag.Lookup("column_skip")
		_, found2 := field.Tag.Lookup("column_skip_insert")
		if !found1 && !found2 {
			vInter := objVal.Field(i).Interface()
			val := reflect.ValueOf(vInter)

			theValue := fmt.Sprintf("%v", val)

			updateValues = append(updateValues, fmt.Sprintf("%s=$%d", field.Name, valueIndex))
			valueIndex++
			values = append(values, theValue)
		}
	}
	filter := fmt.Sprintf("ID=$%d", valueIndex)
	values = append(values, fmt.Sprintf("%v", obj.GetID())) //Add the ID for the WHERE part

	return fmt.Sprintf(updateQueryTemplate, columnName, strings.Join(updateValues, ","), filter), values

}

const insertQueryTemplate = "INSERT INTO %s (%s) VALUES (%s) RETURNING ID;"

func getInsertQuery(obj Saveable) (string, []string) {
	objVal := reflect.ValueOf(obj)
	columnName := objVal.Type().Name()

	// Build Fields and values list
	var fields []string
	var values []string
	var valuePlaceholders []string
	valueIndex := 1
	for i := 0; i < objVal.NumField(); i++ {
		field := objVal.Type().Field(i)
		_, found1 := field.Tag.Lookup("column_skip")
		_, found2 := field.Tag.Lookup("column_skip_insert")
		if !found1 && !found2 {
			fields = append(fields, field.Name)

			// Getting the actual value
			vInter := objVal.Field(i).Interface()
			val := reflect.ValueOf(vInter)
			theValue := fmt.Sprintf("%v", val)

			values = append(values, theValue)
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", valueIndex))
			valueIndex++
		}
	}

	return fmt.Sprintf(insertQueryTemplate, columnName, strings.Join(fields, ", "), strings.Join(valuePlaceholders, ", ")), values

}

const deleteQueryTemplate = "DELETE FROM %s WHERE ID=$1;"

// Delete removes the provided entry from the database, based on the return value of GetID()
func Delete(obj Saveable) error {

	columnName := reflect.ValueOf(obj).Type().Name()
	query := fmt.Sprintf(deleteQueryTemplate, columnName)

	_, err := db.Exec(query, obj.GetID())
	return err
}

// Doc: https://godoc.org/github.com/lib/pq
var db *sql.DB

// Setup connects to the database and makes sure the schema is inserted
func Setup(dbHost string, dbUser string, dbPass string, dbName string, sslMode string, types []Saveable) error {

	_sslMode := "disable"
	if sslMode != "" {
		_sslMode = sslMode
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbName, _sslMode)

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to open database connection. Connection URL: %s", dbURL))
		return errors.New("unable to connect to the database")
	}

	dbOK := false
	// Check connection, retry 10 times.
	for retry := 0; retry < 10; retry++ {
		err = db.Ping()
		if err != nil {
			logger.Error(fmt.Sprintf("Cannot reach database. Waiting 5 seconds before retrying. Try: %d", retry+1))
			time.Sleep(5 * time.Second)
			continue
		}
		dbOK = true
		break
	}

	if !dbOK {
		logger.Debug(fmt.Sprintf("Connection URL: %s", dbURL))
		return errors.New("unable to open database connection")
	}

	db.SetMaxIdleConns(0)
	// Bootstrap the database schema
	for _, t := range types {
		_, err = db.Exec(getCreateTableQuery(t))
		if err != nil {
			logger.Error(err)
			return fmt.Errorf("failed to create table for: %s", reflect.TypeOf(t).Name())
		}
	}

	logger.Log("Database configured")
	return nil
}

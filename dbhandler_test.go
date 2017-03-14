package gopsql

import "testing"

type testType struct {
	ID    int    `column_type:"SERIAL primary key" column_order_by:"yes"`
	Alive bool   `column_type:"bool"`
	Title string `column_type:"varchar(255)"`
}

func (t testType) GetID() int {
	return t.ID
}
func (t testType) SetID(id int) {
	t.ID = id
}

func TestCreateTableQueryCreation(t *testing.T) {
	t.Log(getCreateTableQuery(testType{}))
}

func TestSelectAllQueryCreation(t *testing.T) {
	t.Log(getSelectAllQuery(testType{}))
}

func TestParseFilterString(t *testing.T) {
	res, err := parseFilterString("key1", "value1", "AND", "key2", "value2")
	t.Log(res)
	if err != nil {
		t.Error(err)
	}
}

func TestSelectFilteredQueryCreation(t *testing.T) {
	filter, _ := parseFilterString("Alive", "t")

	res := getSelectFilteredQuery(testType{}, filter)
	t.Log(res)
}

func TestInsertQueryCreation(t *testing.T) {
	tt := testType{0, true, "Hello"}
	t.Log(getInsertQuery(tt))
}

func TestUpdateQueryCreation(t *testing.T) {
	tt := testType{13, true, "Hello"}
	t.Log(getUpdateQuery(tt))
}

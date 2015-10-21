package fakes

import (
	"database/sql"
	"database/sql/driver"
)

type FakeDB struct {
	BeginTxReturn *sql.Tx
	BeginError    error

	CloseError error

	PrepareQueryParam string
	PrepareStmtReturn *sql.Stmt
	PrepareError      error

	DriverDriverReturn driver.Driver

	ExecQueryParam   string
	ExecArgsParam    []interface{}
	ExecResultReturn sql.Result
	ExecError        error

	PingError error

	QueryQueryParam string
	QueryArgsParam  []interface{}
	QueryRowsReturn *sql.Rows
	QueryError      error

	QueryRowQueryParam string
	QueryRowArgsParam  []interface{}
	QueryRowRowReturn  *sql.Row

	SetMaxIdleConnsNParam int
	SetMaxOpenConnsNParam int
}

var (
	OpenDB    *sql.DB
	OpenError error
)

func NewFakeDB() *FakeDB {
	return &FakeDB{}
}

func Open(driverName, dataSourceName string) (*sql.DB, error) {
	return OpenDB, OpenError
}

func (fakeDb *FakeDB) Begin() (*sql.Tx, error) {
	return fakeDb.BeginTxReturn, fakeDb.BeginError
}

func (fakeDb *FakeDB) Close() error {
	return fakeDb.CloseError
}

func (fakeDb *FakeDB) Driver() driver.Driver {
	return fakeDb.DriverDriverReturn
}

func (fakeDb *FakeDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	fakeDb.ExecQueryParam = query
	fakeDb.ExecArgsParam = args

	return fakeDb.ExecResultReturn, fakeDb.ExecError
}

func (fakeDb *FakeDB) Ping() error {
	return fakeDb.PingError
}

func (fakeDb *FakeDB) Prepare(query string) (*sql.Stmt, error) {
	fakeDb.PrepareQueryParam = query
	return fakeDb.PrepareStmtReturn, fakeDb.PrepareError
}

func (fakeDb *FakeDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	fakeDb.QueryQueryParam = query
	fakeDb.QueryArgsParam = args

	return fakeDb.QueryRowsReturn, fakeDb.QueryError
}

func (fakeDb *FakeDB) QueryRow(query string, args ...interface{}) *sql.Row {
	fakeDb.QueryRowQueryParam = query
	fakeDb.QueryRowArgsParam = args

	return fakeDb.QueryRowRowReturn
}

func (fakeDb *FakeDB) SetMaxIdleConns(n int) {
	fakeDb.SetMaxIdleConnsNParam = n
}
func (fakeDb *FakeDB) SetMaxOpenConns(n int) {
	fakeDb.SetMaxOpenConnsNParam = n
}

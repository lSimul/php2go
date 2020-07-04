package std

import (
	"time"

	// TODO: It does not have to be always MySQL.
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var _ Bool = (*SQL)(nil)

// SQL wraps MySQL connection. The main
// purpose is to behave the same way as
// PHP SQL connection does.
type SQL struct {
	server   string
	user     string
	password string

	table string

	db *sqlx.DB

	err error
}

// NewSQL creates struct SQL with defined server,
// user, and its password.
// Function call is the same like mysqli_connect().
func NewSQL(server, user, password string) *SQL {
	return &SQL{
		server:   server,
		user:     user,
		password: password,
	}
}

// SelectDB adds information about table.
// With this information the connection is
// really opened.
// It is opened as "Unsafe" connection. This
// hides issues with not reading every column
// in `SELECT *` etc.
// It is the alias for mysqli_select_db(),
// where connection is the struct, not the
// first argument.
func (s *SQL) SelectDB(table string) {
	s.table = table
	s.db, s.err = sqlx.Open("mysql", s.user+":"+s.password+"@tcp("+s.server+")/"+s.table)
	s.db = s.db.Unsafe()
	s.db.DB.SetConnMaxLifetime(time.Second)
}

// Query performs SQL query defined in q.
// It returns std.Rows.
// It is the alias for mysqli_query(),
// where connection is the struct, not the
// first argument.
func (s *SQL) Query(q string) *Rows {
	if s.err != nil {
		panic(s.err)
	}

	var rows *sqlx.Rows
	rows, s.err = s.db.Queryx(q)
	return &Rows{rows}
}

// Close is another wrapper for connection,
// but this one is not present in PHP. Closes
// SQL connection.
func (s *SQL) Close() {
	s.db.Close()
}

// ToBool implements inteface Bool,
// the simplest way how to check
// if the connection is valid.
func (s SQL) ToBool() bool {
	return s.err == nil
}

// Rows wraps *sqlx.Rows, the main goal
// is to rename a method to make it more
// clear what it does, and hide SQL package
// in the transpiled script.
type Rows struct {
	rows *sqlx.Rows
}

// Next is one of the methods use to implement
// mysqli_fetch_array(). Instead of returning
// struct or a false, which represents if the
// result is empty, Next is used only for this
// flag "is empty".
func (r *Rows) Next() bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

// Scan fills passed pointer to struct t
// with values found in the rows. It is the
// second implementation of mysqli_fetch_array().
// In PHP it has two arguments, first one is here
// represented by Rows, the second one defines
// returned value type. Here I support only
// MYSQLI_ASSOC, which is here implemented as struct.
// This is probably the only way how to make
// strictly typed with the current set of supported
// operations in the transpiler.
func (r *Rows) Scan(t interface{}) {
	r.rows.StructScan(t)
}

package std

import (
	// TODO: It does not have to be always MySQL.
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var _ Bool = (*SQL)(nil)

type SQL struct {
	server   string
	user     string
	password string

	table string

	db *sqlx.DB

	err error
}

func NewSQL(server, user, password string) *SQL {
	return &SQL{
		server:   server,
		user:     user,
		password: password,
	}
}

func (s *SQL) SelectDB(table string) {
	s.table = table
	s.db, s.err = sqlx.Open("mysql", s.user+":"+s.password+"@tcp("+s.server+")/"+s.table)
	s.db = s.db.Unsafe()
	s.db.DB.SetConnMaxLifetime(time.Second)
}

func (s *SQL) Query(q string) *Rows {
	if s.err != nil {
		panic(s.err)
	}

	var rows *sqlx.Rows
	rows, s.err = s.db.Queryx(q)
	return &Rows{rows}
}

func (s *SQL) Close() {
	s.db.Close()
}

func (s SQL) ToBool() bool {
	return s.err == nil
}

type Rows struct {
	rows *sqlx.Rows
}

func (r *Rows) Next() bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

func (r *Rows) Scan(t interface{}) {
	r.rows.StructScan(t)
}

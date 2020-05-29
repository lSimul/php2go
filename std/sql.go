package std

import (
	"database/sql"

	// TODO: It does not have to be always MySQL.
	_ "github.com/go-sql-driver/mysql"
)

type SQL struct {
	server   string
	user     string
	password string

	table string

	db *sql.DB

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
	s.db, s.err = sql.Open("mysql", s.user+":"+s.password+"(tcp"+s.server+")/"+s.table)
}

func (s *SQL) Query(q string) {
	if s.err != nil {
		panic(s.err)
	}
	_, s.err = s.db.Query(q)
}

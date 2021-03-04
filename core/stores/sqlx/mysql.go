package sqlx

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
)

const (
	mysqlDriverName           = "mysql"
	duplicateEntryCode uint16 = 1062
)

type DBConfig struct {
	Host     string
	Port     int
	UserName string
	Passwd   string
	DBName   string
}

func (db *DBConfig) toString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.UserName, db.Passwd, db.Host, db.Port, db.DBName)
}

// NewMysql returns a mysql connection.
func NewMysql(datasource DBConfig, opts ...SqlOption) SqlConn {
	opts = append(opts, withMysqlAcceptable())
	return NewSqlConn(mysqlDriverName, datasource.toString(), opts...)
}

func mysqlAcceptable(err error) bool {
	if err == nil {
		return true
	}

	myerr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	switch myerr.Number {
	case duplicateEntryCode:
		return true
	default:
		return false
	}
}

func withMysqlAcceptable() SqlOption {
	return func(conn *commonSqlConn) {
		conn.accept = mysqlAcceptable
	}
}

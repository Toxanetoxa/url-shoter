package types

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

//type Storage struct {
//	Db *sql.DB
//}

//type QUERYRow interface {
//	QueryRow(query string, args ...interface{}) Row
//	Exec(query string, args ...interface{}) (Result, error)
//}
//
//type Row interface {
//	Scan(dest ...interface{}) error
//}
//
//type Result interface {
//	LastInsertId() (int64, error)
//	RowsAffected() (int64, error)
//}

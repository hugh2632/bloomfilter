package bloomfilter

import (
	"database/sql"
	"hash"
)

type SqlFilter struct {
	filter
	id         string
	db *sql.DB
}

//Mysql默认存储bloom值的库
type Bloom struct {
	Id  int
	Val string
}

func (r *SqlFilter) Write() error {
	rows, err := r.db.Query("select * from bloom where id='" + r.id + "'")
	if err != nil {
		return err
	}
	if rows.Next() {
		_, err = r.db.Exec("update bloom set val='" + string(r.Bytes) + "' where id=" + r.id)
		if err != nil {
			return err
		}
	} else {
		_, err = r.db.Exec("insert into bloom(Id, Val) values (" + r.id + ",'" + string(r.Bytes) + "');")
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SqlFilter) Close() error{
	return r.db.Close()
}

func NewSqlFilter(id string, byteLen int, db *sql.DB, hashes ...hash.Hash64) (res *SqlFilter, err error) {
	rows, err := db.Query("select id, val from bloom where id='" + id + "'")
	if err != nil {
		return nil, err
	}
	res = &SqlFilter{
		filter: filter{
			Bytes:  make([]byte, byteLen,byteLen),
			Hashes: hashes,
		},
		id: id,
		db:db,
	}
	if rows.Next() {
		var bl Bloom
		err = rows.Scan(&bl.Id, &bl.Val)
		if err == nil {
			var bytes = []byte(bl.Val)
			if len(bytes) == byteLen {
				res.Bytes = bytes
			}
		} else {
			return nil, err
		}
	}
	return res, err
}



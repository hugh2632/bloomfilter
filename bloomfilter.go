package bloomfilter

import (
	"database/sql"
	"github.com/go-redis/redis"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"log"
	"math"
)

var DefaultHash = []hash.Hash64{fnv.New64(), crc64.New( crc64.MakeTable(crc64.ISO))}

type filter struct {
	Bytes  []byte
	Hashes []hash.Hash64
}

func (f *filter) Push(str []byte) {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			log.Println(err.Error())
		}
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)
		var yBit = res & 7
		var now = f.Bytes[yByte] | 1 << yBit
		if now != f.Bytes[yByte] {
			f.Bytes[yByte] = now
		}

	}
}

func (f *filter) Exists(str []byte) bool {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			log.Println(err.Error())
		}
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)
		var yBit = res & 7
		if f.Bytes[yByte]|1<<yBit != f.Bytes[yByte] {
			return false
		}
	}
	return true
}

func (f *filter) IsEmpty() bool {
	for i, _ := range f.Bytes {
		if f.Bytes[i] != 0 {
			return false
		}
	}
	return true
}

func GetFlasePositiveRate(m int, n int, k int) float64 {
	return math.Pow(1-math.Pow(1-1/float64(m), float64(k)*float64(n)), float64(k))
}

type RedisFilter struct{
	filter
	address string
	password string
	db	int
	key string
}

func (r *RedisFilter) Write() error{
	var cli = redis.NewClient(&redis.Options{
		Addr:     r.address,
		Password: r.password, // no password set
		DB:       r.db,  // use default DB
	})
	err := cli.Do("HSET", r.key, "Bytes",r.Bytes).Err()
	if err != nil {
		return err
	}
	return cli.Close()
}

func (r *RedisFilter) Close() error{
	return nil
}

func NewRedisFilter(key string, byteLen int, redisAddr string, psd string, db int,  hashes ...hash.Hash64) (res *RedisFilter, err error){
	res = &RedisFilter{
		filter:filter{
			Bytes: make([]byte, byteLen),
			Hashes: hashes,
		},
		address:     redisAddr,
		password: psd, // no password set
		db:       db,  // use default DB
		key: key,
	}
	var cli = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: psd, // no password set
		DB:       db,  // use default DB
	})
	_, err = cli.Ping().Result()
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	var cmd = cli.Do("HGET", key, "Bytes")
	if err = cmd.Err(); err != nil {
		if err.Error() == "redis: nil"{
			cli.Do("HSET", key, "Bytes", res.Bytes)
		}else {
			return nil, err
		}
	}else{
		var val = cmd.Val()
		if val != nil {
			var bytes = []byte(val.(string))
			if len(bytes) == byteLen{
				res.Bytes = bytes
			}
		}
	}
	return res, nil
}

type MysqlFilter struct {
	filter
	datasource string
	id         string
}

//Mysql默认存储bloom值的库
type Bloom struct {
	Id  int
	Val string
}

func (r *MysqlFilter) Write() error {
	db, err := sql.Open("mysql", r.datasource)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query("select * from bloom where id='" + r.id + "'")
	if err != nil {
		return err
	}
	if rows.Next() {
		_, err = db.Exec("update bloom set val='" + string(r.Bytes) + "' where id=" + r.id)
		if err != nil {
			return err
		}
	} else {
		_, err = db.Exec("insert into bloom(Id, Val) values (" + r.id + ",'" + string(r.Bytes) + "');")
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MysqlFilter) Close() error{
	return nil
}

func NewSqlFilter(id string, byteLen int, datasource string, hashes ...hash.Hash64) (res *MysqlFilter, err error) {
	db, err := sql.Open("mysql", datasource)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("select id, val from bloom where id='" + id + "'")
	if err != nil {
		return nil, err
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
	return &MysqlFilter{
		filter: filter{
			Bytes:  make([]byte, byteLen),
			Hashes: hashes,
		},
		id: id,
	}, err
}


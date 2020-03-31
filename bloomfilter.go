package bloomfilter

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"log"
	"math"
	"runtime/debug"
	"strconv"
)

var DefaultHash = []hash.Hash64{fnv.New64(), crc64.New( crc64.MakeTable(crc64.ISO))}

type filter struct {
	Bytes  []byte
	Hashes []hash.Hash64
	AlreadyExistCount int
}

func (f *filter) Push(str []byte) {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		v.Write(str)
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)
		var yBit = res & 7
		//todo 遇到大端模式CPU可能会出现 BUG
		var now = f.Bytes[yByte] | 1 << yBit
		if now != f.Bytes[yByte] {
			f.AlreadyExistCount ++
			f.Bytes[yByte] = now
		}

	}
}

func (f *filter) Exists(str []byte) bool {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		v.Write(str)
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)
		var yBit = res & 7
		//todo 遇到大端模式CPU可能会出现 BUG
		if f.Bytes[yByte]|1<<yBit != f.Bytes[yByte] {
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
	cli redis.Client
	key string
}

func (r *RedisFilter) Write(){
	r.cli.Do("HSET", r.key, "Bytes",r.Bytes, "AlreadyCount", r.AlreadyExistCount )
}

func (r *RedisFilter) Close(){
	r.cli.Do("HSET", r.key, "Bytes",r.Bytes, "AlreadyCount", r.AlreadyExistCount )
	r.cli.Close()
}

func NewRedisFilter(key string, byteLen int, redisAddr string, psd string, db int,  hashes ...hash.Hash64) *RedisFilter{
	var res RedisFilter
	res.filter = filter{
		Bytes: make([]byte, byteLen),
		Hashes: hashes,
	}
	res.cli = *redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: psd, // no password set
		DB:       db,  // use default DB
	})
	res.key = key
	_, err := res.cli.Ping().Result()
	if err != nil {
		panic(err)
	}
	var cmd = res.cli.Do("HGET", key, "Bytes")
	var val = cmd.Val()
	if val != nil {
		var bytes = []byte(val.(string))
		if len(bytes) == byteLen{
			res.filter.Bytes = bytes
		}
	}
	var AlreadyCountcmd = res.cli.Do("HGET", key, "AlreadyCount")
	var alreadyVal = AlreadyCountcmd.Val()
	if alreadyVal != nil {
		res.AlreadyExistCount, err = strconv.Atoi(alreadyVal.(string))
	}
	return &res
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

func (r *MysqlFilter) Write() {
	_ = newMysql(r.datasource, func(db *sql.DB) {
		rows, err := db.Query("select * from bloom where id='" + r.id + "'")
		if err != nil {
			log.Fatal(err)
		}
		if rows.Next() {
			_, err = db.Exec("update bloom set val='" + string(r.Bytes) + "' where id=" + r.id)
			if err != nil {
				log.Println("更新bloom失败")
			}
		} else {
			_, err = db.Exec("insert into bloom(Id, Val) values (" + r.id + ",'" + string(r.Bytes) + "');")
			if err != nil {
				log.Println("插入bloom失败" + err.Error())
			}
		}
	})
}

func NewSqlFilter(id string, byteLen int, datasource string, hashes ...hash.Hash64) *MysqlFilter {
	var res MysqlFilter
	res.filter = filter{
		Bytes:  make([]byte, byteLen),
		Hashes: hashes,
	}
	res.datasource = datasource
	res.id = id
	_ = newMysql(datasource, func(db *sql.DB) {
		rows, err := db.Query("select id, val from bloom where id='" + id + "'")
		if err != nil {
			log.Fatal(err)
		}
		if rows.Next() {
			var bl Bloom
			err = rows.Scan(&bl.Id, &bl.Val)
			if err == nil {
				var bytes = []byte(bl.Val)
				if len(bytes) == byteLen {
					res.filter.Bytes = bytes
				}
			} else {
				log.Println(err.Error())
			}
		}
	})
	return &res
}

func newMysql(datasource string, f func(*sql.DB)) (err error) {
	if p := recover(); p != nil {
		str, ok := p.(string)
		if ok {
			err = errors.New(str)
			log.Println(str)
			fmt.Println(str)
		} else {
			err = errors.New("panic")
		}
		debug.PrintStack()
	}
	db, err := sql.Open("mysql", datasource)
	if err != nil {
		panic("打开数据库失败," + err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	f(db)
	return err
}
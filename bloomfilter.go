package bloomfilter

import (
	"github.com/go-redis/redis"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"math"
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
	*redis.Client
	key string
}

func (r *RedisFilter) Write(){
	r.Client.Do("HSET", r.key, "Bytes",r.Bytes, "AlreadyCount", r.AlreadyExistCount )
}

func NewRedisFilter(key string, byteLen int, redisAddr string, psd string, db int,  hashes ...hash.Hash64) RedisFilter{
	var res RedisFilter
	res.filter = filter{
		Bytes: make([]byte, byteLen),
		Hashes: hashes,
	}
	res.Client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: psd, // no password set
		DB:       db,  // use default DB
	})
	res.key = key
	_, err := res.Client.Ping().Result()
	if err != nil {
		panic(err)
	}
	var cmd = res.Client.Do("HGET", key, "Bytes")
	var val = cmd.Val()
	if val != nil {
		var bytes = []byte(val.(string))
		if len(bytes) == byteLen{
			res.filter.Bytes = bytes
		}
	}
	var AlreadyCountcmd = res.Client.Do("HGET", key, "AlreadyCount")
	var alreadyVal = AlreadyCountcmd.Val()
	if alreadyVal != nil {
		res.AlreadyExistCount, err = strconv.Atoi(alreadyVal.(string))
	}
	return res
}
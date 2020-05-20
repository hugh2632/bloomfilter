package bloomfilter

import (
	"github.com/go-redis/redis"
	"hash"
)

type RedisFilter struct{
	filter
	key string
	cli *redis.Client
}

func (r *RedisFilter) Write() error{
	return r.cli.Do("HSET", r.key, "Bytes",r.Bytes).Err()
}

func (r *RedisFilter) Close() error{
	return r.cli.Close()
}

func NewRedisFilter(cli *redis.Client, key string, byteLen int,  hashes ...hash.Hash64) (res *RedisFilter, err error){
	_, err = cli.Ping().Result()
	if err != nil {
		return nil, err
	}
	res = &RedisFilter{
		filter:filter{
			Bytes: make([]byte, byteLen, byteLen),
			Hashes: hashes,
		},
		key:key,
		cli:cli,
	}
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

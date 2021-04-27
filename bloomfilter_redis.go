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
	return r.cli.HSet("bloom", r.key, r.Bytes).Err()
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
	var cmd = cli.HGet("bloom", key)
	if err = cmd.Err(); err != nil {
		if err.Error() == "redis: nil"{
			cli.HSet("bloom", key, res.Bytes)
		}else {
			return nil, err
		}
	}else{
		var val = cmd.Val()
		var bytes = []byte(val)
		if len(bytes) == byteLen{
			res.Bytes = bytes
		}
	}
	return res, nil
}

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hugh2632/bloomfilter/global"
	"github.com/hugh2632/bloomfilter/memory"
)

// Memory based BloomFilter.
// Only synchronize the bytes to the redis BITMAP when 'Write' method called.
// 基于内存的过滤器
// 只有在使用'Write'方法时才会将本地的字节数组同步到Redis服务器。
type CachedFilter struct {
	*memory.Filter
	hashTableName    string
	key     string
	cli     *redis.Client
	ctx context.Context
}

// Initial the local bytes from the BITMAP on redis server.
// 同步Redis服务器上的BITMAP到本地的字节数组。
func (f *CachedFilter) Init(context context.Context, cli *redis.Client, hashTableName string, key string) error {
	f.cli = cli
	f.hashTableName = hashTableName
	f.key = key
	f.ctx = context
	var cmd = f.cli.HGet(f.ctx, f.hashTableName, f.key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			f.cli.HSet(context, f.hashTableName, f.key, f.Bytes)
		} else {
			return err
		}
	} else {
		var bytes, _ = cmd.Bytes()
		if len(bytes) == len(f.Bytes) {
			f.Bytes = bytes
		} else {
			return global.ErrUnMatchLength
		}
	}
	return nil
}

// Write to the redis server only if the bytes have been changed. This may takes seconds if the byte amount is large.
// 只有在字节数组被改变过的情况下才会同步到服务器。如果字节数很多，会比较费时。
func (f *CachedFilter) Write() error {
	if f.IsChanged {
		cmd := f.cli.HSet(f.ctx, f.hashTableName, f.key, f.Bytes)
		if cmd.Err() != nil {
			return cmd.Err()
		}
		f.IsChanged = false
	}
	return nil
}

// Close the redis client.
func (f *CachedFilter) Close() error {
	return f.cli.Close()
}

package bloomfilter

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/hugh2632/bloomfilter/adapter"
	"github.com/hugh2632/bloomfilter/global"
	"github.com/hugh2632/bloomfilter/memory"
	rf "github.com/hugh2632/bloomfilter/redis"
	sq "github.com/hugh2632/bloomfilter/sql"
	"gorm.io/gorm"
	"hash"
	"hash/fnv"
	"strings"
)

var (
	DefaultHash = []global.HashFunc{
		func() hash.Hash64{return fnv.New64()},
		//func() hash.Hash64{return crc64.New(crc64.MakeTable(crc64.ISO))}, //remove this hash for it is not uniformly distributed. 移除这个哈希函数，因为它并不均匀分布。
		func() hash.Hash64{return adapter.NewMD5()},
		func() hash.Hash64{return adapter.NewSha1()},
		func() hash.Hash64{return adapter.NewSha256()},
		func() hash.Hash64{return adapter.NewSha512()},
	}
	IsDebug = false
)

type RedisFilterType uint32

const (
	// 缓存模式
	RedisFilterType_Cached RedisFilterType = 0
	// 交互模式
	RedisFilterType_Interactive RedisFilterType = 1
)

// 增量过滤器
type IFilter interface {
	//添加内容到过滤器中
	Push(content []byte)
	//将数据持久化
	Write() error
	//判断内容是否存在
	Exists(content []byte) bool
	//判断是否为空
	IsEmpty() bool
	//清除数据
	Clear() error
	//关闭
	Close() error
}

// Memory Filter
func NewMemoryFilter(bytes []byte, hashes ...global.HashFunc) IFilter {
	return &memory.Filter{
		Bytes:     bytes,
		Hashes:    hashes,
		IsChanged: false,
	}
}

// RedisFilter.
// CachedFilter's data is stored in hash map. And you need use "Write" method to truly write the bytes to redis server.
// InteractiveFilter will commit the contents to redis server whenever 'Push' function is called.
// 数据保存到redis的过滤器。
// 如果选择CachedFilter, 其实是在本地维护了一个MemoryFilter, 只有在使用Write方法时才会真正提交到Redis服务器。
// 如果选择InteractiveFilter, 将在每次Push数据的时候，直接提交到redis服务器。(集群中需要使用分布式锁)
func NewRedisFilter(context context.Context, cli *redis.Client, tp RedisFilterType, key string, byteLen uint64, hashes ...global.HashFunc) (res IFilter, err error) {
	switch tp {
	case RedisFilterType_Cached:
		f := &rf.CachedFilter{
			Filter: &memory.Filter{
				Bytes:  make([]byte, byteLen, byteLen),
				Hashes: hashes,
			},
		}
		err := f.Init(context, cli, "bloom", key) // The default hashtable name is 'bloom'
		if err != nil {
			return nil, err
		}
		return f, nil
	case RedisFilterType_Interactive:
		f := &rf.InteractiveFilter{
			Key:     key,
			Cli:     cli,
			Context: context,
			ByteLen: byteLen,
			Hashes:  hashes,
		}
		return f, nil
	default:
		return nil, errors.New("不匹配的过滤器类型")
	}
}

// MemoryBased bloom filter. Synchronize to the sql database when 'Write' method is called. The default tableName is 'bloom'
// 基于内存的过滤器，在使用Write方法时，将同步到SQL数据库中。默认的表名为'bloom'
func SqlFilter(db *gorm.DB, key string, byteLen uint64, hashes ...global.HashFunc) (IFilter, error) {
	f := &sq.SQLFilter{
		Filter: memory.Filter{
			Bytes:     make([]byte, byteLen, byteLen),
			Hashes:    hashes,
			IsChanged: false,
		},
	}
	err := f.Init(db, "bloom", key, f.Bytes)
	if err != nil {
		// TODO:
		// Deal with 'Table xxx.bloom doesn't exist'. More common method may required other than mysql database.
		// 处理表格不存在的情况，除mysql以外的此种错误,可能需要更通用的判断方法.
		if strings.Contains(err.Error(), "Table") && strings.Contains(err.Error(), "exist") {
			global.Logger.Printf("Try to create table:'%s'\n", "bloom")
			e := f.CreateTable()
			if e != nil {
				return nil, e
			}
			return SqlFilter(db, key, byteLen, hashes...)
		}
		return nil, err
	}
	return f, nil
}

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hugh2632/bloomfilter/global"
	"hash"
)

// Interactive mode. Push/Exists/IsEmpty methods will interactive with the BITMAP on the redis server.
// 交互模式。使用Push/Exists/IsEmpty方法将直接操作在REDIS服务器上的BITMAP，本地不缓存数据。
type InteractiveFilter struct {
	Key     string
	Cli     *redis.Client
	Context context.Context
	ByteLen uint64
	Hashes  []hash.Hash64
}

// Just delete the key.
// 直接删除键值
func (f *InteractiveFilter) Clear() error {
	return f.Cli.Del(f.Context, f.Key).Err()
}

// update the bloom table immediately by redis pipeline.
// 使用Redis pipeline更新bloom表格。
func (f *InteractiveFilter) Push(str []byte) {
	var offsets []int64
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			global.Logger.Println(err.Error())
		}
		var res = v.Sum64()
		// For the SetBIT and GetBIT method operate the Bitmap by order. We should reverse each byte to keep up with the Cached Filter.
		// 因为redis的SetBIT和GetBIT都是按顺序读写的，所以要将翻转字节，以使它保存的结构和cached过滤器一致。0:7, 1:6, 2:5, 3:4, 4:3... 7:0
		offsets = append(offsets, int64((res%f.ByteLen)*8+7-(res/f.ByteLen)&7))
	}
	cmds, err := f.Cli.Pipelined(f.Context, func(pipeliner redis.Pipeliner) error {
		for _, v := range offsets {
			pipeliner.SetBit(f.Context, f.Key, v, 1)
		}
		return nil
	})
	if err != nil {
		global.Logger.Println(err)
	}
	for _, e := range cmds {
		if e.Err() != nil {
			global.Logger.Println(err)
		}
	}
}

// should not use this method when use interactive mode.
// 交互模式不应使用Write方法
func (f *InteractiveFilter) Write() error {
	global.Logger.Println("不应使用的方法")
	return nil
}

func (f *InteractiveFilter) Exists(str []byte) bool {
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			global.Logger.Println(err.Error())
		}
		var res = v.Sum64()
		offset := int64((res%f.ByteLen)*8 + 7 - (res/f.ByteLen)&7)
		cmd := f.Cli.GetBit(f.Context, f.Key, offset)
		if cmd.Err() == nil && cmd.Val() == 0 {
			return false
		}
	}
	return true
}

func (f *InteractiveFilter) IsEmpty() bool {
	cmd := f.Cli.BitCount(f.Context, f.Key, &redis.BitCount{
		Start: 0,
		End:   int64(f.ByteLen - 1),
	})
	if cmd.Err() == nil && cmd.Val() == 0 {
		return true
	}
	return false
}

func (f *InteractiveFilter) Close() error {
	return f.Cli.Close()
}

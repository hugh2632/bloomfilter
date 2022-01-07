package memory

import (
	"github.com/hugh2632/bloomfilter/global"
)

// Memory base filter. All operations are in the memory.
// 基于内存的过滤器。所有的操作都在内存中进行。
type Filter struct {
	Bytes     []byte
	Hashes    []global.HashFunc
	IsChanged bool
}

// Reset all the bytes to zero.
// 将所有字节重置为0
func (f *Filter) Clear() error {
	for i, _ := range f.Bytes {
		f.Bytes[i] = 0
	}
	return nil
}

func (f *Filter) Push(content []byte) {
	var byteLen = uint64(len(f.Bytes))
	if byteLen < 1 {
		global.Logger.Println(global.ErrEmptyContent)
		return
	}
	for _, h := range f.Hashes {
		v := h()
		v.Reset()
		_, err := v.Write(content)
		if err != nil {
			global.Logger.Println(err.Error())
		}
		var res = v.Sum64()
		// Get the byte.
		var yByte = res % byteLen
		// Get the bit position in the byte.
		var yBit = (res / byteLen) & 7
		var now = f.Bytes[yByte] | 1<<yBit
		if now != f.Bytes[yByte] {
			f.Bytes[yByte] = now
			f.IsChanged = true
		}
	}
}

func (f *Filter) Write() error {
	f.IsChanged = false
	return nil
}

func (f *Filter) Exists(content []byte) bool {
	var byteLen = uint64(len(f.Bytes))
	for _, h := range f.Hashes {
		v := h()
		v.Reset()
		_, err := v.Write(content)
		if err != nil {
			global.Logger.Println(err.Error())
		}
		var res = v.Sum64()
		var yByte = res % byteLen
		var yBit = (res / byteLen) & 7
		if f.Bytes[yByte]|1<<yBit != f.Bytes[yByte] {
			return false
		}
	}
	return true
}

func (f *Filter) IsEmpty() bool {
	for i, _ := range f.Bytes {
		if f.Bytes[i] != 0 {
			return false
		}
	}
	return true
}

func (f *Filter) Close() error {
	f.IsChanged = false
	return nil
}

package bloomfilter

import (
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


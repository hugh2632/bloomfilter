package adapter

import (
	"crypto/md5"
	"hash"
)

type MD5Adapter struct{
	hash.Hash
}

func (h MD5Adapter) Sum64() uint64 {
	b := h.Sum(nil)
	return Bytes16ToUint64(b)
}

func NewMD5() hash.Hash64{
	m := &MD5Adapter{
		md5.New(),
	}
	return m
}

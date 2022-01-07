package adapter

import (
	"crypto/sha1"
	"hash"
)

type Sha1Adapter struct{
	hash.Hash
}

func (h Sha1Adapter) Sum64() uint64 {
	b := h.Sum(nil)
	return Bytes16ToUint64(b)
}

func NewSha1() hash.Hash64{
	m := &Sha1Adapter{
		sha1.New(),
	}
	return m
}


package adapter

import (
	"crypto/sha256"
	"hash"
)

type Sha256Adapter struct{
	hash.Hash
}

func (h Sha256Adapter) Sum64() uint64 {
	b := h.Sum(nil)
	return Bytes16ToUint64(b)
}

func NewSha256() hash.Hash64{
	m := &Sha256Adapter{
		sha256.New(),
	}
	return m
}

package adapter

import (
	"crypto/sha512"
	"hash"
)

type Sha512Adapter struct{
	hash.Hash
}

func (h Sha512Adapter) Sum64() uint64 {
	b := h.Sum(nil)
	return Bytes16ToUint64(b)
}

func NewSha512() hash.Hash64{
	m := &Sha512Adapter{
		sha512.New(),
	}
	return m
}

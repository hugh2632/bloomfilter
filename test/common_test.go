package test

import (
	"github.com/hugh2632/bloomfilter"
	"hash"
	"hash/crc64"
	"testing"
)

func TestIsHashFuncUniformlyDistributed(t *testing.T) {
	for _, f := range bloomfilter.DefaultHash{
		t.Log(bloomfilter.CalculateInformationEntropy(f))
	}
	//不满足的hash，已舍弃
	t.Log("abandoned hash function:")
	t.Log(bloomfilter.CalculateInformationEntropy(func() hash.Hash64{return crc64.New(crc64.MakeTable(crc64.ISO))}))
}

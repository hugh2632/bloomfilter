package test

import (
	"fmt"
	"github.com/hugh2632/bloomfilter"
	"github.com/hugh2632/bloomfilter/memory"
	"strconv"
	"testing"
)

func TestMemFalsePositiveRate(t *testing.T) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...).(*memory.Filter)
	testFalsePositiveRate(t, memFilter, 10240 * 8, 1000, 3, 100000000)
	//=== RUN   TestMemFalsePositiveRate
	//    common.go:37: 理论误判率 Theoretical false positive rate: 0.0018230817954481005
	//    common.go:49: 实际误判率 Real false positive rate:0.00047424474244742447
	//--- PASS: TestMemFalsePositiveRate (7.43s)
	//PASS
}

func TestMemoryFilter(t *testing.T) {
	// Initial a memory filter
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...)
	// Push 2000-3000 numbers to the filter.
	// 把2000-3000的数字压入过滤器
	fillNums(memFilter, 2000, 3000)
	// Check whether 2500-3000 and 3001-3500 exist in the filter or not.
	// 查看2500-3000，以及3001-3500是否存在于过滤器中
	for i := 2500; i < 3500; i++ {
		t.Logf("%d, %t", i, memFilter.Exists([]byte(strconv.Itoa(i))))
	}
}

func BenchmarkMemoryFilter(b *testing.B) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...).(*memory.Filter)
	fmt.Println("----------------------")
	fillNums(memFilter, 250, 300)
	for i := 0; i < b.N; i++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if memFilter.Exists([]byte(strconv.Itoa(i))) {
					b.Log(i)
				}
			}
		})
	}
}

func TestMemoryClear(t *testing.T) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...).(*memory.Filter)
	for i:=0;i<5;i++{
		t.Log("origin:", memFilter.Exists([]byte(strconv.Itoa(250))))
		fillNums(memFilter, 250, 250)
		t.Log("filled:", memFilter.Exists([]byte(strconv.Itoa(250))))
		memFilter.Clear()
		t.Log("cleared:", memFilter.Exists([]byte(strconv.Itoa(250))))
	}
}

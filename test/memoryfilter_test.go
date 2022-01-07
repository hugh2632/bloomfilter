package test

import (
	"fmt"
	"github.com/hugh2632/bloomfilter"
	"github.com/hugh2632/bloomfilter/memory"
	"strconv"
	"testing"
)

func TestMemFalsePositiveRate(t *testing.T) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash[:3]...).(*memory.Filter)
	testFalsePositiveRate(t, memFilter, 10240 * 8, 1000, 3, 100000000)
	//=== RUN   TestMemFalsePositiveRate
	//    common.go:37: 理论误判率 Theoretical false positive rate: 0.0018230817954481005
	//    common.go:49: 实际误判率 Real false positive rate:0.00047424474244742447
	//--- PASS: TestMemFalsePositiveRate (7.43s)
	//PASS
}

func TestMemoryFilter(t *testing.T) {
	// Initial a memory filter
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash[:5]...)
	// Push 0-9999 numbers to the filter.
	// 0-9999
	fillNums(memFilter, 0, 9999)
	// Check whether 10000-1000000 exist in the filter or not.
	// 查看10000-1000000 是否存在于过滤器中
	var counter = 0
	for i := 10000; i < 1000000; i++ {
		if  memFilter.Exists([]byte(strconv.Itoa(i))){
			counter++
			t.Log("hashed:", i)
		}
		//t.Logf("%d, %t", i, memFilter.Exists([]byte(strconv.Itoa(i))))
	}
	t.Log("碰撞个数：",counter)
}

func BenchmarkMemoryFilter(b *testing.B) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...).(*memory.Filter)
	fmt.Println("----------------------")
	fillNums(memFilter, 250, 300)
	var counter = 0
	for i := 0; i < b.N; i++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if memFilter.Exists([]byte(strconv.Itoa(i))) {
					//b.Log(i)
					counter++
				}
			}
		})
	}
	b.Log(counter)
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

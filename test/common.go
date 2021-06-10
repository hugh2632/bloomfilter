package test

import (
	"fmt"
	"github.com/hugh2632/bloomfilter"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// fill some nums into f
func fillNums(filter bloomfilter.IFilter, begin int, end int) {
	for i := begin; i < end+1; i++ {
		filter.Push([]byte(strconv.Itoa(i)))
	}
	fmt.Printf("已填入%d-%d的数据\n", begin, end)
}

func getRandomNums(top int, quantity int) map[int]struct{}{
	rand.Seed(time.Now().Unix())
	var m = make(map[int]struct{})
	for i:=0;i<quantity;i++{
		r := rand.Intn(top)
		_, exists := m[r]
		for exists{
			r = rand.Intn(top)
			_, exists = m[r]
		}
		m[r] = struct{}{}
	}
	return m
}

func testFalsePositiveRate(t *testing.T, filter bloomfilter.IFilter, bitsCount, existedCount, hashCount, top int) {
	rate := bloomfilter.GetFlasePositiveRate(uint(bitsCount), uint(existedCount), uint(hashCount))
	t.Log("理论误判率 Theoretical false positive rate:", rate)
	nums := getRandomNums(top, existedCount)
	for m, _ := range nums {
		filter.Push([]byte(strconv.Itoa(m)))
	}
	count := 0
	for i:=0;i<top;i++{
		if filter.Exists([]byte(strconv.Itoa(i))){
			count ++
		}
	}
	realRate := float64(count - existedCount) / float64(top - existedCount)
	t.Log("实际误判率 Real false positive rate:", realRate)
}
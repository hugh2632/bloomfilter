package bloomfilter

import (
	"github.com/hugh2632/bloomfilter/global"
	"math"
	"strconv"
)

// Calculate the false positive rate. biteCount-The length of bits, existedCount-The existed elements in the filter, hashCount-The count of hash functions.
// 计算误判率, biteCount-比特长度, existedCount-元素, hashCount-哈希函数个数
func GetFalsePositiveRate(biteCount uint, existedCount uint, hashCount uint) float64 {
	return 1 - math.Pow(1.0-math.Exp(-float64(hashCount))*(float64(existedCount)+0.5)/(float64(biteCount)-1), float64(hashCount))
}

/*
	get the information entropy to test if a hash function is uniformly distributed. 获取信息熵以测试一个哈希方法是不是均匀分布的
 	-**-** H = -sum(p*log2(p))
	the target result should be: -64*(1/64*log2(1/64))=6. 理想的均匀分布最终得值是-64*(1/64*log2(1/64))=6=6。
*/
func CalculateInformationEntropy(f global.HashFunc) float64 {
	//Number of samples. 模拟样本个数
	n := 10000000
	//counter all the 1 bits. 所有比特位是1的计数
	counter := 0
	//Counter of every bit with value 1 return by Sum64(). 计算Sum64()方法返回值的每个位是1的个数
	var c = make([]int, 64)
	for i:=0;i<n;i++{
		var h = f()
		_, _ = h.Write([]byte(strconv.Itoa(i)))
		res := h.Sum64()
		for j:=0;j<64;j++{
			if res | 1 <<j == res {
				c[j]++
				counter++
			}
		}
	}
	//get the information entropy.
	var entropy float64
	for i:=0;i<64;i++{
		p := float64(c[i]) / float64(counter)
		entropy += - p * math.Log2(p)
	}
	return entropy
}

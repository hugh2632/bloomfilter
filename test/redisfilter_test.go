package test

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hugh2632/bloomfilter"
	"strconv"
	"testing"
)

var options = &redis.Options{
	Addr:     "192.168.20.101:6379",
	Username: "",
	Password: "",
	DB:       0,
}

var key = "test"

func TestInteractiveRedisFalsePositiveRate(t *testing.T) {
	cli := redis.NewClient(options)
	interactiveFilter, err := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Interactive, key, 10240, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	testFalsePositiveRate(t, interactiveFilter, 10240 * 8, 1000, 3, 1000000)
	// 理论误判率 Theoreticalfalse positive rate: 0.0018230817954481005
	// 实际误判率 Real false positive rate: 0.0005625625625625626
}

// To test if the CachedFilter's bytes can been used for InteractiveFilter
// 测试基于CachedFilter产生的数据是否可以直接使用在InteractiveFilter上。
func TestIsSame(t *testing.T) {
	cli := redis.NewClient(options)
	// Load CachedFilter
	cachedFilter, err := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Cached, key, 10240, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	fillNums(cachedFilter, 500, 600)
	// Cached Filter need use 'Write' method to truly write the bytes to redis.
	// CachedFilter需要使用‘Write’去真正把数据写入到Reids
	err = cachedFilter.Write()
	if err != nil {
		t.Fatal(err)
	}

	// copy the CachedFilter's bytes to the Interactive Filter.
	// 使用lua脚本将产生的数据，复制到InteractiveFilter存取的表中。
	err = cli.Eval(context.TODO(), "return redis.call('set','test',redis.call('hget','bloom','test'))", nil, nil).Err()
	if err != nil {
		t.Fatal(err)
	}

	// load an Interactive Filter to judge if nums have been stored in the filter.
	// 加载一个交互式过滤器来判断这些数字是否已存在在过滤器中
	interactiveFilter, _ := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Cached, "test", 10240, bloomfilter.DefaultHash...)
	for i := 550; i < 650; i++ {
		t.Logf("%d, %t\n", i, interactiveFilter.Exists([]byte(strconv.Itoa(i))))
	}
}

func TestRedisCachedFilter(t *testing.T) {
	cli := redis.NewClient(options)
	cachedFilter, err := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Cached, key, 10240, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("误判率，False positive rate:", bloomfilter.GetFlasePositiveRate(10240 * 8, 3, 2)) // 5.364385288193686e-11
	fillNums(cachedFilter, 250, 300)
	t.Log(cachedFilter.Exists([]byte(strconv.Itoa(290)))) // true
	t.Log(cachedFilter.Exists([]byte(strconv.Itoa(299)))) // true
	t.Log(cachedFilter.Exists([]byte(strconv.Itoa(350)))) // false
	// must use write to save the data to redis.
	// 必须使用write方法将数据全部提交到redis服务器
	cachedFilter.Write()
}

func TestInteractiveFilter(t *testing.T) {
	cli := redis.NewClient(options)
	interactiveFilter, err := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Interactive, key, 10240, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	fillNums(interactiveFilter, 250, 300)

	anotherFilter, _ := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Interactive, key, 10240, bloomfilter.DefaultHash...)

	t.Log(interactiveFilter.Exists([]byte(strconv.Itoa(290)))) // true
	t.Log(anotherFilter.Exists([]byte(strconv.Itoa(290))))     // true

	t.Log(interactiveFilter.Exists([]byte(strconv.Itoa(299)))) // true
	t.Log(anotherFilter.Exists([]byte(strconv.Itoa(299))))     // true

	t.Log(interactiveFilter.Exists([]byte(strconv.Itoa(320)))) // false
	t.Log(anotherFilter.Exists([]byte(strconv.Itoa(320))) )    // false
}

func TestRedisCachedClear(t *testing.T) {
	cli := redis.NewClient(options)
	filter, _ := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Cached, "test", 10240, bloomfilter.DefaultHash...)
	for i:=0;i<5;i++{
		t.Log("origin:", filter.Exists([]byte(strconv.Itoa(250))))
		fillNums(filter, 250, 250)
		t.Log("filled:", filter.Exists([]byte(strconv.Itoa(250))))
		filter.Clear()
		t.Log("cleared:", filter.Exists([]byte(strconv.Itoa(250))))
	}
	filter.Write()
}

func TestRedisInteractiveClear(t *testing.T) {
	cli := redis.NewClient(options)
	filter, _ := bloomfilter.NewRedisFilter(context.TODO(), cli, bloomfilter.RedisFilterType_Interactive, "test", 10240, bloomfilter.DefaultHash...)
	for i:=0;i<5;i++{
		t.Log("origin:", filter.Exists([]byte(strconv.Itoa(250))))
		fillNums(filter, 250, 250)
		t.Log("filled:", filter.Exists([]byte(strconv.Itoa(250))))
		filter.Clear()
		t.Log("cleared:", filter.Exists([]byte(strconv.Itoa(250))))
	}
}
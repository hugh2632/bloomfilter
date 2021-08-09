# 布隆过滤器

## 什么是布隆过滤器
**布隆过滤器本质上是一个数据结构，它可以用来判断某个元素是否在集合内，具有运行快速，内存占用小的特点。
而高效插入和查询的代价就是，Bloom Filter 是一个基于概率的数据结构：它只能告诉我们一个元素`绝对`不在集合内或`可能`在集合内。
A Bloom filter is a data structure designed to tell you, rapidly and memory-efficiently, whether an element is present in a set.
The price paid for this efficiency is that a Bloom filter is a probabilistic data structure: it tells us that the element either `definitely` is not in the set or `may be` in the set.**

## 过滤器类型 FilterTypes

| 类型 Type        | 特点    |  Feature  |
| --------   | -----  | :----:  |
| MemoryFilter      | 基于内存的过滤器，可以通过将字节内容加载到内存后进行操作   |   Filter based on memory. All the operations would be done in memory.     |
| CachedRedisFilter        | 基于MemoryFilter, 在使用`Write`方法时上传数据到Redis   |   Filter based on MemoryFilter, the result will be uploaded to redis server when the `Write` function is called.   |
| SQLFilter        |   基于MemoryFilter, 在使用`Write`方法时上传数据到Sql数据库。    |  Filter based on MemoryFilter, the result will be uploaded to sql database when the `Write` function is called.  |
| InteractiveRedisFilter        |    交互式的Redis布隆过滤器，使用`Push`和`Exists`方法时，都将实时与Redis服务器通讯。    |  Interactive filter which will update the redis `BitMap` through `SetBit` and `GetBit` by `Push` and `Exists` function.  |

### 使用方法 Usage

#### MemoryFilter
```
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

```

#### CachedRedisFilter
```
var options = &redis.Options{
	Addr:     "192.168.20.101:6379",
	Username: "",
	Password: "",
	DB:       0,
}

var key = "test"

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
```

#### SQLFilter
```
var sqlDSN = "user:password@tcp(ip:port)/database?charset=utf8mb4&parseTime=True&loc=Local"

func TestSqlFilter(t *testing.T) {
	// Init gorm.DB
	db, err := gorm.Open(mysql.Open(sqlDSN))
	if err != nil {
		t.Fatal(err)
	}
	// Init SQLFilter
	sqlFilter, err := bloomfilter.SqlFilter(db, "test", 1000, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	// Push 250-300 numbers to the filter.
	// 把250-300的数字压入过滤器
	fillNums(sqlFilter, 250, 300)
	sqlFilter.Write()
}

func TestSqlFilterExist(t *testing.T) {
	db, err := gorm.Open(mysql.Open(sqlDSN))
	if err != nil {
		t.Fatal(err)
	}
	sqlFilter, err := bloomfilter.SqlFilter(db, "test", 1000, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	// 280-300 should exist in filter, and 301-320 doesn't.
	// 280-300应该在过滤器中，而301-320不应该在。
	for i:=280;i<320;i++{
		t.Logf("%d: %t", i, sqlFilter.Exists([]byte(strconv.Itoa(i))))
	}
}

```

#### InteractiveRedisFilter
```
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
```

## 计算误差 Calculate the false positive rate
```
func TestMemFalsePositiveRate(t *testing.T) {
	memFilter := bloomfilter.NewMemoryFilter(make([]byte, 10240), bloomfilter.DefaultHash...).(*memory.Filter)
	testFalsePositiveRate(t, memFilter, 10240 * 8, 1000, 3, 100000000)
	//=== RUN   TestMemFalsePositiveRate
	//    common.go:37: 理论误判率 Theoretical false positive rate: 0.0018230817954481005
	//    common.go:49: 实际误判率 Real false positive rate:0.00047424474244742447
	//--- PASS: TestMemFalsePositiveRate (7.43s)
	//PASS
}
```

# 联系方式 Contact
讲解视频 <https://www.bilibili.com/video/BV1bJ411t7A8/>
邮箱    [mailto:hugh2632@hotmail.com]

# End

package test

import (
	"github.com/hugh2632/bloomfilter"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"testing"
)

var sqlDSN = "user:password@tcp(192.168.20.103:3306)/monitor?charset=utf8mb4&parseTime=True&loc=Local"

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

func TestSqlClear(t *testing.T) {
	db, err := gorm.Open(mysql.Open(sqlDSN))
	if err != nil {
		t.Fatal(err)
	}
	filter, err := bloomfilter.SqlFilter(db, "test", 1000, bloomfilter.DefaultHash...)
	if err != nil {
		t.Fatal(err)
	}
	//filter.Clear()
	for i:=0;i<5;i++{
		t.Log("origin:", filter.Exists([]byte(strconv.Itoa(250))))
		fillNums(filter, 250, 250)
		t.Log("filled:", filter.Exists([]byte(strconv.Itoa(250))))
		filter.Write()
		filter.Clear()
		t.Log("cleared:", filter.Exists([]byte(strconv.Itoa(250))))
	}
}
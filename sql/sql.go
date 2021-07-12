package sql

import (
	"github.com/hugh2632/bloomfilter/global"
	"github.com/hugh2632/bloomfilter/memory"
	"gorm.io/gorm"
)

// Memory based filter. Save data to database when use 'Write' function.
// 基于内存的过滤器。使用'Write'方法时上传结果到数据库。
type SQLFilter struct {
	memory.Filter
	id        string
	db        *gorm.DB
	tableName string
}

type Bloom struct {
	Id  string	`gorm:"primarykey;size:50"`
	Val []byte
}

// Delete the related row.
// 删除相关行
func (f *SQLFilter) Clear() error {
	bloom := Bloom{
		Id:  f.id,
		Val: nil,
	}
	err := f.db.Table(f.tableName).Delete(&bloom).Error
	if err != nil {
		return err
	}
	return f.Filter.Clear()
}

// Initial the filter. If the table does not exist in the database, it will be automatically created.
// 初始化过滤器。如果数据库中没有相应的表，将会被自动创建。
func (f *SQLFilter) Init(db *gorm.DB, tableName string, key string, bytes []byte) error {
	if key == "" {
		return global.ErrEmptyKey
	}
	f.db = db
	f.tableName = tableName
	f.id = key
	var bloom Bloom
	err := db.Table(f.tableName).Where("id=?", f.id).Limit(1).Scan(&bloom).Error
	if err != nil {
		return err
	}
	if bloom.Id == "" {
		bloom = Bloom{
			Id:  key,
			Val: bytes,
		}
		err = db.Table(f.tableName).Create(bloom).Error
		if err != nil {
			return err
		}
	}else{
		var bytes = bloom.Val
		if len(bytes) == len(f.Bytes) {
			f.Bytes = bytes
		} else {
			return global.ErrUnMatchLength
		}
	}
	return nil
}

func (f *SQLFilter) CreateTable() error{
	return f.db.Table(f.tableName).AutoMigrate(&Bloom{})
}

func (f *SQLFilter) Write() error {
	return f.db.Table(f.tableName).Where("id=?", f.id).Update("val", f.Bytes).Error
}

func (f *SQLFilter) Close() error {
	db, err := f.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

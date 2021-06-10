package global

import (
	"errors"
	"log"
)

var (
	ErrEmptyContent  = errors.New("输入内容为空")
	ErrEmptyKey = errors.New("key不能为空")
	ErrUnMatchLength = errors.New("字节长度不匹配")
	Logger           = log.Default()
)

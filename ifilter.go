package bloomfilter

//增量过滤器
type IFilter interface {
	Push([]byte)
	Exists([]byte) bool
	Close() error
	Write() error
	IsEmpty() bool
}

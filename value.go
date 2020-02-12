package ksl

// IValue 持久化数据接口
type IValue interface {
	// 获取表名
	TableName() string
}

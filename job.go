package ksl

import (
	"sync"
)

var (
	_tableDesc    sync.Map
	_table2JobKey BidiMap
)

type tableDesc struct {
	sync.Mutex
	ValueCount int
	UniqueNum  int
}

func lockValueOp(tableName string) {
	v, _ := _tableDesc.LoadOrStore(tableName, &tableDesc{})
	v.(*tableDesc).Lock()
}

func unlockValueOp(tableName string) {
	if v, exist := _tableDesc.Load(tableName); exist {
		v.(*tableDesc).Unlock()
	}
}

func getUniqueNum(tableName string) int {
	v, exist := _tableDesc.LoadOrStore(tableName, &tableDesc{})
	if !exist {
		return 0
	}

	desc := v.(*tableDesc)
	desc.UniqueNum++
	_tableDesc.Store(tableName, desc)
	// log.Println("assgin unique num:", desc.UniqueNum)
	return desc.UniqueNum
}

func putJobKey(tableName, filename string) {
	_table2JobKey.Put(tableName, filename)

	// log.Println("new job", tableName, filename)
}

func getJobKey(tableName string) string {
	v, exist := _table2JobKey.GetByKey(tableName)
	if !exist {
		return ""
	}
	return v.(string)
}

func removeJob(filename string) {
	_table2JobKey.RemoveByValue(filename)
}

func isJobExist(filename string) bool {
	return _table2JobKey.ExistValue(filename)
}

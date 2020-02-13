package ksl

import (
	"testing"

	"github.com/west3316/ksl/model"
)

// go test -v -count=1 github.com/west3316/ksl  -run=Test_fieldsFromStruct
func Test_fieldsFromStruct(t *testing.T) {
	t.Log(fieldsFromStruct(&model.UserCharge{}, "db"))
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_sqlxBulkInsertSQL
func Test_sqlxBulkInsertSQL(t *testing.T) {
	ob := &model.UserCharge{}
	sqlText := sqlxBulkInsertSQL(ob.TableName(), fieldsFromStruct(ob, "db"))
	t.Log(sqlText)

	ob2 := &model.UserOperateRecords{}
	sqlText = sqlxBulkInsertSQL(ob2.TableName(), fieldsFromStruct(ob2, "db"))
	t.Log(sqlText)
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_sqlxUpdateSQL
func Test_sqlxUpdateSQL(t *testing.T) {
	ob := &model.UserCharge{}
	sqlText := sqlxUpdateSQL(ob.TableName(), fieldsFromStruct(ob, "db"))
	t.Log(sqlText)

	ob2 := &model.UserOperateRecords{}
	sqlText = sqlxUpdateSQL(ob2.TableName(), fieldsFromStruct(ob2, "db"))
	t.Log(sqlText)
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_sqlxDeleteSQL
func Test_sqlxDeleteSQL(t *testing.T) {
	ob := &model.UserCharge{}
	sqlText := sqlxDeleteSQL(ob.TableName(), fieldsFromStruct(ob, "db"))
	t.Log(sqlText)

	ob2 := &model.UserOperateRecords{}
	sqlText = sqlxDeleteSQL(ob2.TableName(), fieldsFromStruct(ob2, "db"))
	t.Log(sqlText)
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_extractTableName
func Test_extractTableName(t *testing.T) {
	t.Log(extractTableName("user_charge-1578637499-6.ksl"))
}

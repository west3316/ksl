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
	sqlText := sqlxBulkInsertSQL("user_charge", fieldsFromStruct(&model.UserCharge{}, "db"))
	t.Log(sqlText)
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_sqlxUpdateSQL
func Test_sqlxUpdateSQL(t *testing.T) {
	sqlText := sqlxUpdateSQL("user_charge", fieldsFromStruct(&model.UserCharge{}, "db"))
	t.Log(sqlText)
}

// go test -v -count=1 github.com/west3316/ksl  -run=Test_extraceTableName
func Test_extraceTableName(t *testing.T) {
	t.Log(extractTableName("user_charge_1578637499_6.ksl"))
}

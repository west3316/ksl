package model

import "time"

const ()

// UserOperateRecords 表名: user_operate_records
type UserOperateRecords struct {

	//
	ID int `db:"id" mark:"primary_key"`

	//
	Type string `db:"type"`

	//
	At time.Time `db:"at"`
}

// TableName 表名
func (*UserOperateRecords) TableName() string {
	return "user_operate_records"
}

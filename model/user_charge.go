package model

import "time"

const (
	// 充值结果
	ResultNoResult = "NoResult"
	ResultSuccess  = "Success"
	ResultFail     = "Fail"
	ResultLocked   = "Locked"
)

// IsValidResult 检测枚举，充值结果
func IsValidResult(v string) bool {
	return map[string]bool{
		ResultNoResult: true,
		ResultSuccess:  true,
		ResultFail:     true,
		ResultLocked:   true,
	}[v]
}

// UserCharge 表名: user_charge
type UserCharge struct {

	// 唯一ID
	ID int `db:"id" mark:"primary_key,auto_increment"`

	// 用户ID
	UserID int `db:"user_id"`

	// 充值时间
	CreateAt time.Time `db:"create_at"`

	// 充值金额
	Value float64 `db:"value"`

	// 充值结果
	Result string `db:"result" faker:"enum-result"`

	// 备注
	Desc *string `db:"desc"`
}

// TableName 表名
func (*UserCharge) TableName() string {
	return "user_charge"
}

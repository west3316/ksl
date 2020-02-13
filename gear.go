package ksl

import (
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	autoIncrementMark           = "+"
	primaryKeyMark              = "."
	primaryKeyAutoIncrementMark = "*"

	tagPrimaryKey    = "primary_key"
	tagAutoIncrement = "auto_increment"
)

func isAutoIncrementMark(char byte) bool {
	return char == autoIncrementMark[0] || char == primaryKeyAutoIncrementMark[0]
}

func isPrimaryKeyMark(char byte) bool {
	return char == primaryKeyMark[0] || char == primaryKeyAutoIncrementMark[0]
}

func isPrimaryKeyAutoIncrementMark(char byte) bool {
	return char == primaryKeyAutoIncrementMark[0]
}

func isKeyPrefix(char byte) bool {
	for _, e := range []string{
		autoIncrementMark,
		primaryKeyMark,
		primaryKeyAutoIncrementMark,
	} {
		if char == e[0] {
			return true
		}
	}
	return false
}

func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// fieldsFromStruct 从struct中获取数据表字段
func fieldsFromStruct(v interface{}, tag string) []string {
	markToOpChar := func(s string) string {
		var opChar string
		if s == tagPrimaryKey {
			opChar = primaryKeyMark
		} else if s == tagAutoIncrement {
			opChar = autoIncrementMark
		}
		return opChar
	}

	var fields []string
	ty := reflect.Indirect(reflect.ValueOf(v)).Type()
	for i := 0; i < ty.NumField(); i++ {
		name := ty.Field(i).Tag.Get(tag)
		if name == "" {
			continue
		}

		// 标记自增字段
		var opChar string
		parts := strings.Split(ty.Field(i).Tag.Get("mark"), ",")
		if len(parts) == 1 {
			opChar = markToOpChar(parts[0])
		} else if len(parts) == 2 {
			if (parts[0] == tagPrimaryKey && parts[1] == tagAutoIncrement) ||
				(parts[1] == tagPrimaryKey && parts[0] == tagAutoIncrement) {
				opChar = primaryKeyAutoIncrementMark
			}
		}

		fields = append(fields, opChar+name)
	}
	return fields
}

func sqlxBulkInsertSQL(table string, fields []string) string {
	if len(fields) == 0 {
		return ""
	}

	var tableFields, valuesFields string
	for _, field := range fields {
		if isAutoIncrementMark(field[0]) {
			// 忽略自增字段
			continue
		}

		if isKeyPrefix(field[0]) {
			field = field[1:]
		}

		tableFields += "`" + field + "`,"
		valuesFields += ":" + field + ","
	}
	tableFields = tableFields[:len(tableFields)-1]
	valuesFields = valuesFields[:len(valuesFields)-1]

	return "INSERT INTO " + table + "(" + tableFields + ") VALUES(" + valuesFields + ")"
}

func sqlxUpdateSQL(table string, fields []string) string {
	if len(fields) == 0 {
		return ""
	}

	var fieldValues, pk string
	for _, field := range fields {
		if isKeyPrefix(field[0]) {
			field = field[1:]
			pk += "`" + field + "`=:" + field + " AND "
			continue
		}

		fieldValues += "`" + field + "`=:" + field + ","
	}
	pk = pk[:len(pk)-5]
	fieldValues = fieldValues[:len(fieldValues)-1]

	return "UPDATE " + table + " SET " + fieldValues + " WHERE " + pk
}

func sqlxDeleteSQL(table string, fields []string) string {
	if len(fields) == 0 {
		return ""
	}

	var pk string
	for _, field := range fields {
		if isPrimaryKeyMark(field[0]) {
			field = field[1:]
			pk += "`" + field + "`=:" + field + " AND "
			continue
		}
	}
	pk = pk[:len(pk)-5]

	return "DELETE FROM " + table + " WHERE " + pk
}

// WaitOSSignal 等待系统信号
func WaitOSSignal(sig ...os.Signal) {
	waitSig := make(chan os.Signal)
	signal.Notify(waitSig, sig...)
	<-waitSig
}

func getFileName(filename string) string {
	_, f := filepath.Split(filename)
	return f[:len(f)-len(filepath.Ext(f))]
}

type filenameDetail struct {
	TableName   string
	CreateAt    time.Time
	AccessTimes int
}

func (c *filenameDetail) Extract(filename string) bool {
	parts := strings.Split(getFileName(filename), "-")
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	c.CreateAt = time.Unix(ts, 0)
	accessTimes, err := strconv.Atoi(parts[2])
	if err != nil {
		return false
	}
	c.AccessTimes = accessTimes
	c.TableName = parts[0]
	return true
}

func (c filenameDetail) Make() string {
	return filepath.Join(storeDir, c.TableName+
		"-"+strconv.FormatInt(c.CreateAt.Unix(), 10)+"-"+
		strconv.Itoa(c.AccessTimes)+
		dataSuffix)
}

func extractTableName(filename string) string {
	fnd := &filenameDetail{}
	fnd.Extract(filename)
	return fnd.TableName
}

func extractAccessTimes(filename string) int {
	fnd := &filenameDetail{}
	fnd.Extract(filename)
	return fnd.AccessTimes
}

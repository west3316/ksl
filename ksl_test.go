package ksl

import (
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/bxcodec/faker"
	"github.com/vmihailenco/msgpack"

	"github.com/west3316/ksl/model"

	_ "github.com/go-sql-driver/mysql"
)

const dsn = "root:root@tcp(10.0.0.3:3306)/test?timeout=3s&writeTimeout=5s&readTimeout=2s&charset=utf8&parseTime=true&loc=Local"

func init() {
	DebugSetupProvider()
}

// go test -v -count=1 github.com/west3316/ksl -run=TestWrite
func TestWrite(t *testing.T) {
	Init(Option{
		DSN:            dsn,
		SyncValueCount: 30,
		SyncTimeout:    2,
	})

	DebugTruncate()

	data := &model.UserCharge{}
	err := faker.FakeData(data)
	if err != nil {
		t.Fatal("fake data:", err)
	}

	WriteInsert(data)

	// 等待系统终止信号
	WaitOSSignal(syscall.SIGTERM, os.Interrupt)
}

// go test -v -count=1 github.com/west3316/ksl -run=TestBulkWrite
func TestBulkWrite(t *testing.T) {
	Init(Option{
		DSN:            dsn,
		SyncValueCount: 20,
		SyncTimeout:    -1,
		BulkSize:       20,
	})

	DebugTruncate()

	var updateCount int
	tm := time.Now()
	for i := 0; i < 1000; i++ {
		ob := &model.UserCharge{}
		faker.FakeData(ob)
		WriteInsert(ob)
		if i%10 == 0 {
			ob2 := &model.UserCharge{}
			faker.FakeData(ob2)
			ob2.ID = i + 1
			s := "更新测试数据"
			ob2.Desc = &s
			WriteUpdate(ob2)
			updateCount++
		}
	}

	t.Log("更新测试数据", updateCount)
	// 等待系统终止信号
	WaitOSSignal(syscall.SIGTERM, os.Interrupt)
	t.Log(time.Since(tm))
}

// go test -v -count=1 github.com/west3316/ksl -run=TestGoBulkWrite
func TestGoBulkWrite(t *testing.T) {
	Init(Option{
		DSN:            dsn,
		SyncValueCount: 5,
		SyncTimeout:    2,
		BulkSize:       3,
	})

	DebugTruncate()

	var updateCount int
	tm := time.Now()
	for i := 0; i < 1000; i++ {
		go func(i int) {
			ob := &model.UserCharge{}
			faker.FakeData(ob)
			WriteInsert(ob)

			if i%10 == 0 {
				ob2 := &model.UserCharge{}
				faker.FakeData(ob2)
				ob2.ID = i + 1
				s := "更新测试数据"
				ob2.Desc = &s
				WriteUpdate(ob2)
				updateCount++
			}
		}(i)

		go func(id int) {
			ob2 := &model.UserOperateRecords{}
			faker.FakeData(ob2)
			ob2.ID = id
			WriteInsert(ob2)

			if id%10 == 5 {
				WriteDelete(ob2)
			}
		}(i + 1)
	}

	t.Log("更新测试数据", updateCount)
	// 等待系统终止信号
	WaitOSSignal(syscall.SIGTERM, os.Interrupt)
	t.Log(time.Since(tm))
}

// go test -v -count=1 github.com/west3316/ksl -run=TestDelayUpdateWrite
func TestDelayUpdateWrite(t *testing.T) {
	Init(Option{
		DSN:            dsn,
		SyncValueCount: 100,
		SyncTimeout:    3,
		BulkSize:       100,
	})

	DebugTruncate()

	ob := &model.UserOperateRecords{}
	faker.FakeData(ob)
	WriteInsert(ob)

	time.Sleep(time.Second * 4)
	ob.Type = "TestDelayUpdateWrite"
	WriteUpdate(ob)
	time.Sleep(time.Second * 8)
}

// go test -v -count=1 github.com/west3316/ksl -run Test_decodeArrayToMap
func Test_decodeArrayToMap(t *testing.T) {
	filename := "sqlData/user_charge_1581499298_1.ksl"
	f, err := os.Open(filename)
	if err != nil {
		log.Println("打开数据文件 "+filename+" 失败：", err)
		return
	}

	header, reader := readHeader(f)
	if header == "" {
		log.Println("从"+filename+"读取文件头失败：", err)
		return
	}

	// 获得字段名
	fields := strings.Split(header, " ")
	// 生成批量插入语句
	sqlText := sqlxBulkInsertSQL(extractTableName(filename), fields)

	var opChar byte
	var value map[string]interface{}
	var values []map[string]interface{}
	decoder := msgpack.NewDecoder(reader)
	for err == nil {
		opChar, value, err = decodeArrayToMap(decoder, fields)
		if err == nil || err == io.EOF {
			if len(value) != 0 {
				values = append(values, value)
				log.Println("value:", value)
			}
		} else {
			log.Println("decodeArrayToMap错误:", err)
		}
		_ = opChar
	}

	_ = sqlText
}

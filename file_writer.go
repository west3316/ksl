package ksl

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

const storeDir = "sqlData"

// 数据文件后缀
const dataSuffix = ".ksl"

// 定义存盘操作符
const (
	opInsert = 'C'
	opUpdate = 'U'
)

type valueWarpper struct {
	Operator byte
	Value    IValue
}

var (
	// 达到数据条数写入DB
	_syncValueCount = 300
	// 超时入库
	_syncValueTimeout = time.Duration(5 * time.Second)
	// 数据文件存储值数量
	_bulkSize = 300
)

// WriteInsert 插入数据
func WriteInsert(v IValue) {
	write(v, opInsert)
}

// WriteUpdate 更新数据
func WriteUpdate(v IValue) {
	write(v, opUpdate)
}

func initDataDir() {
	// 创建数据存储目录
	if !dirExists(storeDir) {
		err := os.Mkdir(storeDir, 0700)
		if err != nil {
			log.Fatalln("无法创建数据存储目录", storeDir, "：", err)
		}
	} else {
		// 恢复未完成的job
		syncBulkFileToDB()
	}
}

func writeHeader(f *os.File, v IValue) bool {
	fields := fieldsFromStruct(v, "db")
	content := strings.Join(fields, " ") + "\n"
	n, err := f.WriteString(content)
	if err != nil || n != len(content) {
		return false
	}
	return true
}

func readHeader(f *os.File) (string, *bufio.Reader) {
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	if err != nil {
		return "", nil
	}

	return string(line), r
}

// write 写入数据
func write(v IValue, opChar byte) {
	var sync bool
	var ctx *writeOpContext
	_opWriteMgr.lockValueOp(v.TableName())
	defer func() {
		if sync {
			go syncBulkFileToDB()
		}
		_opWriteMgr.storeValue(v.TableName(), ctx)
		_opWriteMgr.unlockValueOp(v.TableName())
	}()

	for hasNext := true; hasNext; {
		ctx = assignWriteOpContext(v)
		if ctx == nil {
			log.Println("无法创建数据文件")
			return
		}

		fi, err := ctx.File.Stat()
		if err != nil {
			log.Println("获取文件", fi.Name(), "状态失败", err)
			return
		}

		hasNext = time.Since(fi.ModTime()) >= _syncValueTimeout
		if hasNext {
			resetWriteFile(v.TableName())
		}
	}

	err := ctx.Encoder.Encode(valueWarpper{Value: v, Operator: opChar})
	if err != nil {
		log.Println("序列化对象失败：", err)
		return
	}

	err = ctx.File.Sync()
	if err != nil {
		log.Println("写入磁盘失败：", err)
		return
	}

	// 增加数列化计数
	ctx.Count++

	if ctx.Count >= _bulkSize {
		// 文件记录上限，则使用新文件
		resetWriteFile(v.TableName())
		// log.Println("产生数据文件：", filename)
		sync = ctx.Count >= _syncValueCount
	}
}

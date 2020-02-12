package ksl

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack"
)

const storeDir = "sqlData"

// 数据文件后缀
const dataSuffix = ".ksl"

type writerDesc struct {
	Encoder *msgpack.Encoder
	Count   int
	File    *os.File
	// value样本
	Value IValue
}

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
	// 文件名映射到 writerDesc
	_w2desc = make(map[string]*writerDesc)
	// 达到数据条数写入DB
	_syncValueCount = 300
	// 超时入库
	_syncValueTimeout = time.Duration(5 * time.Second)
	// 数据文件存储值数量
	_bulkSize = 300
)

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

func makeFilename(v IValue) string {
	filename := getJobKey(v.TableName())
	if filename != "" && isJobExist(filename) {
		return filename
	}
	filename = filepath.Join(storeDir, v.TableName()+
		"_"+strconv.Itoa(int(time.Now().Unix()))+"_"+
		strconv.Itoa(getUniqueNum(v.TableName()))+
		dataSuffix)
	putJobKey(v.TableName(), filename)
	return filename
}

func newWriterDesc(filename string) *writerDesc {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeExclusive)
	if err != nil {
		log.Println("创建文件失败", filename, "：", err)
		return nil
	}

	return &writerDesc{
		Encoder: msgpack.NewEncoder(f).StructAsArray(true),
		File:    f,
	}
}

func removeWriterDesc(filename string) {
	desc, exist := _w2desc[filename]
	if exist {
		err := desc.File.Chmod(0600)
		if err != nil {
			log.Println("修改文件失败", filename, "：", err)
		}
		desc.File.Close()
	}
	delete(_w2desc, filename)
	removeJob(filename)
}

func getWriterDesc(filename string, val IValue) *writerDesc {
	var desc *writerDesc
	desc, exist := _w2desc[filename]
	if !exist {
		desc = newWriterDesc(filename)
		if desc == nil {
			return nil
		}

		desc.Value = val
		// 写入文件头
		if !writeHeader(desc.File, val) {
			return nil
		}
		_w2desc[filename] = desc
	}
	return desc
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

// WriteInsert 插入数据
func WriteInsert(v IValue) {
	write(v, opInsert)
}

// WriteUpdate 更新数据
func WriteUpdate(v IValue) {
	write(v, opUpdate)
}

// write 写入数据
func write(v IValue, opChar byte) {
	var sync bool
	lockValueOp(v.TableName())
	defer func() {
		if sync {
			go syncBulkFileToDB()
		}
		unlockValueOp(v.TableName())
	}()

	var desc *writerDesc
	var filename string
	for hasNext := true; hasNext; {
		filename = makeFilename(v)
		desc = getWriterDesc(filename, v)
		if desc == nil {
			log.Println("无法创建数据文件", filename)
			return
		}

		fi, err := desc.File.Stat()
		if err != nil {
			log.Println("获取文件", filename, "状态失败", err)
			return
		}

		hasNext = time.Since(fi.ModTime()) >= _syncValueTimeout
		if hasNext {
			removeWriterDesc(filename)
		}
	}

	err := desc.Encoder.Encode(valueWarpper{Value: v, Operator: opChar})
	if err != nil {
		log.Println("序列化对象失败：", err)
		return
	}

	err = desc.File.Sync()
	if err != nil {
		log.Println("写入磁盘失败：", err)
		return
	}

	// 增加数列化计数
	desc.Count++

	if desc.Count >= _bulkSize {
		// 文件记录上限，则使用新文件
		removeWriterDesc(filename)
		// log.Println("产生数据文件：", filename)
		sync = desc.Count >= _syncValueCount
	}
}

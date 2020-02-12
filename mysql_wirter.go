package ksl

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack"

	_ "github.com/go-sql-driver/mysql"

	// 暂时使用支持[]map[string][]interface{}的版本
	"github.com/jmoiron/sqlx"
)

var (
	_db *sqlx.DB
)

// Option 选项
type Option struct {
	// mysql连接配置
	DSN string
	// 数据文件中数据最大条数
	BulkSize int
	// 入库数量
	SyncValueCount int
	// 超时入库，秒
	SyncTimeout int
}

// Init 初始化数据源
func Init(opt Option) {
	if opt.SyncTimeout > 1 {
		_syncValueTimeout = time.Duration(opt.SyncTimeout) * time.Second
		log.Println("超时：", _syncValueTimeout)
	}

	if opt.BulkSize > 1 {
		_bulkSize = opt.BulkSize
		log.Println("数据文件存储value数量：", _bulkSize)
	}

	_syncValueCount = opt.SyncValueCount
	log.Println("同步value数量：", _syncValueCount)

	var err error
	_db, err = sqlx.Connect("mysql", opt.DSN)
	if err != nil {
		log.Fatalln("连接MySQL失败：", err)
	}

	if opt.SyncTimeout != -1 {
		// 超时自动同步
		time.AfterFunc(_syncValueTimeout, onSyncTimeout)
	}

	initDataDir()
}

func onSyncTimeout() {
	syncBulkFileToDB()
	time.AfterFunc(_syncValueTimeout, onSyncTimeout)
}

func syncBulkFileToDB() {
	fis, err := ioutil.ReadDir(storeDir)
	if err != nil {
		log.Println("无法打开目录[" + storeDir + "]")
		return
	}

	for _, fi := range fis {
		if fi.IsDir() || filepath.Ext(fi.Name()) != dataSuffix || fi.Size() == 0 {
			continue
		}

		tableName := extractTableName(fi.Name())
		filename := filepath.Join(storeDir, fi.Name())
		go func(filename, tableName string) {
			// log.Println("同步开始：", filename)
			if syncToMySQL(filename) {
				// 同步成功后删除数据文件
				os.Remove(filename)
			}
		}(filename, tableName)
	}
}

func syncToMySQL(filename string) (ok bool) {
	for hasNext := true; hasNext; {
		fi, err := os.Stat(filename)
		if err != nil {
			log.Println("读取文件信息失败：", err)
			return
		}

		hasNext = fi.Mode() == 0 && time.Since(fi.ModTime()) > _syncValueTimeout
		if hasNext {
			os.Chmod(filename, 0600)
		} else if fi.Mode() == 0 || fi.Size() == 0 {
			return
		}
		// log.Println("文件大小：", fi.Size())
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModeExclusive)
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
	sqlInsertText := sqlxBulkInsertSQL(extractTableName(filename), fields)
	sqlUpdateText := sqlxUpdateSQL(extractTableName(filename), fields)

	var opChar byte
	var value map[string]interface{}
	var insertValues []map[string]interface{}
	var updateValues []map[string]interface{}
	decoder := msgpack.NewDecoder(reader)
	for err == nil {
		opChar, value, err = decodeArrayToMap(decoder, fields)
		if err == nil || err == io.EOF {
			if len(value) == 0 {
				continue
			}
			if opChar == opInsert {
				insertValues = append(insertValues, value)
			} else if opChar == opUpdate {
				updateValues = append(updateValues, value)
			}
		} else {
			log.Println("decodeArrayToMap错误:", err)
		}
	}

	f.Close()

	// 批量插入
	if len(insertValues) != 0 {
		_, err2 := _db.NamedExec(sqlInsertText, insertValues)
		if err2 != nil {
			log.Println("入库失败\nSQL语句：", sqlInsertText, "，错误：", err2)
			return
		}
	}

	// 逐条更新
	for _, e := range updateValues {
		_, err2 := _db.NamedExec(sqlUpdateText, e)
		if err2 != nil {
			log.Println("入库失败\nSQL语句：", sqlUpdateText, "，错误：", err2)
			return
		}
	}

	if !errors.Is(err, io.EOF) {
		log.Println("解析数据文件失败：", err)
		return
	}

	// log.Println("DB同步完成["+filename+"]：")
	// 标记同步完成
	removeJob(filename)

	return true
}

func decodeArrayToMap(decoder *msgpack.Decoder, keys []string) (byte, map[string]interface{}, error) {
	l, err := decoder.DecodeArrayLen()
	if err != nil {
		return 0, nil, err
	}

	if l != 2 {
		return 0, nil, errors.New("wrong format, exmaple: [operate char, value]")
	}

	// 获取操作符
	var opChar uint8
	opChar, err = decoder.DecodeUint8()
	if err != nil {
		return 0, nil, errors.New("operate char format error")
	}

	l, err = decoder.DecodeArrayLen()
	if err != nil {
		return 0, nil, err
	}

	if l != len(keys) {
		return 0, nil, errors.New("key value not match:" + strconv.Itoa(l) + "-" + strconv.Itoa(len(keys)))
	}

	var result = make(map[string]interface{}, l)
	for _, key := range keys {
		v, err := decoder.DecodeInterface()
		if err != nil {
			return 0, nil, errors.New("decode error")
		}

		if byte(opChar) == opInsert && isAutoIncrementMark(key[0]) {
			// 忽略自增字段
			continue
		}

		if isKeyPrefix(key[0]) {
			key = key[1:]
		}

		result[key] = v
	}

	return byte(opChar), result, nil
}

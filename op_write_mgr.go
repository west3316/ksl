package ksl

import (
	"log"
	"os"
	"time"

	"github.com/vmihailenco/msgpack"
)

var (
	_opWriteMgr opMgr
)

type writeOpContext struct {
	File        *os.File
	Encoder     *msgpack.Encoder
	Count       int
	AccessTimes int
}

func resetWriteFile(tableName string) {
	v := _opWriteMgr.loadValue(tableName)
	if v == nil {
		return
	}

	ctx := v.(*writeOpContext)
	err := ctx.File.Chmod(0600)
	if err != nil {
		log.Println("修改文件失败", 1, "：", err)
	}
	ctx.File.Close()
	// 是否文件对象
	ctx.File = nil
	ctx.Count = 0
}

func assignWriteOpContext(v IValue) *writeOpContext {
	var err error
	var ctx *writeOpContext
	_opWriteMgr.updateValue(v.TableName(), func(value interface{}) interface{} {
		if value != nil {
			ctx = value.(*writeOpContext)
			ctx.AccessTimes++
		} else {
			ctx = &writeOpContext{AccessTimes: 1}
		}

		if ctx.File == nil {
			filename := filenameDetail{
				TableName:   v.TableName(),
				CreateAt:    time.Now(),
				AccessTimes: ctx.AccessTimes,
			}.Make()

			ctx.File, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeExclusive)
			if err != nil {
				log.Println("创建文件失败", filename, "：", err)
				return nil
			}

			// 写入文件头
			if !writeHeader(ctx.File, v) {
				return nil
			}

			ctx.Encoder = msgpack.NewEncoder(ctx.File).StructAsArray(true)
		}
		return ctx
	})

	return ctx
}

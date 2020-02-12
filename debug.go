package ksl

import (
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/west3316/ksl/model"

	"github.com/bxcodec/faker"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DebugTruncate 清空表
func DebugTruncate() {
	_, err := _db.Exec("TRUNCATE TABLE `user_charge`")
	if err != nil {
		log.Fatalln("TRUNCATE TABLE failed!")
	}
}

var (
	enumResult = []string{model.ResultNoResult, model.ResultSuccess, model.ResultFail, model.ResultLocked}
)

func oneOf(results []string) string {
	return results[rand.Intn(len(results))]
}

// DebugSetupProvider _
func DebugSetupProvider() {
	err := faker.AddProvider("enum-result", func(v reflect.Value) (interface{}, error) {
		return oneOf(enumResult), nil
	})
	if err != nil {
		log.Fatalln("add faker provider fail:", err)
	}
}

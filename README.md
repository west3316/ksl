# ksl

ksl（开塞露）用来处理mysql（或mariadb）统计数据，采用异步、批量写入，从而提高入库性能，解决统计数据频繁单条写入卡死问题。流程，ksl先将BulkSize条数据写到文件中，并间隔SyncTimeout秒后，将数据文件写入数据库。

## 使用说明

```bash
go get github.com/west3316/ksl
```
配合[auto-model](https://github.com/west3316/auto-model)工具食用快乐加倍

```go
import (
    ksl "github.com/west3316/ksl"
)

const dsn = "root:root@tcp(localhost:3306)/test?timeout=3s&writeTimeout=5s&readTimeout=2s&charset=utf8&parseTime=true&loc=Local"

func main() {
    ksl.Init(ksl.Option{
		DSN: dsn,
    // 每个数据文件最多记录30条数据
		BulkSize: 30,
    // 每隔2秒将未入库数据库文件同步到数据库
    SyncTimeout: 2,
    // 数据条数达到30条，立即同步到数据库
    // 与SyncTimeout一同使用，分散入库时间点，防止拥塞
    SyncValueCount: 30,
	})

  // model.UserCharge 根据数据表生成的go结构，采用auto-model工具生成
  // 此方法并发安全
  // 批量写入采用了sqlx库，struct的成员变量必须定义db tag
  ksl.WriteInsert(&model.UserCharge{})

  // 更新单条数据，入库非批量操作
  ksl.WriteUpdate(&model.UserCharge{ID: 10, Desc: "更新数据"})
}
```

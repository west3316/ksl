# ksl

ksl（开塞露）是一个mysql（或mariadb）操作库，实现了数据异步入库、文件缓存、批量插入，适用于非实时数据存储场景，例如：日志、统计数据入库

## 工作流程

工作流程，ksl先将`BulkSize`条数据写到文件中，并间隔`SyncTimeout`秒后，将数据文件写入数据库，然后删除数据缓存文件。同时从文件中读取数据入库，也会有拥塞，设置`SyncValueCount`参数后，当写入数据条数达到`SyncValueCount`值，会立即将数据文件入库


> 注意：SyncTimeout 可以为负值，表示不使用定时入库策略，这时候，务必将 BulkSize、 SyncValueCount 设置成相同值，以防止部分数据不会入库。

## 使用说明

```bash
go get github.com/west3316/ksl
```

ksl 提供三个数据操作接口 `WriteInsert、 WriteUpdate、WriteDelete`，传入参数 value 除了需要实现接口 TableName，还需要使用 `db` tag 指明表中的字段名，类似 json 解析，和用 `mark` tag标记主键字段和自增字段。这些都可以使用[auto-model](https://github.com/west3316/auto-model)工具来自动完成。

```go
import (
    ksl "github.com/west3316/ksl"
)

const dsn = "root:root@tcp(localhost:3306)/test?timeout=3s&writeTimeout=5s&readTimeout=2s&charset=utf8&parseTime=true&loc=Local"

func main() {
    ksl.Init(ksl.Option{
    // 数据库连接 DSN 
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

  // 删除单条数据，入库非批量操作
  ksl.WriteDelete(&model.UserCharge{ID: 10})
}
```

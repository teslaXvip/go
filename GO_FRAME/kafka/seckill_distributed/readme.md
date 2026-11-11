已按要求完成「Redis 令牌桶 → Redis 库存减 1 → Kafka 订单队列 → 数据库串行落库」的分布式秒杀。

## 代码结构（`kafka/seckill_distributed`）

| 文件 | 对应图中环节 |
|---|---|
| [token_bucket.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/token_bucket.go) | Redis 令牌桶：Lua 原子「读-按时间补令牌-扣 1」，单次 Redis 调用，支持高并发与多实例 |
| [stock_cache.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/stock_cache.go) | Redis 库存减 1：Lua 判 `stock>0` 才 DECR，Redis 层直接挡住超卖；入队失败可回补 |
| [order_queue.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/order_queue.go) | 订单队列生产端：kafka-go Writer，key=商品 ID 哈希分区，`RequireAll` |
| [consumer.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/consumer.go) | 订单队列消费端：**单 goroutine 顺序消费**（并行写库 → 串行写库），处理完手动提交 offset |
| [store.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/store.go) | 数据库：MySQL/GORM 事务（`stock>0` 条件更新 + 插订单）；无 MySQL 时退回内存实现 |
| [main.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/main.go) | 主链路 `App.Do` + 并发压测 + 等消费追平后统计 |
| [config.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/config.go) | 参数，全部可用 `SECKILL_*` 环境变量覆盖 |
| [seckill_test.go](file:///d:/go/GO_FRAME/kafka/seckill_distributed/seckill_test.go) | 并发不超卖、库存不被重置、单条订单处理单测 |

关键点：库存以 MySQL 为准 → 启动时 `EnsureStock` 后预热到 Redis；Redis 只是拦截层，Kafka 削峰后由单消费者串行落库，数据库层还有 `WHERE stock > 0` 兜底。

## 依赖

新增 `github.com/segmentio/kafka-go v0.4.51`（已写入 go.mod/go.sum），Redis 复用已有 `go-redis/v9`。

## 验证结果

- `go build ./kafka/...` → 0
- `go vet ./kafka/seckill_distributed/` → 0
- `go test -count=1 ./kafka/seckill_distributed/` → ok，0.300s（并发 1000 抢 100 库存，成功恰好 100、库存 0、订单 100）
- `go run ./kafka/seckill_distributed/` → 退出码 1，提示 `连接 Redis(127.0.0.1:6379) 失败`：本机 6379/9092 当前都没监听，**端到端链路尚未真实验证**

## 你起好 Redis + Kafka 后这样跑

```powershell
go run ./kafka/seckill_distributed/
```

默认 `127.0.0.1:6379`、`127.0.0.1:9092`、topic `seckill-orders`、库存 100、请求 5000、令牌桶容量 200 + 1000/s。可按需覆盖：

```powershell
$env:SECKILL_KAFKA_BROKERS="127.0.0.1:9092"
$env:SECKILL_STOCK="100"
$env:SECKILL_REQUESTS="5000"
$env:SECKILL_MYSQL_DSN="tester:123456@tcp(127.0.0.1:3307)/blog?charset=utf8mb4&parseTime=True&loc=Local"
```

两点提醒：topic 建议单分区（或让同商品 key 落同一分区），否则「串行写库」只在单分区内成立；MySQL 连不上会自动退回内存存储并打警告，此时数据库结果仅用于演示。
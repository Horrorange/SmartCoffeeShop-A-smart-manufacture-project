# 数据库模式与集成说明

## 表结构

```sql
CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  coffee_type VARCHAR(32) NOT NULL,
  bool_ice BOOLEAN NOT NULL,
  table_num INT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finish_time TIMESTAMPTZ,
  error_msg TEXT
);
```

## 字段设计说明

- `status`: 流水线的状态机锚点，避免重复处理与并发竞态。典型取值：`pending` → `queued` → `done` 或 `error`。
- `create_time`: 订单入库时间，便于轮询排序与审计。
- `finish_time`: 流水线完成时间，用于 SLA 统计与对账。
- `error_msg`: 失败原因记录，便于问题定位与重试策略。

## 接入真实数据库的配置入口

- 代码中通过接口抽象了数据库与队列：
  - `pipeline.DB` 提供 `PollPending`、`MarkQueued`、`UpdateResult` 方法；当前演示实现为 `InMemoryDB`。
  - `pipeline.Queue` 抽象了 `queue_orders` 与 `queue_completed`；当前演示实现为 `InMemoryQueue`。
- 接入 PostgreSQL/MySQL 时，新增实现：
  - `type SQLDB struct { db *sql.DB }`，在 `NewSQLDB(driver, dsn)` 读取环境变量 `DB_DRIVER`、`DB_DSN` 完成初始化。
  - 使用 `SELECT ... FOR UPDATE SKIP LOCKED` 或分页扫描实现 `PollPending`，在同一事务内将命中的记录更新为 `queued`。
  - 使用 `UPDATE orders SET status=..., finish_time=..., error_msg=... WHERE id=?` 实现 `UpdateResult`。
- 接入 RabbitMQ 时，新增实现：
  - `type AMQPQueue struct { conn *amqp.Connection; ch *amqp.Channel }`，读取 `AMQP_URL`、`AMQP_QUEUE_ORDERS`、`AMQP_QUEUE_COMPLETED`。
  - 将 `PublishOrder/PublishResult` 映射到 `basic.publish`，`ConsumeOrders/ConsumeResults` 映射到 `basic.consume` 并进行 JSON 编解码。

以上接口位置均在 `smart_gateway/pipeline` 包中，替换演示实现即可无缝对接生产环境。
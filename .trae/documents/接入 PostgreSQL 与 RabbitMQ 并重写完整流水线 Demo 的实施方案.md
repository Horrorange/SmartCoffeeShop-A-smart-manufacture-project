## 目标
- 将现有 `gateway/pipeline` 与真实 PostgreSQL（5432）和 RabbitMQ 对接，跑通三明治架构：SQL → 队列 → 流水线 → 结果回写 SQL。
- 新增一个 Go 运行入口加载环境变量并启动三组件；新增一个 Python Demo，向 SQL 插入订单（pending），触发流水线处理并最终在 SQL 中看到完成状态。

## 架构接入点
- 数据库：新增 `SQLDB`（实现 `gateway/pipeline.DB`），通过 `database/sql` + `pgx`/`lib/pq` 连接 PostgreSQL。
- 队列：新增 `AMQPQueue`（实现 `gateway/pipeline.Queue`），通过 `amqp091-go` 连接 RabbitMQ，声明/消费 `queue_orders` 与 `queue_completed`。
- 设备：沿用 `gateway` 内的真实设备驱动 `Grinder/CoffeeMachine/IceMaker/DeliveryRobot`。

## 数据库层实现（SQLDB）
- 连接配置（全部用环境变量，避免硬编码密码）：
  - `PG_HOST=localhost`、`PG_PORT=5432`、`PG_DB=smartshop`、`PG_USER=postgres`、`PG_PASS=885658`
  - DSN 示例：`postgres://postgres:885658@localhost:5432/smartshop?sslmode=disable`
- 建表与索引（启动时自动执行，若不存在则创建）：
  - DDL：
    ```sql
    CREATE TABLE IF NOT EXISTS orders (
      id BIGSERIAL PRIMARY KEY,
      coffee_type VARCHAR(32) NOT NULL,
      bool_ice BOOLEAN NOT NULL,
      table_num INT NOT NULL,
      status VARCHAR(16) NOT NULL DEFAULT 'pending',
      create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      finish_time TIMESTAMPTZ,
      error_msg TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_orders_status_ctime ON orders(status, create_time DESC);
    ```
- 方法实现：
  - `PollPending(limit)`：`BEGIN`；`SELECT id, coffee_type, bool_ice, table_num, create_time FROM orders WHERE status='pending' ORDER BY create_time ASC FOR UPDATE SKIP LOCKED LIMIT $1;` 在同一事务内将这些 `id` 批量 `UPDATE status='queued'`；`COMMIT`；返回订单。
  - `MarkQueued(ids)`：用于与内存实现对齐，SQLDB中由 `PollPending` 内部完成（返回 `nil`）。
  - `UpdateResult(res)`：根据 `res.BoolIsDone` 写入 `status='done'/'error'`、`finish_time=NOW()`、`error_msg`。

## 队列层实现（AMQPQueue）
- 连接配置：`AMQP_URL=amqp://guest:guest@localhost:5672/`；队列名：`AMQP_QUEUE_ORDERS=queue_orders`、`AMQP_QUEUE_COMPLETED=queue_completed`。
- 初始化：建立连接与通道，声明两个队列（耐心创建，`durable=true`）。
- `PublishOrder(o)`：将订单 JSON 发布到 `queue_orders`，属性持久化。
- `ConsumeOrders(ctx)`：消费 `queue_orders`，按消息解码为 `Order`；自动 `Ack`；输出到只读通道。
- `PublishResult(r)` / `ConsumeResults(ctx)` 同理操作 `queue_completed`。

## 与流水线核心整合
- 在新的 Go 入口中：
  - 优先使用 SQLDB 与 AMQPQueue；若环境变量未提供，则回落到内存实现（方便本地演示）。
  - 构建 `OrderPoller(DB, Queue)`、`Core(Queue, Grinder, Coffee, Ice, Robot, ICE_MIN, ICE_OUT)`、`ResultSyncer(DB, Queue)` 并启动。
- 错误处理与重试：
  - SQL/AMQP 连接加入重试（指数退避）；
  - 消费异常记录日志并继续；
  - 设备阶段失败即刻发布 `error` 结果，由 `UpdateResult` 写回。

## 新 Demo 入口
- Go 入口：`smart_gateway/cmd/rabbit_sql_pipeline/main.go`
  - 读取设备、SQL、AMQP 的所有环境变量；自动建表；启动三组件；输出进度与最终结果。
- Python Demo：`script/pipeline_demo/send_order.py`
  - 通过 `psycopg2` 连接 Postgres（使用用户提供的密码 `885658`，推荐从环境变量 `PG_PASS` 读取）
  - 插入一条/多条 `pending` 订单（如：`LATTE`、`AMERICANO` 等）；
  - 可选：查询队列消费完成后状态（循环查询 `status` 字段）并打印。

## 环境变量清单
- 设备：`COFFEE_HOST/COFFEE_PORT`、`GRINDER_HOST/GRINDER_PORT`、`ICE_HOST/ICE_RACK/ICE_SLOT`、`MQTT_HOST/MQTT_PORT`
- SQL：`PG_HOST`、`PG_PORT`、`PG_DB`、`PG_USER`、`PG_PASS`
- AMQP：`AMQP_URL`、`AMQP_QUEUE_ORDERS`、`AMQP_QUEUE_COMPLETED`
- 制冰参数：`ICE_MIN_STOCK`（默认 200）、`ICE_DISPENSE_AMOUNT`（默认 100）

## 验证方案
- 启动 RabbitMQ（Docker 或本地）与设备模拟器；配置好 SQL 环境变量。
- 运行 Go 入口，确保建表成功并消费队列。
- 执行 Python Demo 插入订单；观察 Go 程序日志、设备调用与 SQL 状态从 `pending → queued → done/error`。

## 交付物
- 新增：`gateway/pipeline/sqldb.go`、`gateway/pipeline/amqpqueue.go`
- 新增：`cmd/rabbit_sql_pipeline/main.go`（Go 入口），`script/pipeline_demo/send_order.py`（Python 入口）
- 更新：`GO_GUIDE.md` 增加 SQL/RabbitMQ 部分与运行说明

确认后我将开始实现上述代码、接入配置并提供可直接运行的 Demo。
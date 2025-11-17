## 目标
- 将订单流水线（订单搬运工→流水线核心→结果记录员）接入现有 `smart_gateway/gateway`，直接调用真实设备驱动（磨豆机/咖啡机/制冰机/MQTT机器人）。
- 新增一个 Demo 入口，加载 `gateway.Config` 环境变量，对接数据库与 RabbitMQ，跑通全流程并打印结果。

## 集成思路
- 位置与命名：在 `smart_gateway/gateway/pipeline/` 下实现三组件，复用现有设备类型：`Grinder`、`CoffeeMachine`、`IceMaker`、`DeliveryRobot`。
- 设备调用映射：
  - 磨豆：`Grinder.GrindAuto()`
  - 制作：`CoffeeMachine.Make(order.CoffeeType)`
  - 取冰（仅当 `order.BoolIce`）：`IceMaker.ProduceUntil(min)` → `IceMaker.Dispense(amount)`（`min/amount` 可配置，默认 `min=200, amount=100`）
  - 送餐：`DeliveryRobot.Deliver(order.CoffeeType, order.BoolIce, order.TableNum)`（3s 超时返回 `deliver_timeout`）
- 资源互斥：四类设备各自一个信号量（容量=1），以保证同一资源同一时刻仅被一个订单占用；保持流水线并行（不同订单可在不同阶段同时工作）。

## 代码改动（文件结构）
- `smart_gateway/gateway/pipeline/types.go`
  - `Order { ID, CoffeeType, BoolIce, TableNum, CreateTime, Status }`
  - `Result { ID, BoolIsDone, FinishTime, ErrorMsg }`
- `smart_gateway/gateway/pipeline/db.go`
  - `interface DB { PollPending, MarkQueued, UpdateResult }`（抽象真实数据库）
  - `SQLDB` 生产实现：读取 `DB_DRIVER/DB_DSN`，使用 `SELECT ... FOR UPDATE SKIP LOCKED` 拉取 `pending`，事务内更新为 `queued`，完成后 `UPDATE ... status=done/error, finish_time, error_msg`
  - 演示实现：`InMemoryDB`（保留，便于本地快速跑通）
- `smart_gateway/gateway/pipeline/queue.go`
  - `interface Queue { PublishOrder, ConsumeOrders, PublishResult, ConsumeResults }`
  - `AMQPQueue` 生产实现：读取 `AMQP_URL/AMQP_QUEUE_ORDERS/AMQP_QUEUE_COMPLETED`，用 JSON 编解码
  - 演示实现：`InMemoryQueue`（保留）
- `smart_gateway/gateway/pipeline/poller.go`
  - 订单搬运工：周期性从 `DB.PollPending(batch)` 拉取，`DB.MarkQueued`，发布至 `queue_orders`
- `smart_gateway/gateway/pipeline/core.go`
  - 流水线核心：消费 `queue_orders`，按步骤调用真实设备；四资源信号量：`grinder/brewer/icemaker/robot`
  - 错误处理：每步错误立即发布 `Result{BoolIsDone=false, ErrorMsg=err}` 到 `queue_completed`
  - 可配置的最小制冰量/出冰量与重试次数（环境变量）
- `smart_gateway/gateway/pipeline/syncer.go`
  - 结果记录员：消费 `queue_completed` 并 `DB.UpdateResult(result)`

## 新 Demo（入口）
- 新增：`smart_gateway/cmd/pipeline_demo/main.go`
  - 加载 `gateway.Config`（现有 `env/envInt` 工具）
  - 加载数据库与队列配置：
    - `DB_DRIVER`（`postgres`/`mysql`）
    - `DB_DSN`（如 `postgres://user:pass@host:port/db?sslmode=disable`）
    - `AMQP_URL`（如 `amqp://user:pass@host:port/`）
    - `AMQP_QUEUE_ORDERS=queue_orders`
    - `AMQP_QUEUE_COMPLETED=queue_completed`
    - `ICE_MIN_STOCK`、`ICE_DISPENSE_AMOUNT`（默认 `200/100`）
  - 构造 `Gateway` → 构造 `pipeline.Core`（注入 `g.grinder/g.coffee/g.ice/g.robot` 与队列）→ 启动 `poller/core/syncer`
  - 打印进度日志与最终结果；若未配置 DB/AMQP 则回退到内存实现（演示模式）

## 配置与环境变量
- 设备：沿用现有 `COFFEE_HOST/COFFEE_PORT`、`GRINDER_HOST/GRINDER_PORT`（默认回退到 `5021`）、`ICE_HOST/ICE_RACK/ICE_SLOT`、`MQTT_HOST/MQTT_PORT`。
- 数据库：`DB_DRIVER`、`DB_DSN`
- RabbitMQ：`AMQP_URL`、`AMQP_QUEUE_ORDERS`、`AMQP_QUEUE_COMPLETED`
- 制冰参数：`ICE_MIN_STOCK`、`ICE_DISPENSE_AMOUNT`

## 并发与资源控制
- 信号量实现：带缓冲 `chan struct{}` 容量=1；`acquire/release` 包裹真实设备调用，保证阶段串行而跨订单并行。
- 送餐超时：复用 `robot.Deliver` 的 3 秒超时行为；失败将写入 `Result.ErrorMsg=deliver_timeout`。

## 错误处理与重试
- 每阶段错误立刻失败并发布结果；可选重试策略（环境变量控制重试次数与退避间隔）。
- 数据库与队列断线自动重连（AMQP/SQL 生产实现加入重连与心跳）。

## 验证计划
- 本地演示：无 DB/AMQP 配置时使用内存实现，跑通 5~10 条订单并输出最终状态。
- 生产联调：提供 `.env` 示例，配置真实 DB 与 RabbitMQ，观察日志与 DB 状态变化。
- 压力测试：提高订单量与并发验证吞吐与互斥正确性。

## 交付物
- `gateway/pipeline/*` 全套代码（三组件 + 接口 + 内存实现）
- `cmd/pipeline_demo/main.go` 新 Demo（加载真实 Gateway 配置）
- 更新 `GO_GUIDE.md` 增加“流水线模式”说明与环境变量清单
- 保留 `DB_Schema.md`（已包含 SQL 与接口说明）

请确认以上方案，确认后我将把流水线接入 `gateway`，并重写 Demo 入口以使用真实设备连接（你的机器配置）。
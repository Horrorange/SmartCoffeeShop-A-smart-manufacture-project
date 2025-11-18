## 目标
- 删除多余的示例入口，保留一条清晰的端到端运行路径（数据库→队列→设备→机器人→结果回写）。
- 修复/避免 MQTT 指令的 JSON 格式错误，消除 `deliver_timeout`。
- 在 Windows 下提供可执行的最小运行步骤与验证方法。

## 问题定位
- `send_order.py` 仅向 PostgreSQL 写入订单并轮询打印，不直接发 MQTT，需同时运行流水线程序。
- 流水线入口：`smart_gateway/cmd/rabbit_sql_pipeline/main.go` 会从 SQL 拉单→发到 RabbitMQ→经真实设备驱动执行→通过 MQTT 通知配送→结果写回 SQL。
- 送餐机器人模拟器只接受合法 JSON（script/delivery_robots/deliveryrobots_sim.py:95），你之前的错误日志显示发送到 `test/delivery_robot/command` 的消息不是合法 JSON，导致未发送 ACK，从而出现 `deliver_timeout`（smart_gateway/gateway/robot.go:52-56）。
- 网关正确发布 JSON（smart_gateway/gateway/robot.go:39-47），常见原因是不同组件连接到了不同的 Broker 或外部手工发布了错误 JSON。

## 精简方案（删除多余 Demo）
- 删除：`smart_gateway/cmd/demo/`（HTTP 顺序调用示例，容易与主流程混淆）。
- 删除：`smart_gateway/cmd/pipeline_demo/`（内存队列+内存DB的演示版，与生产流水线重复）。
- 保留：
  - `smart_gateway/cmd/rabbit_sql_pipeline/`（实际流水线入口）。
  - `smart_gateway/cmd/main.go`（统一 HTTP 入口，可选）。
  - Python 设备模拟器 `script/*` 与 Docker Compose（都必须）。
  - `script/pipeline_demo/send_order.py`（用来向 PostgreSQL 投递订单）。

## 运行步骤（Windows）
- 启动设备模拟与 MQTT Broker：
  - `docker compose -f script\docker_sim\docker_compose.yml up -d --build`
- 安装 Python 依赖：
  - `pip install -r requirements.txt`
- 确认数据库与消息队列：
  - PostgreSQL 与 RabbitMQ 就绪，并设置环境变量：
    - `setx PG_HOST localhost` `setx PG_PORT 5432` `setx PG_DB smartshop` `setx PG_USER postgres` `setx PG_PASS <你的密码>`
    - `setx AMQP_URL amqp://guest:guest@localhost:5672/` `setx AMQP_QUEUE_ORDERS queue_orders` `setx AMQP_QUEUE_COMPLETED queue_completed`
  - 统一 MQTT：若使用 Docker Broker，网关也用 `localhost:1883`；若全部在容器网，网关用 `mqtt-broker:1883`。
- 启动流水线：
  - `cd smart_gateway && go mod tidy && go run cmd\rabbit_sql_pipeline\main.go`
- 投递订单并观察状态变化：
  - `python script\pipeline_demo\send_order.py`（插入 `pending` 订单并每2秒打印一次）。
  - 预期 10~40 秒内看到订单状态从 `pending`→`queued`→`done` 或 `error`。

## 验证与排错
- 验证 MQTT 主题：
  - 网关发布命令（smart_gateway/gateway/robot.go:46）到 `test/delivery_robot/command`。
  - 机器人模拟器收到后立即发 ACK 到 `test/delivery_robot/status`（script/delivery_robots/deliveryrobots_sim.py:98-105）。
- 若仍超时：
  - 确认双方连接的是同一 Broker（容器内用 `mqtt-broker`，宿主用 `localhost:1883`）。
  - 用内置测试脚本发送合法 JSON：`python test\delivery_robot\delivery_robot_test.py`（确保 payload 来自 `json.dumps`）。
  - 如用命令行测试，PowerShell 正确示例：
    - `mosquitto_pub -h localhost -t test/delivery_robot/command -m '{"order_id":999,"coffee_type":"TEST","need_ice":false,"table_number":99}'`
- 查看错误来源：机器人日志有“不是有效的JSON格式”的提示（script/delivery_robots/deliveryrobots_sim.py:117-120），一旦修复为合法 JSON，将立刻发送 ACK，网关不再出现 `deliver_timeout`。

## 代码层面微调（如需要）
- 将网关配送等待从 10 秒改为更宽松或打印收到的原始 ACK（smart_gateway/gateway/robot.go:52-56），便于现场诊断。
- 可将 `order_id` 改为传入真实订单 ID，以便在日志中一致对齐（smart_gateway/gateway/robot.go:39-45 与 pipeline/core.go:54-61）。

## 下一步
- 我将按上述方案删除两个 Demo 目录，保留并跑通 RabbitSQL 流水线，统一 MQTT 配置，验证数据库状态与机器人 ACK。请确认后我来执行。
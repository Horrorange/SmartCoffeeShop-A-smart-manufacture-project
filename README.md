Smart Coffee Shop — 智能制造与网关流水线

一个面向智能制造场景的端到端演示系统：设备模拟（磨豆机、咖啡机、制冰机、送餐机器人）+ 统一网关（HTTP/MQTT/Modbus/S7/TCP）+ 流水线（PostgreSQL + RabbitMQ）。

目标
- 跑通“订单入库 → 队列发布 → 设备执行 → 配送回执 → 结果回写”的闭环。
- 提供清晰的技术架构、数据库结构、消息协议与运行方法。

架构总览
- 设备模拟（Python）：`script/grinder`（Modbus TCP）、`script/coffeemachine`（TCP）、`script/ice_maker`（S7/snap7）、`script/delivery_robots`（MQTT 客户端）
- 网关与流水线（Go）：`smart_gateway/gateway`（设备驱动、统一 HTTP、MQTT 交互）、`smart_gateway/gateway/pipeline`（DB/队列/核心处理）
- 中间件：PostgreSQL（订单）、RabbitMQ（订单与结果队列）、Mosquitto（MQTT Broker）

目录结构
- `smart_gateway/cmd/main.go`：统一 HTTP 入口（`POST /cmd`）
- `smart_gateway/cmd/rabbit_sql_pipeline/main.go`：RabbitMQ + PostgreSQL 流水线入口
- `smart_gateway/gateway/*.go`：设备驱动、统一命令与 HTTP 处理
- `smart_gateway/gateway/pipeline/*.go`：订单表访问、队列接口、AMQP、核心工序、同步器
- `script/docker_sim/docker_compose.yml`：一键启动设备与 MQTT Broker
- `script/pipeline_demo/send_order.py`：向 PostgreSQL 插入订单并轮询结果
- `test/*`：各设备测试脚本（MQTT 指令、磨豆机、咖啡机、制冰机）

技术栈
- Go：`github.com/rabbitmq/amqp091-go`、`github.com/lib/pq`、`github.com/eclipse/paho.mqtt.golang`
- Python：`paho-mqtt`、`colorlog`、`python-snap7`
- 中间件：PostgreSQL、RabbitMQ（`http://localhost:15672` 管理，默认 `guest/guest`）、Eclipse Mosquitto（MQTT）

统一 HTTP 端点
- 地址：`http://localhost:9090/cmd`
- 载体：`smart_gateway/gateway/types.go:18-26`
- 请求示例：`{"device":"delivery_robots","action":"deliver","coffee_type":"LATTE","need_ice":true,"table_number":8}`
- 启动：`smart_gateway/cmd/main.go:25-27`

MQTT 协议
- 命令主题：`test/delivery_robot/command`
- ACK/状态主题：`test/delivery_robot/status`
- 机器人模拟器处理：`script/delivery_robots/deliveryrobots_sim.py:84-116`
- 网关发布命令：`smart_gateway/gateway/robot.go:39-47`
- 有效 JSON 示例：`{"order_id":999,"coffee_type":"TEST","need_ice":false,"table_number":99}`

数据库结构（PostgreSQL）
- 表：`orders`
- 建表 SQL：`smart_gateway/gateway/pipeline/sqldb.go:31-45`
-  字段：`id BIGSERIAL`、`coffee_type VARCHAR(32)`、`bool_ice BOOLEAN`、`table_num INT`、`status VARCHAR(16)`、`create_time TIMESTAMPTZ`、`finish_time TIMESTAMPTZ`、`error_msg TEXT`
- 参考文档：`DB_Schema.md`

队列与流水线
- 队列：RabbitMQ（AMQP）`smart_gateway/gateway/pipeline/amqpqueue.go`
- 队列名：`AMQP_QUEUE_ORDERS=queue_orders`、`AMQP_QUEUE_COMPLETED=queue_completed`
- 轮询发布：`OrderPoller`（`smart_gateway/gateway/pipeline/poller.go:16-35`）
- 核心工序：`Core`（磨豆→制作→出冰→配送）（`smart_gateway/gateway/pipeline/core.go:52-61`）
- 结果写回：`ResultSyncer`（`smart_gateway/gateway/pipeline/syncer.go:11-19`）

服务与端口
- 磨豆机：主机 `5021/5022` → 容器 `502`
- 咖啡机：主机 `8888`
- 制冰机：主机 `102`
- MQTT Broker：主机 `1883`
- RabbitMQ：主机 `5672`（AMQP）、`15672`（管理界面）
- 网关 HTTP：`9090`

环境变量
- 设备：`COFFEE_HOST`、`COFFEE_PORT`、`GRINDER_HOST`、`GRINDER_PORT`、`ICE_HOST`、`ICE_RACK`、`ICE_SLOT`、`MQTT_HOST`、`MQTT_PORT`
- 流水线：`ICE_MIN_STOCK`（默认 200）、`ICE_DISPENSE_AMOUNT`（默认 100）
- PostgreSQL：`PG_HOST`、`PG_PORT`、`PG_DB`、`PG_USER`、`PG_PASS`
- RabbitMQ：`AMQP_URL`、`AMQP_QUEUE_ORDERS`、`AMQP_QUEUE_COMPLETED`

运行方法（Windows/PowerShell）
- 安装：Docker Desktop、Go（1.24+）、Python（3.11+）；在项目根执行 `pip install -r requirements.txt`
- 一键设备与 MQTT Broker：`docker compose -f script\docker_sim\docker_compose.yml up -d --build`
- RabbitMQ：`docker run -d --name rabbitmq --restart unless-stopped -p 5672:5672 -p 15672:15672 rabbitmq:3-management`
- 流水线：
  - 在 `smart_gateway`：`go mod tidy`
  - 设置环境：
    - `setx PG_HOST localhost`
    - `setx PG_PORT 5432`
    - `setx PG_DB smartshop`
    - `setx PG_USER postgres`
    - `setx PG_PASS <你的密码>`
    - `setx AMQP_URL amqp://guest:guest@localhost:5672/`
    - `setx AMQP_QUEUE_ORDERS queue_orders`
    - `setx AMQP_QUEUE_COMPLETED queue_completed`
    - `setx MQTT_HOST localhost`
    - `setx MQTT_PORT 1883`
  - 运行：`go run cmd\rabbit_sql_pipeline\main.go`
- 投递订单：`python script\pipeline_demo\send_order.py`（输出中出现 `done` 即成功）
- 统一 HTTP（可选）：
  - `go run cmd\main.go`
  - `Invoke-WebRequest -Method Post -ContentType 'application/json' -Body '{"device":"grinder","action":"grind"}' -Uri http://localhost:9090/cmd`

验证与示例
- MQTT（PowerShell 转义）：`mosquitto_pub -h localhost -t test/delivery_robot/command -m '{\"order_id\":999,\"coffee_type\":\"TEST\",\"need_ice\":false,\"table_number\":99}'`
- Python 测试：`python test\delivery_robot\delivery_robot_test.py`

常见问题
- deliver_timeout：一般为无效 JSON 或不同 Broker。确保使用 `json.dumps`，统一连接 `localhost:1883`（宿主）/ `mqtt-broker:1883`（容器）。参考 `smart_gateway/gateway/robot.go:52-56`。
- 端口占用：查看 `docker ps`，避免冲突。
- PostgreSQL 连通性：检查 `PG_*` 环境变量并确认库存在；流水线自动建表。
- Windows IPv6：连接异常时用 `127.0.0.1` 代替 `localhost`。

代码位置参考
- HTTP 入口：`smart_gateway/cmd/main.go:25-27`
- 设备驱动：磨豆机 `smart_gateway/gateway/grinder.go`，咖啡机 `smart_gateway/gateway/coffee.go`，制冰机 `smart_gateway/gateway/ice.go`
- 配送机器人 MQTT：`smart_gateway/gateway/robot.go:21-57`
- 流水线核心：`smart_gateway/gateway/pipeline/core.go:52-61`
- 队列 AMQP：`smart_gateway/gateway/pipeline/amqpqueue.go:29-37`
- 数据库访问：`smart_gateway/gateway/pipeline/sqldb.go:48-71`、`smart_gateway/gateway/pipeline/sqldb.go:75-105`

安全提示
- 不要将数据库密码或密钥提交到仓库；通过环境变量配置敏感信息。

许可证与作者
- License: MIT
- 作者：Orange（horrorange@qq.com）
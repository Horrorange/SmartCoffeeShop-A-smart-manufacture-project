# Smart Gateway（Go）运行指南与详细说明

本指南面向 Windows 用户，说明 `smart_gateway` 模块的实际架构、运行方式、统一指令接口与联调方法，并提供常见故障排查与扩展建议。

---

## 1. 环境准备

- Go 版本：建议使用 `go 1.25+`（项目 `go.mod` 当前为 `go 1.25.0`）
- 操作系统：Windows（PowerShell）
- Python（可选，用于设备模拟器与测试）：建议 `Python 3.11+`
- Docker（可选，用于一键联调环境）：`Docker Desktop` + `docker compose`

---

## 2. 代码结构与核心概念

- `smart_gateway/`（Go 模块根目录）
  - `cmd/main.go`：HTTP 服务入口，暴露统一端点 `POST /cmd`
  - `cmd/demo/main.go`：示例客户端，按序调用 `/cmd`
  - `gateway/`：设备适配与统一调度
    - `server.go`：`Gateway` 聚合器与 HTTP 处理逻辑
    - `types.go`：配置结构 `Config`、统一指令 `UnifiedCommand`、响应 `Result`
    - `grinder.go`：磨豆机（Modbus TCP）适配
    - `coffee.go`：咖啡机（TCP）适配
    - `ice.go`：制冰机（S7）适配（`gos7`）
    - `robot.go`：送餐机器人（MQTT 客户端）适配

---

## 3. 安装依赖

在 `smart_gateway` 目录下：

```powershell
cd smart_gateway
go mod tidy
```

- 该命令会自动拉取缺失的 Go 依赖（如 `github.com/google/uuid`、`github.com/robinson/gos7`、`github.com/eclipse/paho.mqtt.golang` 等），并更新 `go.sum`。
- 如网络受限，可临时设置代理：
  - `setx GOPROXY https://goproxy.io,direct`

Python 模拟器依赖（可选，在项目根目录）：

```powershell
cd ..  # 回到仓库根目录
pip install -r requirements.txt
```

---

## 4. 构建与运行

在 `smart_gateway` 目录下：

```powershell
cd smart_gateway
go mod tidy
go run cmd\main.go   # 启动 HTTP 服务（端口 9090）
```

另：示例客户端（需服务已启动）

```powershell
go run cmd\demo\main.go
```

注意：PowerShell 不支持 `&&` 作为语句连接符，请逐条执行命令。

---

## 5. 设备模拟与联调（可选）

### 5.1 手动启动 Python 模拟器

在仓库根目录下：

```powershell
python script\grinder\grinder_sim.py         # 磨豆机（Modbus TCP）默认端口 502（示例映射 5021/5022）
python script\coffeemachine\coffeemachine_sim.py  # 咖啡机（TCP）端口 8888
python script\ice_maker\icemaker_sim.py      # 制冰机（S7）端口 102
```

MQTT 测试（发送配送指令）：

```powershell
python test\delivery_robot\delivery_robot_test.py
```

### 5.2 Docker 一键联调

```powershell
docker compose -f script\docker_sim\docker_compose.yml up -d
```

- 包含磨豆机（两个实例）、咖啡机、制冰机、送餐机器人（MQTT 客户端）以及 MQTT Broker（Eclipse Mosquitto）。
- Broker 端口：`1883`（开发模式允许匿名访问，配置见 `script/docker_sim/mosquitto.conf`）。

---

## 6. 统一端点与设备适配

统一端点：`POST http://localhost:9090/cmd`

请求体示例：

```json
{ "device": "grinder", "action": "grind" }
{ "device": "coffeemachine", "action": "make", "coffee_type": "LATTE" }
{ "device": "ice_maker", "action": "produce", "ice_amount": 500 }
{ "device": "ice_maker", "action": "dispense", "ice_amount": 200 }
{ "device": "delivery_robots", "action": "deliver", "coffee_type": "LATTE", "need_ice": true, "table_number": 8 }
```

响应：

- 成功：`200 {"ok":true,"message":"..."}`
- 失败：`400 {"ok":false,"message":"..."}`（例如 `dial tcp ... refused`、`grind_timeout`、`deliver_timeout`）

### 6.4 送餐机器人（MQTT 客户端）

- 驱动类型：`DeliveryRobotMqttDriver`
- 构造函数：`NewDeliveryRobotMqttDriver(broker, clientID, commandTopic, statusTopic)`
- 发布（命令话题）：`task.Command == "DELIVER"`，`task.Params` 至少包含 `order_id`（若缺省会自动生成 UUID）
- 回执（状态话题）示例负载：

```json
{
  "order_id": "<uuid>",
  "status": "DELIVERY_COMPLETE" | "DELIVERY_FAILED"
}
```

- 驱动内部会在 `statusTopic` 上接收消息并将结果投递到对应 `order_id` 的“信箱”（channel），`ExecuteTask` 会阻塞等待回执或超时（默认 30s）。

---

## 7. 配置项（环境变量）

HTTP 服务固定监听 `:9090`。

可用环境变量（含默认）：

- `COFFEE_HOST=localhost`
- `COFFEE_PORT=8888`
- `GRINDER_HOST=localhost`
- `GRINDER_PORT=5021`（若设置为 `502`，服务会自动回退到 `5021`）
- `ICE_HOST=localhost`
- `ICE_RACK=0`
- `ICE_SLOT=2`
- `MQTT_HOST=localhost`
- `MQTT_PORT=1883`

---

## 8. 常见问题与故障排查

- 连接被拒绝（`dial tcp [::1]:502 ... refused`）：
  - 请确认磨豆机模拟器已在主机端口 `5021/5022` 映射到容器 `502`；
  - 未使用 Docker 时，将 `GRINDER_PORT` 设置为 `5021`；
  - Windows 上 `localhost` 解析为 `::1`（IPv6），如网络策略限制可改为 `127.0.0.1`。
- 配送超时：`robot.go` 在 3s 内未收到 `test/delivery_robot/status` 回执将返回 `deliver_timeout`；检查 MQTT Broker（`1883`）与模拟器是否在线。
- PowerShell 命令：不要使用 `&&`，逐条执行。
- 依赖缺失：在 `smart_gateway` 下执行 `go mod tidy`；Python 依赖用 `pip install -r requirements.txt`。
- S7 读写失败：确认 `rack/slot`、DB 区块号与偏移；网络连通性与端口 `102`；可适当增大 `handler.Timeout`。

---

## 9. 未来开发文档（规范与建议）

### 9.1 新增设备驱动的步骤

- 在 `pkg/drivers/` 新建 `<device>_driver.go` 文件，定义结构体并实现 `Device` 接口四个方法：`Connect`、`Disconnect`、`GetStatus`、`ExecuteTask`。
- 在文件顶部加入编译期断言：`var _ Device = (*YourDriver)(nil)`，确保实现完整。
- 统一使用 `DeviceStatus` 与 `Task`，避免跨包循环导入（不要自引用 `smartgateway/pkg/drivers`）。

### 9.2 错误处理与日志

- 返回错误时使用 `fmt.Errorf("...: %w", err)` 进行错误包装。
- 重要步骤打印提示信息，便于联调（后续可替换为结构化日志）。

### 9.3 配置与参数

- 将 host/port/topic 等配置提取为构造函数参数或环境变量（如 `MQTT_HOST/MQTT_PORT`）。
- 为时延敏感的驱动（如 S7/MQTT）支持超时与重试策略。

### 9.4 并发与同步

- 避免在驱动内部长时间阻塞主线程；如需等待回执（MQTT），使用 channel + 超时控制。
- 注意 `map` 并发访问需要加锁（参考 `mqtt_driver.go` 的 `pendingTasks` 与 `requestMutex`）。

### 9.5 测试与联调

- 为每个驱动提供最小可运行示例或集成测试脚本。
- 增加端到端联调说明（与 Python 模拟器/容器配合）。

### 9.6 代码风格

- 保持简洁，避免不必要的抽象；按需补充注释，但不写版权头。
- 只改动与任务相关的代码，避免“顺手修复”不在范围内的问题。

---

## 10. 参考

- `github.com/robinson/gos7`（S7 协议实现，提供 `TCPClientHandler`、`Client` 以及 `Helper` 编解码工具）
- `github.com/eclipse/paho.mqtt.golang`（MQTT 客户端）
- `github.com/goburrow/modbus`（Modbus TCP 客户端）

如需进一步统一或扩展驱动（例如增加状态字段、引入配置文件、完善日志与测试），可在此文档基础上迭代。
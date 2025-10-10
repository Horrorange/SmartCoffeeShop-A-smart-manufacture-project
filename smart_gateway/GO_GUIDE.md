# Smart Gateway（Go）运行指南与未来开发文档

本指南面向 Windows 用户，详细说明 `smart_gateway` 模块及其 `pkg/drivers` 下各 Go 驱动的运行方法，并给出后续开发规范与建议。

---

## 1. 环境准备

- Go 版本：建议使用 `go 1.25+`（项目 `go.mod` 当前为 `go 1.25.0`）
- 操作系统：Windows（PowerShell）
- Python（可选，用于设备模拟器与测试）：建议 `Python 3.11+`
- Docker（可选，用于一键联调环境）：`Docker Desktop` + `docker compose`

---

## 2. 代码结构与核心概念

- `smart_gateway/`（Go 模块根目录）
  - `main.go`：入口程序，演示如何调用驱动并打印设备状态
  - `pkg/drivers/`：驱动实现与公共接口
    - `interface.go`：统一设备接口与模型
      - `type Task struct { Command string; Params map[string]interface{} }`
      - `type DeviceStatus struct { IsIdle bool; ErrorCode int; Inventory map[string]int }`
      - `type Device interface { Connect() error; Disconnect() error; GetStatus() (DeviceStatus, error); ExecuteTask(task Task) error }`
    - `modbus_driver.go`：磨豆机（Modbus TCP）示例/实现
    - `tcp_driver.go` / `coffeemachine_driver.go`：咖啡机（TCP）示例/实现
    - `s7_driver.go`：制冰机（Siemens S7）实现，使用 `github.com/robinson/gos7`
    - `mqtt_driver.go`：送餐机器人（MQTT 客户端）实现，使用 `github.com/eclipse/paho.mqtt.golang`

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
go build ./pkg/drivers   # 先确保驱动包可编译
go build .               # 构建可执行文件 smartgateway.exe
.\n+smartgateway.exe         # 运行（或使用 go run .）
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

## 6. 各驱动使用说明

### 6.1 磨豆机（Modbus TCP）

- 驱动类型：`GrinderModbusDriver`
- 常用参数：host、port
- 常用任务（示例）：
  - `Task{ Command: "GRIND", Params: map[string]interface{}{ "grams": 10 } }`
- 状态读取：寄存器映射见模拟器脚本；`DeviceStatus.Inventory["BEANS"]` 可表示豆量等。

### 6.2 咖啡机（TCP）

- 驱动类型：`CoffeeTCPDriver`
- 常用参数：host、port
- 常用任务（示例）：
  - `Task{ Command: "MAKE_COFFEE", Params: map[string]interface{}{ "type": "Americano" } }`
- 状态读取：包含 `IsIdle` / `ErrorCode` / `Inventory`（如水量、杯数等）。

### 6.3 制冰机（S7）

- 驱动类型：`IceMakerS7Driver`
- 构造函数：`NewIceMakerS7Driver(host, port, rack, slot)`
- 内存布局（DB1）：
  - 读取：`stock@0 (INT)`、`status@2 (INT)`
  - 写入：`command@4 (INT)`、`amount@6 (INT)`
- 任务（示例）：
  - `Task{ Command: "MAKE_ICE" }` → 写入 `command=1`
  - `Task{ Command: "DISPENSE_ICE", Params: { "amount_grams": 50 } }` → 先写入 `amount=50`，再写入 `command=3`
- 实现要点：使用 `gos7.Helper.SetValueAt/GetValueAt` 编解码 `uint16`；读写使用 `client.AGReadDB/AGWriteDB`；连接使用 `handler.Connect/Close`。

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

## 7. 在 `main.go` 中使用驱动（示例）

```go
package main

import (
    "smartgateway/pkg/drivers"
)

func main() {
    // 以制冰机为例
    ice := drivers.NewIceMakerS7Driver("localhost", "102", 0, 2)
    if err := ice.Connect(); err != nil { panic(err) }
    defer ice.Disconnect()

    status, err := ice.GetStatus()
    if err != nil { panic(err) }
    _ = status

    // 执行任务
    task := drivers.Task{ Command: "MAKE_ICE", Params: map[string]interface{}{} }
    if err := ice.ExecuteTask(task); err != nil { panic(err) }
}
```

---

## 8. 常见问题与故障排查

- PowerShell 命令拼接：不要使用 `&&`，请逐条执行。
- 端口占用：确保模拟器/容器未重复启动；关闭冲突进程后重试。
- 依赖缺失：在 `smart_gateway` 下执行 `go mod tidy`；Python 依赖用 `pip install -r requirements.txt`。
- MQTT 无回执：确认 `statusTopic` 与模拟器发送的 `order_id` 一致；Broker 已在 `1883` 端口监听；`delivery_robots` 服务已正确连接。
- S7 读写失败：确认 `rack/slot`、DB 区块号与偏移；网络连通性与端口 `102`；尝试增大 `handler.Timeout`。

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
# 🍵 Smart Coffee Shop - 智能咖啡店制造系统

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python 3.7+](https://img.shields.io/badge/python-3.7+-blue.svg)](https://www.python.org/downloads/)
[![Modbus TCP](https://img.shields.io/badge/protocol-Modbus%20TCP-green.svg)](https://modbus.org/)

一个基于工业物联网协议的智能咖啡店制造系统，模拟真实的咖啡制作流程，包含磨粉机、咖啡机和智能代理的完整协作。

## 🎯 项目简介

本项目模拟了一个完整的智能咖啡店制造系统，展示了工业4.0环境下设备间的协作模式：

- **磨粉机模拟器** - 基于Modbus TCP协议的工业设备模拟
- **咖啡机模拟器** - 基于自定义TCP协议的智能设备
- **智能代理** - 边缘计算网关，协调设备间的通信和决策

## ✨ 功能特性

### 🔧 设备模拟
- ✅ **磨粉机模拟器**：Modbus TCP服务器，支持磨粉控制、状态监控、豆量管理
- ✅ **咖啡机模拟器**：TCP服务器，支持多种咖啡制作、原料管理、库存查询
- ✅ **实时状态监控**：设备状态实时更新，支持空闲/工作/故障状态

### 🤖 智能代理
- ✅ **多协议通信**：同时支持Modbus TCP和自定义TCP协议
- ✅ **本地决策**：边缘计算，无需云端即可完成订单处理
- ✅ **原料检查**：智能检查库存，自动判断是否可以制作
- ✅ **设备协调**：协调磨粉机和咖啡机的工作流程

### ☕ 咖啡制作
- ✅ **多种咖啡类型**：支持LATTE、CAPPUCCINO、ESPRESSO、AMERICANO等
- ✅ **原料管理**：牛奶、燕麦奶、抹茶酱、巧克力酱、焦糖糖浆等
- ✅ **自动补货**：原料不足时自动补充
- ✅ **制作时间模拟**：真实的磨粉和制作时间

## 🏗️ 系统架构

```
┌─────────────────┐    Modbus TCP     ┌─────────────────┐
│   磨粉机模拟器    │◄─────────────────►│                 │
│  (grinder_sim)  │     Port 502      │                 │
└─────────────────┘                   │   智能代理       │
                                      │ (agent)         │
┌─────────────────┐   Custom TCP      │                 │
│  咖啡机模拟器     │◄─────────────────►│                 │
│(coffeemachine)  │    Port 8888      └─────────────────┘
└─────────────────┘
```

### 协议映射

| 设备 | 协议 | 端口 | 寄存器/命令 |
|------|------|------|-------------|
| 磨粉机 | Modbus TCP | 502 | CMD_REG(0), STATUS_REG(1), BEAN_LEVEL_REG(2) |
| 咖啡机 | Custom TCP | 8888 | MAKE, STATUS, REFILL |

## 🚀 快速开始

### 环境要求

- Python 3.7+
- pip 包管理器

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project.git
cd SmartCoffeeShop-A-smart-manufacture-project

# 安装依赖
pip install -r requirements.txt
```

### 启动系统

#### 1. 启动磨粉机模拟器

```bash
python3 script/grinder/grinder_sim.py
```

#### 2. 启动咖啡机模拟器

```bash
python3 script/coffeemachine/coffeemachine_sim.py
```

#### 3. 运行智能代理测试

```bash
python3 test/coffeemachine/coffeemachine_agent.py
```

### 预期输出

系统启动后，你将看到：

1. **磨粉机模拟器**：Modbus服务器启动，监听端口502
2. **咖啡机模拟器**：TCP服务器启动，监听端口8888
3. **智能代理**：连接两个设备，开始处理咖啡订单

## 📋 使用示例

### 基本咖啡制作流程

```python
# 智能代理处理订单的完整流程
def handle_order(coffee_type):
    # 1. 检查配方和库存
    # 2. 向磨粉机发送磨粉指令
    # 3. 等待磨粉完成（5秒）
    # 4. 向咖啡机发送制作指令
    # 5. 等待咖啡制作完成
    # 6. 返回制作结果
```

### 支持的咖啡类型

| 咖啡类型 | 所需原料 | 制作时间 |
|----------|----------|----------|
| ESPRESSO | 无 | 5-10秒 |
| AMERICANO | 无 | 5-10秒 |
| LATTE | 牛奶 x3 | 5-10秒 |
| CAPPUCCINO | 牛奶 x3 | 5-10秒 |
| OAT LATTE | 燕麦奶 x3 | 5-10秒 |
| MOCHA | 牛奶 x2, 巧克力酱 x1 | 5-10秒 |

## 🔧 配置说明

### 磨粉机配置

```python
# 寄存器地址配置
CMD_REG = 0         # 命令寄存器
STATUS_REG = 1      # 状态寄存器
BEAN_LEVEL_REG = 2  # 豆量寄存器
ERROR_CODE_REG = 3  # 错误代码寄存器
```

### 咖啡机配置

```python
# 库存配置
MAX_STORAGE = 50    # 最大库存
inventory = {
    "MILK": 50,
    "OAT_MILK": 50,
    "MATCHA_SAUCE": 50,
    "CHOCOLATE_SAUCE": 50,
    "CARAMEL_SYRUP": 50,
}
```

## 🧪 测试

### 运行单元测试

```bash
# 测试磨粉机
python3 test/grinder/client_test.py

# 测试咖啡机代理
python3 test/coffeemachine/coffeemachine_agent.py
```

### 测试场景

1. **正常制作流程**：测试完整的咖啡制作流程
2. **原料不足**：测试库存不足时的自动补货
3. **设备故障**：测试设备异常时的错误处理
4. **并发订单**：测试多个订单的处理能力

## 📁 项目结构

```
SmartCoffeeShop-A-smart-manufacture-project/
├── README.md                    # 项目说明文档
├── LICENSE                      # MIT许可证
├── requirements.txt             # Python依赖列表
├── setup.py                     # 项目安装配置
├── .gitignore                   # Git忽略规则
├── script/                      # 核心模拟器脚本
│   ├── grinder/
│   │   └── grinder_sim.py       # 磨粉机模拟器
│   └── coffeemachine/
│       └── coffeemachine_sim.py # 咖啡机模拟器
└── test/                        # 测试脚本
    ├── grinder/
    │   └── client_test.py       # 磨粉机客户端测试
    └── coffeemachine/
        └── coffeemachine_agent.py # 智能代理测试
```

## 🛠️ 故障排除

### 常见问题

1. **连接被拒绝**
   ```bash
   # 确保模拟器已启动
   python3 script/grinder/grinder_sim.py
   python3 script/coffeemachine/coffeemachine_sim.py
   ```

2. **端口被占用**
   ```bash
   # 检查端口占用情况
   lsof -i :502  # 磨粉机端口
   lsof -i :8888 # 咖啡机端口
   ```

3. **依赖包缺失**
   ```bash
   # 重新安装依赖
   pip install -r requirements.txt
   ```

4. **磨粉机不响应**
   - 检查Modbus TCP连接
   - 确认寄存器地址配置正确
   - 查看磨粉机日志输出

## 🤝 贡献指南

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

### 贡献方式

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 👨‍💻 作者

- **Orange** - *项目创建者* - [horrorange@qq.com](mailto:horrorange@qq.com)

## 🙏 致谢

- 感谢 [pyModbusTCP](https://github.com/sourceperl/pyModbusTCP) 提供的Modbus TCP库
- 感谢 [colorlog](https://github.com/borntyping/python-colorlog) 提供的彩色日志支持

## 🔗 相关链接

- [Modbus协议官方文档](https://modbus.org/)
- [工业物联网最佳实践](https://www.iiconsortium.org/)
- [边缘计算架构指南](https://www.edgecomputing.org/)

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！

## 🐳 Docker 使用指南

本项目提供基于 Docker Compose 的一键式设备模拟与联调环境，覆盖磨豆机、咖啡机、制冰机、送餐机器人以及 MQTT Broker。

### 目录结构（Docker）
- `script/docker_sim/docker_compose.yml`：Compose 主文件，定义所有服务与网络
- `script/docker_sim/mosquitto.conf`：MQTT Broker 配置（开发环境开启匿名访问）
- `script/grinder/`：磨豆机镜像构建上下文与模拟器脚本
- `script/coffeemachine/`：咖啡机镜像构建上下文与模拟器脚本
- `script/ice_maker/`：制冰机镜像构建上下文与模拟器脚本
- `script/delivery_robots/`：送餐机器人镜像构建上下文与模拟器脚本
- `test/`：各设备的客户端/联调测试脚本

### 环境准备
- 安装 Docker Desktop（推荐 24+，Compose v2）
- 打开终端并切换到项目根目录 `SmartCoffeeShop-A-smart-manufacture-project`
- Windows 用户使用 PowerShell 执行命令

### 一键启动
- 构建并后台启动全部服务：
  - `docker compose -f script/docker_sim/docker_compose.yml up -d --build`
- 查看运行状态与端口：
  - `docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"`

### 服务与端口
- `grinder1`：主机 `5021` → 容器 `502`（Modbus TCP）
- `grinder2`：主机 `5022` → 容器 `502`（Modbus TCP）
- `coffee_machine`：主机 `8888` → 容器 `8888`（自定义 TCP）
- `ice_maker`：主机 `102` → 容器 `102`（S7/snap7）
- `mqtt-broker`：主机 `1883` → 容器 `1883`（MQTT Broker）
- `delivery_robots`：MQTT 客户端，不暴露端口，连接到 `mqtt-broker`

### 常用操作
- 启动指定服务：
  - `docker compose -f script/docker_sim/docker_compose.yml up -d grinder1 grinder2`
- 重新构建并启动：
  - `docker compose -f script/docker_sim/docker_compose.yml up -d --build <service>`
- 查看日志：
  - `docker logs --tail=100 -f <container>`（如 `delivery_robots`、`mqtt-broker`）
- 停止并清理：
  - `docker compose -f script/docker_sim/docker_compose.yml down`
  - 移除卷：`docker compose -f script/docker_sim/docker_compose.yml down -v`

### MQTT 配置说明
- Broker 镜像：`eclipse-mosquitto:2`
- 配置挂载：`script/docker_sim/mosquitto.conf`（开发模式）
  - `listener 1883 0.0.0.0`
  - `allow_anonymous true`
- 送餐机器人环境变量（已在 Compose 注入）：
  - `MQTT_HOST=mqtt-broker`
  - `MQTT_PORT=1883`

### 验证与测试
- 磨豆机（Modbus TCP）：
  - 连接 `localhost:5021` 或 `localhost:5022`
  - 示例：`python test/grinder/client_test.py`
- 咖啡机（TCP）：
  - 连接 `localhost:8888`（`telnet` 或 `nc`）
- 制冰机（S7）：
  - 使用 S7 客户端连接 `localhost:102` 进行区块读写测试
- 送餐机器人（MQTT）：
  - 连接 `localhost:1883`
  - 发布到 `test/delivery_robot/command`，订阅 `test/delivery_robot/status`
  - 示例消息：`{"order_id": 1, "coffee_type": "LATTE", "need_ice": false, "table_number": 3}`

### 多实例扩展
- 已在 Compose 中提供两台磨豆机：`grinder1` 与 `grinder2`
- 如需更多实例，可复制服务块并映射新的主机端口（例如 `5023:502`）

### 故障排查
- 端口冲突（port is already allocated）：
  - `docker rm -f <container>` 后重启对应服务
- MQTT 连接被拒绝：
  - 确认 `mqtt-broker` 已启动并加载 `mosquitto.conf`
  - 确认 `delivery_robots` 使用 `MQTT_HOST=mqtt-broker`
- `python-snap7` 加载失败：
  - 使用 `python:3.11-slim` 作为 `ice_maker` 基础镜像（已配置）

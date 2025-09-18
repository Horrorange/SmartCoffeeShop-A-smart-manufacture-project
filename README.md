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

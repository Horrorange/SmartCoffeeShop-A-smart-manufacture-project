# Smart Coffee Shop - 智能咖啡店制造项目

一个基于Modbus TCP协议的智能咖啡磨豆机模拟系统，用于演示工业物联网设备通信。

## 项目简介

本项目模拟了一个智能咖啡店的磨豆机系统，包含：
- 磨豆机模拟器服务端（grinder_sim.py）
- 客户端测试程序（client_test.py）
- 基于Modbus TCP协议的设备通信

## 功能特性

- ✅ Modbus TCP服务器模拟磨豆机设备
- ✅ 实时状态监控（空闲/磨粉中/故障）
- ✅ 咖啡豆库存管理
- ✅ 设备占用检测
- ✅ 错误处理和日志记录

## 安装依赖

```bash
# 克隆项目
git clone <your-repo-url>
cd SmartCoffeeShop-A-smart-manufacture-project

# 安装Python依赖
pip install -r requirements.txt
```

## 使用方法

### 1. 启动磨豆机模拟器

```bash
python3 script/grinder/grinder_sim.py
```

服务器将在localhost:502端口启动，初始状态：
- 状态：空闲
- 咖啡豆库存：100%
- 错误代码：无

### 2. 运行客户端测试

```bash
# 连接本地服务器
python3 script/grinder/client_test.py

# 连接远程服务器
python3 script/grinder/client_test.py --host 192.168.1.100
```

## 系统架构

### Modbus寄存器映射

| 寄存器地址 | 名称 | 描述 | 取值范围 |
|-----------|------|------|----------|
| 0 | CMD_REG | 命令寄存器 | 0=无命令, 1=开始磨粉 |
| 1 | STATUS_REG | 状态寄存器 | 0=空闲, 1=磨粉中, 2=故障 |
| 2 | BEAN_LEVEL_REG | 咖啡豆库存 | 0-100% |
| 3 | ERROR_CODE_REG | 错误代码 | 0=无错误, 1=咖啡豆不足 |

### 工作流程

1. 客户端连接到Modbus服务器
2. 读取当前磨豆机状态
3. 检查设备是否被占用
4. 发送磨粉启动命令
5. 服务器处理命令并更新状态
6. 模拟磨粉过程（5秒）
7. 更新咖啡豆库存并返回空闲状态

## 项目结构

```
SmartCoffeeShop-A-smart-manufacture-project/
├── LICENSE                    # MIT许可证
├── README.md                  # 项目说明文档
├── requirements.txt           # Python依赖列表
├── script/
│   └── grinder/
│       ├── grinder_sim.py     # 磨豆机模拟器服务端
│       └── client_test.py     # 客户端测试程序
└── test/
    └── grinder/
        └── client_test.py     # 测试版本客户端
```

## 开发者信息

- **作者**: Orange
- **邮箱**: horrorange@qq.com
- **最后修改**: 2025-09-17
- **许可证**: MIT License

## 技术栈

- **Python 3.x**
- **pyModbusTCP** - Modbus TCP通信库
- **logging** - 日志记录
- **argparse** - 命令行参数解析

## 故障排除

### 常见问题

1. **连接被拒绝**
   - 确保磨豆机模拟器已启动
   - 检查端口502是否被占用

2. **设备占用提示**
   - 等待当前磨粉任务完成
   - 或重启模拟器重置状态

3. **咖啡豆不足错误**
   - 重启模拟器恢复100%库存
   - 或修改代码中的初始库存值

## 贡献指南

欢迎提交Issue和Pull Request来改进项目！

## 许可证

本项目采用MIT许可证 - 详见 [LICENSE](LICENSE) 文件
This is a smart coffee shop powered by the smart manufacture technology. It uses a blend platform that enables the devices using different protocols to communicate with the server.

package drivers

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

// --- 1. 常量定义 ---
const (
	grinderCmdReg       = 0
	grinderStatusReg    = 1
	grinderBeanLevelReg = 2
	grinderErrorCodeReg = 3
)

// --- 2. 驱动结构体定义 ---
// 这是一个GrinderModbusDriver的结构体，里面封装了通信地址，和通信所需要用到的Modbus客户端实例
type GrinderModbusDriver struct {
	host    string // IP地址
	port    string // 端口
	slaveID byte   // Modbus从站ID

	handler *modbus.TCPClientHandler // 底层的TCP连接处理器
	client  modbus.Client            // Modbus客户端实例
}

// 把一个空的磨豆器指针赋值给Device接口，确保实现了所有方法，否则编译会报错
var _ Device = (*GrinderModbusDriver)(nil)

// --- 3. 构造函数 ---

// NewGrinderModbusDriver 是创建磨粉机驱动实例的工厂函数。
// 可以理解为是类的创建，会返回一个指针，指向一个GrinderModbusDriver的实例
func NewGrinderModbusDriver(host, port string, slaveID byte) *GrinderModbusDriver {
	// 初始化handler，设置连接参数和超时时间
	// 增加了一个handler示例，用于处理Modbus TCP连接
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%s", host, port))
	handler.Timeout = 10 * time.Second // 10秒超时
	handler.SlaveId = slaveID

	// 创建client实例
	client := modbus.NewClient(handler)

	return &GrinderModbusDriver{
		host:    host,
		port:    port,
		slaveID: slaveID,
		handler: handler,
		client:  client,
	}
}

// --- 4. 接口方法实现 ---

// Connect 实现了Device接口的Connect方法。
// 负责建立与Modbus服务器的TCP连接。
func (d *GrinderModbusDriver) Connect() error {
	// 调用底层handler的Connect方法，如果链接发生错误，err会返回一个错误，而不是空值
	err := d.handler.Connect()
	if err != nil {
		// 返回一个带有上下文信息的错误，便于调试
		return fmt.Errorf("failed to connect to grinder modbus server at %s:%s: %w", d.host, d.port, err)
	}
	fmt.Printf("Grinder Modbus driver connected to %s:%s\n", d.host, d.port)
	return nil
}

// Disconnect 实现了Device接口的Disconnect方法。
// 负责关闭TCP连接。
func (d *GrinderModbusDriver) Disconnect() error {
	// 调用底层handler的Close方法
	d.handler.Close()
	fmt.Println("Grinder Modbus driver disconnected.")
	return nil
}

// GetStatus 实现了Device接口的GetStatus方法。
// 它从设备读取多个寄存器，并将结果翻译成标准的DeviceStatus模型。
func (d *GrinderModbusDriver) GetStatus() (DeviceStatus, error) {
	// Modbus的 ReadHoldingRegisters(起始地址, 数量)
	// 我们需要从地址1开始，连续读取3个寄存器 (状态, 库存, 故障码)
	results, err := d.client.ReadHoldingRegisters(grinderStatusReg, 3)
	if err != nil {
		// 创建一个空的status对象用于返回
		emptyStatus := DeviceStatus{}
		return emptyStatus, fmt.Errorf("failed to read grinder status registers: %w", err)
	}

	// ReadHoldingRegisters 返回的是一个[]byte，每2个字节代表一个寄存器的值 (uint16)。
	// 我们需要将字节流解析成数字。
	if len(results) < 6 {
		emptyStatus := DeviceStatus{}
		return emptyStatus, fmt.Errorf("invalid response length from modbus server: got %d bytes, expected 6", len(results))
	}

	statusVal := binary.BigEndian.Uint16(results[0:2])    // 寄存器1：状态
	beanLevelVal := binary.BigEndian.Uint16(results[2:4]) // 寄存器2：库存
	errorCodeVal := binary.BigEndian.Uint16(results[4:6]) // 寄存器3：故障码

	// 将读取到的原始值，翻译成我们统一的、标准化的DeviceStatus结构体
	status := DeviceStatus{
		IsIdle:    statusVal == 0,
		ErrorCode: int(errorCodeVal),
		Inventory: map[string]int{
			"bean": int(beanLevelVal),
		},
	}

	return status, nil
}

// ExecuteTask 实现了Device接口的ExecuteTask方法。
// 它将一个通用的Task命令，翻译成具体的Modbus写寄存器操作。
func (d *GrinderModbusDriver) ExecuteTask(task Task) error {
	// 使用 switch 处理不同的命令
	switch task.Command {
	case "START":
		// 对应 Python 模拟器中的 command == 1
		_, err := d.client.WriteSingleRegister(grinderCmdReg, 1)
		if err != nil {
			return fmt.Errorf("failed to execute 'START' command on grinder: %w", err)
		}
		return nil

	case "REFILL":
		// 对应 Python 模拟器中的 command == 2
		_, err := d.client.WriteSingleRegister(grinderCmdReg, 2)
		if err != nil {
			return fmt.Errorf("failed to execute 'REFILL' command on grinder: %w", err)
		}
		return nil

	default:
		// 如果是未知的命令，返回错误
		return fmt.Errorf("unsupported command for grinder: '%s'", task.Command)
	}
}

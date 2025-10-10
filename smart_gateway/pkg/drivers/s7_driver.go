// pkg/drivers/s7_driver.go

package drivers

import (
	"fmt"

	"github.com/robinson/gos7"
)

// --- 1. S7 PLC (制冰机) 内存布局定义 ---
//
// 所有数据都存放在数据块 DB 1 中。
const (
	// 数据块编号
	iceMakerDBNumber = 1

	// 读取地址偏移 (单位: byte)
	iceMakerStockOffset  = 0 // INT, 当前冰块库存 (克)
	iceMakerStatusOffset = 2 // INT, 设备状态 (0=待机, 1=制冰中, 2=出冰中, 3=故障)

	// 写入地址偏移 (单位: byte)
	iceMakerCommandOffset = 4 // INT, 网关指令 (1=制冰, 3=取冰)
	iceMakerAmountOffset  = 6 // INT, 本次取冰量 (克)
)

// --- 2. 驱动结构体定义 ---
type IceMakerS7Driver struct {
	client  gos7.Client // gos7 客户端实例
	handler *gos7.TCPClientHandler
}

// 编译时检查，确保 IceMakerS7Driver 实现了 Device 接口
var _ Device = (*IceMakerS7Driver)(nil)

// --- 3. 构造函数 ---
//
// NewIceMakerS7Driver 创建一个新的制冰机驱动实例。
// rack 和 slot 通常对于S7-1200/1500等型号是固定的。
func NewIceMakerS7Driver(host, port string, rack, slot int) *IceMakerS7Driver {
	handler := gos7.NewTCPClientHandler(fmt.Sprintf("%s:%s", host, port), rack, slot)
	// gos7库的默认超时是2秒，对于模拟器足够了
	// handler.Timeout = 5 * time.Second
	// handler.IdleTimeout = 5 * time.Second

	client := gos7.NewClient(handler)

	return &IceMakerS7Driver{
		client:  client,
		handler: handler,
	}
}

//
// --- 4. 接口方法实现 ---
//

// Connect 负责建立与S7 PLC的连接。
func (d *IceMakerS7Driver) Connect() error {
	if err := d.handler.Connect(); err != nil {
		return fmt.Errorf("failed to connect to S7 PLC (ice maker): %w", err)
	}
	fmt.Println("S7 PLC (ice maker) driver connected.")
	return nil
}

// Disconnect 负责断开与S7 PLC的连接。
func (d *IceMakerS7Driver) Disconnect() error {
	defer fmt.Println("S7 PLC (ice maker) driver disconnected.")
	return d.handler.Close()
}

// GetStatus 从DB1中读取数据，并转换为标准状态模型。
func (d *IceMakerS7Driver) GetStatus() (DeviceStatus, error) {
	emptyStatus := DeviceStatus{Inventory: make(map[string]int)}

	// 我们需要从地址0开始，读取4个字节，以同时获取库存(INT, 2字节)和状态(INT, 2字节)
	buffer := make([]byte, 4)
	if err := d.client.AGReadDB(iceMakerDBNumber, 0, 4, buffer); err != nil {
		return emptyStatus, fmt.Errorf("failed to read DB1 from ice maker: %w", err)
	}

	var helper gos7.Helper
	var stock uint16
	var statusVal uint16
	helper.GetValueAt(buffer, 0, &stock)
	helper.GetValueAt(buffer, 2, &statusVal)

	status := DeviceStatus{
		IsIdle:    statusVal == 0, // 0=待机
		ErrorCode: 0,
		Inventory: map[string]int{
			"ICE": int(stock), // 将库存存入map
		},
	}
	if statusVal == 3 { // 3=故障
		status.ErrorCode = 1
	}

	return status, nil
}

// ExecuteTask 将通用任务翻译成对DB1的具体写操作。
func (d *IceMakerS7Driver) ExecuteTask(task Task) error {
	switch task.Command {
	case "MAKE_ICE":
		// 制冰指令的值为 1
		commandValue := 1
		buffer := make([]byte, 2) // 一个INT是2个字节
		var helper gos7.Helper
		helper.SetValueAt(buffer, 0, uint16(commandValue))

		// 将指令写入DB1的指令地址
		if err := d.client.AGWriteDB(iceMakerDBNumber, iceMakerCommandOffset, 2, buffer); err != nil {
			return fmt.Errorf("failed to execute 'MAKE_ICE' command: %w", err)
		}
		// 指令写入后，PLC会自动开始工作并重置指令位，所以驱动的任务就完成了。
		return nil

	case "DISPENSE_ICE":
		// 取冰任务需要两个参数：取冰量 和 取冰指令

		// 1. 获取取冰量
		amount, ok := task.Params["amount_grams"].(int)
		if !ok || amount <= 0 {
			return fmt.Errorf("task 'DISPENSE_ICE' requires a positive integer 'amount_grams' parameter")
		}

		// 2. 准备写入取冰量的数据
		amountBuffer := make([]byte, 2)
		var helper gos7.Helper
		helper.SetValueAt(amountBuffer, 0, uint16(amount))
		if err := d.client.AGWriteDB(iceMakerDBNumber, iceMakerAmountOffset, 2, amountBuffer); err != nil {
			return fmt.Errorf("failed to write dispense amount: %w", err)
		}

		// 3. 准备并写入取冰指令 (值为3)
		commandValue := 3
		commandBuffer := make([]byte, 2)
		helper.SetValueAt(commandBuffer, 0, uint16(commandValue))
		if err := d.client.AGWriteDB(iceMakerDBNumber, iceMakerCommandOffset, 2, commandBuffer); err != nil {
			return fmt.Errorf("failed to execute 'DISPENSE_ICE' command: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported command for ice maker: '%s'", task.Command)
	}
}

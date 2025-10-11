// pkg/drivers/s7_driver.go

package drivers

import (
	"fmt"
	"time"
	"github.com/robinson/gos7"
)


const (
	iceMakerDBNumber = 1			// 所有内存都在DB1中

	// 读取地址偏移 (单位: byte)
	iceMakerStockOffset  = 0 		// INT, 当前冰块库存 (克)
	iceMakerStatusOffset = 2 		// INT, 设备状态 (0=待机, 1=制冰中, 2=出冰中, 3=故障)	

	// 写入地址偏移 (单位: byte)
	iceMakerCommandOffset = 4 		// INT, 网关指令 (1=制冰, 3=取冰)
	iceMakerAmountOffset  = 6 		// INT, 本次取冰量 (克)
)

// --- 2. 驱动结构体定义 ---
type IceMakerDriver struct {
	client  gos7.Client // gos7 客户端实例
	handler *gos7.TCPClientHandler
}

// 编译时检查，确保 IceMakerDriver 实现了 Device 接口
var _ Device = (*IceMakerDriver)(nil)

// --- 3. 构造函数 ---
func NewIceMakerDriver(host, port string, rack, slot int) *IceMakerDriver {
	handler := gos7.NewTCPClientHandler(fmt.Sprintf("%s:%s", host, port), rack, slot)
	// rack = 0, slot = 1 是默认值，通常不需要修改
	// 设置默认超时时间为5秒
	handler.Timeout = 5 * time.Second
	handler.IdleTimeout = 5 * time.Second

	// 创建 gos7 客户端实例
	client := gos7.NewClient(handler)

	// 返回 IceMakerDriver 实例
	return &IceMakerDriver{
		client:  client,
		handler: handler,
	}
}

// --- 4. 接口方法实现 ---
// Connect 负责建立与S7 PLC的连接。
func (d *IceMakerDriver) Connect() error {
    if err := d.handler.Connect(); err != nil {
        return fmt.Errorf("connect: %w", err)
    }
    fmt.Println("S7 PLC (ice maker) driver connected.")
    return nil
}

// Disconnect 负责断开与S7 PLC的连接。
func (d *IceMakerDriver) Disconnect() error {
	defer fmt.Println("S7 PLC (ice maker) driver disconnected.")
	return d.handler.Close()
}

// GetIceAmount 从DB1中读取当前冰块库存。
func (d *IceMakerDriver) GetIceAmount() int {
    v, err := d.readU16(iceMakerStockOffset)
    if err != nil { return 0 }
    return int(v)
}

// getIceMakerStatus 从DB1中读取当前冰块设备状态。
func (d *IceMakerDriver) getIceMakerStatus() int {
    v, err := d.readU16(iceMakerStatusOffset)
    if err != nil { return 0 }
    return int(v)
}

// RefillIce 向DB1写入指令，请求补充冰块库存。
func (d *IceMakerDriver) RefillIce() error {
    // 下发补充指令=1
    if err := d.writeU16(iceMakerCommandOffset, 1); err != nil {
        return fmt.Errorf("write cmd: %w", err)
    }
	// 等待冰块补充完成，用设备状态和冰量判断是否补充完成
	for {
		status := d.getIceMakerStatus()
		iceAmount := d.GetIceAmount()
		if status == 0 && iceAmount > 999{
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (d *IceMakerDriver) Execute(command string) (*ExecutionResult, error) {
    // command = "ICEMAKER:LESS" or "ICEMAKER:MUCH"
    switch command{
    case "ICEMAKER:LESS":
        // 先读取冰块库存
        iceAmount := d.GetIceAmount()
        if iceAmount <= 50{
            if err := d.RefillIce(); err != nil{
                return nil, fmt.Errorf("refill: %w", err)
            }
        }
        // 先写取冰量，再下发取冰指令
        if err := d.writeU16(iceMakerAmountOffset, 50); err != nil {
            return nil, fmt.Errorf("write amount: %w", err)
        }
        if err := d.writeU16(iceMakerCommandOffset, 3); err != nil {
            return nil, fmt.Errorf("write command: %w", err)
        }
        // 等待复位
        if err := d.waitIdleAndReset(); err != nil { return nil, err }
        // 检查取冰是否成功
        newAmount := d.GetIceAmount()
        if newAmount != maxInt(iceAmount-50, 0) {
            return nil, fmt.Errorf("dispense failed")
        }
        // 取冰完成后，返回取冰结果
        return &ExecutionResult{
            Message:       "DONE",
            InventoryJson: fmt.Sprintf(`{"ice_maker_stock": %d}`, newAmount),
        }, nil
    case "ICEMAKER:MUCH":
        // 先读取冰块库存
        iceAmount := d.GetIceAmount()
        if iceAmount <= 100{
            if err := d.RefillIce(); err != nil{
                return nil, fmt.Errorf("refill: %w", err)
            }
        }
        // 先写取冰量，再下发取冰指令
        if err := d.writeU16(iceMakerAmountOffset, 70); err != nil {
            return nil, fmt.Errorf("write amount: %w", err)
        }
        if err := d.writeU16(iceMakerCommandOffset, 3); err != nil {
            return nil, fmt.Errorf("write command: %w", err)
        }
        // 等待复位
        if err := d.waitIdleAndReset(); err != nil { return nil, err }
        // 检查取冰是否成功
        newAmount := d.GetIceAmount()
        if newAmount != maxInt(iceAmount-70, 0) {
            return nil, fmt.Errorf("dispense failed")
        }
        // 取冰完成后，返回取冰结果
        return &ExecutionResult{
            Message:       "DONE",
            InventoryJson: fmt.Sprintf(`{"ice_maker_stock": %d}`, newAmount),
        }, nil
    default:
        return nil, fmt.Errorf("unknown command: %s", command)
    }
}

// maxInt 返回两个整数的较大值
func maxInt(a, b int) int {
    if a > b { return a }
    return b
}

// --- 简化的读写与等待辅助 ---
func (d *IceMakerDriver) readU16(offset int) (uint16, error) {
    buf := make([]byte, 2)
    if err := d.client.AGReadDB(iceMakerDBNumber, offset, 2, buf); err != nil {
        return 0, fmt.Errorf("read db1@%d: %w", offset, err)
    }
    var h gos7.Helper
    var v uint16
    h.GetValueAt(buf, 0, &v)
    return v, nil
}

func (d *IceMakerDriver) writeU16(offset int, v uint16) error {
    buf := make([]byte, 2)
    var h gos7.Helper
    h.SetValueAt(buf, 0, v)
    if err := d.client.AGWriteDB(iceMakerDBNumber, offset, 2, buf); err != nil {
        return fmt.Errorf("write db1@%d: %w", offset, err)
    }
    return nil
}

func (d *IceMakerDriver) waitIdleAndReset() error {
    for {
        status := d.getIceMakerStatus()
        cmd, err := d.readU16(iceMakerCommandOffset)
        if err != nil { return fmt.Errorf("read cmd: %w", err) }
        if status == 0 && cmd == 0 { return nil }
        time.Sleep(time.Second)
    }
}
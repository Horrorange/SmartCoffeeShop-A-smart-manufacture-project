// pkg/drivers/tcp_driver.go

package drivers

import (
    "bufio"
    "fmt"
    "net"
    "strconv"
    "strings"
    "time"
)

//
// --- 1. 新协议说明 ---
//
// ｜ 指令                     ｜ 成功返回值                            ｜ 失败返回值
// ｜--------------------------|-----------------------------------------|----------------------------------------------------
// ｜ MAKE:<COFFEE_TYPE>       | ACK:MAKE (立刻返回) ... DONE:SUCCESS (制作后返回) | ERROR:INSUFFICIENT_INGREDIENT:<...> 或 ERROR:UNKNOWN_COFFEE_TYPE
// ｜ REFILL:<INGREDIENT/ALL>  | ACK:REFILL_SUCCESS:<INGREDIENT/ALL>     | ERROR:INVALID_INGREDIENT
// ｜ STATUS:INGREDIENTS       | STATUS:INGREDIENTS:MILK=50,OAT_MILK=50... | N/A
//

type CoffeeTCPDriver struct {
	host   string
	port   string
	conn   net.Conn
	reader *bufio.Reader // 将reader作为结构体成员，避免重复创建
}

var _ Device = (*CoffeeTCPDriver)(nil)

func NewCoffeeTCPDriver(host, port string) *CoffeeTCPDriver {
	return &CoffeeTCPDriver{
		host: host,
		port: port,
	}
}

func (d *CoffeeTCPDriver) Connect() error {
	conn, err := net.DialTimeout("tcp", d.host+":"+d.port, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to coffee machine at %s:%s: %w", d.host, d.port, err)
	}
	d.conn = conn
	// 从一个连接创建一个reader是推荐的做法
	d.reader = bufio.NewReader(d.conn)
	fmt.Printf("Coffee Machine TCP driver connected to %s:%s\n", d.host, d.port)
	return nil
}

func (d *CoffeeTCPDriver) Disconnect() error {
	if d.conn != nil {
		fmt.Println("Coffee Machine TCP driver disconnected.")
		return d.conn.Close()
	}
	return nil
}

// readResponse 是一个新的辅助函数，专门用于读取一行响应
func (d *CoffeeTCPDriver) readResponse() (string, error) {
	// 为读取操作设置15秒超时，因为制作咖啡可能需要一些时间
	err := d.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}
	response, err := d.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return strings.TrimSpace(response), nil
}

// sendCommand 是一个新的辅助函数，专门用于发送命令
func (d *CoffeeTCPDriver) sendCommand(command string) error {
	if d.conn == nil {
		return fmt.Errorf("coffee machine is not connected")
	}
	err := d.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}
	_, err = fmt.Fprintf(d.conn, "%s\n", command)
	if err != nil {
		return fmt.Errorf("failed to send command '%s': %w", command, err)
	}
	return nil
}

// GetStatus 完全重写以解析新的库存格式
func (d *CoffeeTCPDriver) GetStatus() (DeviceStatus, error) {
    emptyStatus := DeviceStatus{Inventory: make(map[string]int)}

	if err := d.sendCommand("STATUS:INGREDIENTS"); err != nil {
		return emptyStatus, err
	}

	response, err := d.readResponse()
	if err != nil {
		return emptyStatus, err
	}

	// 预期响应: "STATUS:INGREDIENTS:MILK=50,OAT_MILK=50,..."
	parts := strings.SplitN(response, ":", 3)
	if len(parts) < 3 || parts[0] != "STATUS" || parts[1] != "INGREDIENTS" {
		return emptyStatus, fmt.Errorf("invalid status response: '%s'", response)
	}

	inventoryStr := parts[2]
	inventoryItems := strings.Split(inventoryStr, ",")
	inventoryMap := make(map[string]int)

	for _, item := range inventoryItems {
		kv := strings.Split(item, "=")
		if len(kv) == 2 {
			val, err := strconv.Atoi(kv[1])
			if err == nil {
				inventoryMap[kv[0]] = val
			}
		}
	}

	// 在这个模拟器中，我们通过库存是否为0来判断是否空闲，这是一个简化
	// 实际应用中可能需要一个单独的状态指令
    status := DeviceStatus{
        IsIdle:    true, // 假设能查状态就是空闲的，MAKE任务会阻塞
        ErrorCode: 0,
        Inventory: inventoryMap,
    }

	return status, nil
}

// ExecuteTask 完全重写以处理新的MAKE和REFILL指令
func (d *CoffeeTCPDriver) ExecuteTask(task Task) error {
	switch task.Command {
	case "MAKE":
		coffeeType, ok := task.Params["coffee_type"].(string)
		if !ok || coffeeType == "" {
			return fmt.Errorf("task 'MAKE' requires 'coffee_type' parameter")
		}
		command := fmt.Sprintf("MAKE:%s", strings.ToUpper(coffeeType))
		
		if err := d.sendCommand(command); err != nil {
			return err
		}

		// --- 这是与之前最大的不同：处理两次响应 ---
		// 1. 读取第一次响应 (ACK)
		ackResponse, err := d.readResponse()
		if err != nil {
			return fmt.Errorf("did not receive ACK for MAKE command: %w", err)
		}
		if ackResponse != "ACK:MAKE" {
			// 如果第一次返回的不是ACK，说明是错误（如原料不足）
			return fmt.Errorf("failed to start making coffee: %s", ackResponse)
		}

		// 2. 读取第二次响应 (DONE)
		doneResponse, err := d.readResponse()
		if err != nil {
			return fmt.Errorf("did not receive DONE for MAKE command: %w", err)
		}
		if doneResponse != "DONE:SUCCESS" {
			return fmt.Errorf("making coffee failed: %s", doneResponse)
		}

		// 只有在收到DONE:SUCCESS后，任务才算真正完成
		return nil

	case "REFILL":
		ingredient, ok := task.Params["ingredient"].(string)
		if !ok || ingredient == "" {
			return fmt.Errorf("task 'REFILL' requires 'ingredient' parameter (e.g., 'MILK' or 'ALL')")
		}
		command := fmt.Sprintf("REFILL:%s", strings.ToUpper(ingredient))
		
		if err := d.sendCommand(command); err != nil {
			return err
		}

		// REFILL只需要一次响应
		ackResponse, err := d.readResponse()
		if err != nil {
			return fmt.Errorf("did not receive ACK for REFILL command: %w", err)
		}
		// 检查返回是否以 "ACK:REFILL_SUCCESS" 开头
		if !strings.HasPrefix(ackResponse, "ACK:REFILL_SUCCESS") {
			return fmt.Errorf("refill failed: %s", ackResponse)
		}
		return nil

	default:
		return fmt.Errorf("unsupported command for coffee machine: '%s'", task.Command)
	}
}
// pkg/drivers/mqtt_driver.go

package drivers

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/google/uuid" // 用于生成唯一的订单ID
)

//
// --- 1. 驱动结构体定义 ---
//
type DeliveryRobotMqttDriver struct {
	client       mqtt.Client
	commandTopic string
	statusTopic  string

	// "回执中心"：用于同步等待异步的回调
	// a. Mutex 用于保护 pendingTasks 的并发访问
	requestMutex sync.Mutex
	// b. Map 用于存储 order_id 到其专属“信箱”(channel)的映射
	pendingTasks map[string]chan string
}

// 编译时检查
var _ Device = (*DeliveryRobotMqttDriver)(nil)

//
// --- 2. 构造函数 ---
//
func NewDeliveryRobotMqttDriver(broker, clientID, commandTopic, statusTopic string) *DeliveryRobotMqttDriver {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true)

	driver := &DeliveryRobotMqttDriver{
		commandTopic: commandTopic,
		statusTopic:  statusTopic,
		pendingTasks: make(map[string]chan string),
	}

	// 设置MQTT消息处理回调，这是异步逻辑的核心
	opts.SetDefaultPublishHandler(driver.messageHandler)

	// 设置连接成功后的回调，用于订阅话题
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("MQTT (delivery robot) driver connected.")
		// 订阅状态话题，QoS=1保证消息至少被送达一次
		if token := c.Subscribe(driver.statusTopic, 1, nil); token.Wait() && token.Error() != nil {
			fmt.Printf("Failed to subscribe to status topic: %v\n", token.Error())
		} else {
			fmt.Printf("Subscribed successfully to topic: %s\n", driver.statusTopic)
		}
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Printf("MQTT connection lost: %v\n", err)
	}

	client := mqtt.NewClient(opts)
	driver.client = client
	return driver
}

// messageHandler 是所有订阅消息的总入口
func (d *DeliveryRobotMqttDriver) messageHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received status message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	
	var statusPayload map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &statusPayload); err != nil {
		fmt.Printf("Error decoding status JSON: %v\n", err)
		return
	}

	orderID, ok := statusPayload["order_id"].(string)
	if !ok {
		return
	}
	
	status, _ := statusPayload["status"].(string)

	// --- 核心：将回执投递到专属信箱 ---
	d.requestMutex.Lock()
	defer d.requestMutex.Unlock()

	// 查找这个order_id是否有正在等待的信箱
	if doneChan, found := d.pendingTasks[orderID]; found {
		// 将结果投递回去
		doneChan <- status
		// 完成后关闭信箱
		close(doneChan)
	}
}


//
// --- 3. 接口方法实现 ---
//

func (d *DeliveryRobotMqttDriver) Connect() error {
	if token := d.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	return nil
}

func (d *DeliveryRobotMqttDriver) Disconnect() error {
	d.client.Disconnect(250) // 250ms优雅退出
	fmt.Println("MQTT (delivery robot) driver disconnected.")
	return nil
}

// GetStatus 对于MQTT设备，我们可以通过正在等待的任务数来判断其是否空闲
func (d *DeliveryRobotMqttDriver) GetStatus() (DeviceStatus, error) {
	d.requestMutex.Lock()
	defer d.requestMutex.Unlock()

	// 如果pendingTasks不为空，说明有任务正在执行
	isIdle := len(d.pendingTasks) == 0

    return DeviceStatus{
        IsIdle:    isIdle,
        Inventory: make(map[string]int),
        ErrorCode: 0,
    }, nil
}

// ExecuteTask 实现了同步发布-异步等待的完整逻辑
func (d *DeliveryRobotMqttDriver) ExecuteTask(task Task) error {
	if task.Command != "DELIVER" {
		return fmt.Errorf("unsupported command for delivery robot: '%s'", task.Command)
	}

	// 确保任务参数中有 order_id，如果没有，就生成一个
	if _, ok := task.Params["order_id"]; !ok {
		task.Params["order_id"] = uuid.New().String()
	}
	orderID := task.Params["order_id"].(string)

	// --- 登记“信箱”并准备等待 ---
	doneChan := make(chan string, 1) // buffer=1防止死锁
	d.requestMutex.Lock()
	d.pendingTasks[orderID] = doneChan
	d.requestMutex.Unlock()

	// 任务完成后，无论成功与否，都要从map中清理掉这个信箱
	defer func() {
		d.requestMutex.Lock()
		delete(d.pendingTasks, orderID)
		d.requestMutex.Unlock()
	}()

	// --- 发布任务 ---
	payload, err := json.Marshal(task.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal task params to JSON: %w", err)
	}
	if token := d.client.Publish(d.commandTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish command to topic %s: %w", d.commandTopic, token.Error())
	}
	fmt.Printf("DELIVER command for order [%s] published successfully.\n", orderID)


	// --- 同步等待回执 ---
	// 使用 select 结构可以优雅地处理超时
	select {
	case result := <-doneChan:
		// 从信箱收到了回执
		if result == "DELIVERY_COMPLETE" {
			fmt.Printf("Delivery for order [%s] confirmed complete.\n", orderID)
			return nil // 成功
		}
		return fmt.Errorf("delivery for order [%s] failed with status: %s", orderID, result)

	case <-time.After(30 * time.Second):
		// 等待超过30秒，任务超时
		return fmt.Errorf("delivery task for order [%s] timed out", orderID)
	}
}
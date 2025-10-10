package drivers

// 通用的任务结构体
type Task struct {
	Command string                 // 命令名称
	Params  map[string]interface{} // 参数映射
}

// 设备状态结构体
type DeviceStatus struct {
	IsIdle    bool
	ErrorCode int
	Inventory map[string]int
}

// 设备驱动接口
type Device interface {

	// Connect 连接设备
	Connect() error

	// Disconnect 断开设备连接
	Disconnect() error

	// GetStatus 获取设备状态
	GetStatus() (DeviceStatus, error)

	// ExecuteTask 执行任务
	ExecuteTask(task Task) error
}

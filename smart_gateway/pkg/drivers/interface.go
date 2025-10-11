package drivers

// 完成任务后，driver向决策层传递的通用结构体
type ExecutionResult struct {
	Message       string // 表示完成的文本信号
	InventoryJson string // JSON字符串，包含设备的库存信息
}

type Device interface {
	// 所有的Device必须包含以下三个方法
	Connect() error    // 连接设备，如果有错误则返回
	Disconnect() error // 断开设备连接，如果有错误则返回

	Execute(command string) (*ExecutionResult, error) // 执行命令，返回传递的通用结构体和错误
}

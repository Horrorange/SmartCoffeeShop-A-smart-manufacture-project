package gateway

import "encoding/json"

// Config 描述各设备与MQTT的连接参数，由环境变量填充
type Config struct {
    CoffeeHost string
    CoffeePort int
    GrinderHost string
    GrinderPort int
    IceHost string
    IceRack int
    IceSlot int
    MqttHost string
    MqttPort int
}

// UnifiedCommand 为上层统一指令的载体，由HTTP端点接收
type UnifiedCommand struct {
    Device string `json:"device"`
    Action string `json:"action"`
    CoffeeType string `json:"coffee_type"`
    NeedIce bool `json:"need_ice"`
    TableNumber int `json:"table_number"`
    IceAmount int `json:"ice_amount"`
}

// Result 为统一端点的标准响应
type Result struct {
    Ok bool `json:"ok"`
    Message string `json:"message"`
}

func (r Result) Bytes() []byte {
    b, _ := json.Marshal(r)
    return b
}
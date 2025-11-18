package main

import (
	"context"
	"fmt"
	"os"
	"smart_gateway/gateway"
	gp "smart_gateway/gateway/pipeline"
	"time"
)

// pipeline_demo 演示：从“数据库”拉取订单，经 gateway 真实设备执行四步流水线
func main() {
	// 加载设备配置（可由环境变量覆盖）
	cfg := gateway.Config{
		CoffeeHost:  env("COFFEE_HOST", "localhost"),
		CoffeePort:  envInt("COFFEE_PORT", 8888),
		GrinderHost: env("GRINDER_HOST", "localhost"),
		GrinderPort: envInt("GRINDER_PORT", 5021),
		IceHost:     env("ICE_HOST", "localhost"),
		IceRack:     envInt("ICE_RACK", 0),
		IceSlot:     envInt("ICE_SLOT", 2),
		MqttHost:    env("MQTT_HOST", "localhost"),
		MqttPort:    envInt("MQTT_PORT", 1883),
	}
	g := gateway.New(cfg)
	// 注入演示订单（生产环境替换为真实 DB）
	db := gp.NewInMemoryDB()
	db.Seed([]gp.Order{
		{ID: 101, CoffeeType: "LATTE", BoolIce: true, TableNum: 7},
		{ID: 102, CoffeeType: "AMERICANO", BoolIce: false, TableNum: 3},
		{ID: 103, CoffeeType: "ESPRESSO", BoolIce: false, TableNum: 5},
		{ID: 104, CoffeeType: "MOCHA", BoolIce: true, TableNum: 8},
	})
	// 初始化队列与组件
	q := gp.NewInMemoryQueue(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	iceMin := envInt("ICE_MIN_STOCK", 200)
	iceDispense := envInt("ICE_DISPENSE_AMOUNT", 100)
	poller := &gp.OrderPoller{DB: db, Q: q, Interval: 500 * time.Millisecond, Batch: 50}
	core := gp.NewCore(q, g.Grinder(), g.Coffee(), g.Ice(), g.Robot(), iceMin, iceDispense)
	syncer := &gp.ResultSyncer{DB: db, Q: q}
	// 启动三明治架构：搬运工→核心→记录员
	poller.Start(ctx)
	core.Start(ctx)
	syncer.Start(ctx)
	// 等待处理完成并输出结果
	time.Sleep(40 * time.Second)
	for _, o := range db.All() {
		if o.Status == "error" {
			fmt.Println(o.ID, o.CoffeeType, o.BoolIce, o.TableNum, o.Status, o.ErrorMsg)
		} else {
			fmt.Println(o.ID, o.CoffeeType, o.BoolIce, o.TableNum, o.Status)
		}
	}
}

// env/envInt 读取环境变量（缺省返回默认值）
func env(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
func envInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	var x int
	_, err := fmt.Sscanf(v, "%d", &x)
	if err != nil {
		return d
	}
	return x
}

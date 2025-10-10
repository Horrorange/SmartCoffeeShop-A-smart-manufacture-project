// cmd/gateway/main.go

package main

import (
	"fmt"
	"log"
	"smartgateway/pkg/drivers" // 引用你的drivers包
	"time"
)

// --- 辅助函数 ---
// printIceMakerStatus 用于清晰地打印制冰机的当前状态。
func printIceMakerStatus(d drivers.Device) {
	status, err := d.GetStatus()
	if err != nil {
		log.Printf("  [ERROR] Failed to get status: %v", err)
		return
	}
	iceAmount := 0
	if val, ok := status.Inventory["ICE"]; ok {
		iceAmount = val
	}

	fmt.Println("  --------------------------")
	fmt.Printf("  | Is Idle      | %-7v |\n", status.IsIdle)
	fmt.Printf("  | Ice Stock (g)| %-7d |\n", iceAmount)
	fmt.Printf("  | Error Code   | %-7d |\n", status.ErrorCode)
	fmt.Println("  --------------------------")
}

func main() {
	log.Println("--- S7 PLC (Ice Maker) Driver Verification Program ---")

	// --- 1. 初始化驱动 ---
	// rack=0, slot=1 是根据你的Python模拟器配置
	iceMaker := drivers.NewIceMakerS7Driver("localhost", "102", 0, 1)

	// --- 2. 连接设备 ---
	log.Println("[ACTION] Connecting to the S7 PLC (ice maker)...")
	if err := iceMaker.Connect(); err != nil {
		log.Fatalf("Fatal: Could not connect to the S7 simulator. Is it running? Error: %v", err)
	}
	defer iceMaker.Disconnect()
	log.Println("[SUCCESS] Connected successfully!")

	// --- 3. 获取初始状态 ---
	log.Println("\n[ACTION] Getting initial status...")
	printIceMakerStatus(iceMaker)

	// --- 4. 执行制冰任务 (MAKE_ICE) ---
	log.Println("\n[ACTION] Executing task: MAKE_ICE...")
	makeIceTask := drivers.Task{Command: "MAKE_ICE"}
	if err := iceMaker.ExecuteTask(makeIceTask); err != nil {
		log.Printf("[ERROR] Failed to send MAKE_ICE task: %v", err)
	} else {
		log.Println("[SUCCESS] MAKE_ICE command sent. Polling for completion (will take ~10s)...")
		time.Sleep(1 * time.Second)
		// 轮询等待任务完成。驱动本身是非阻塞的，所以我们需要在这里等待。
		for i := 0; i < 15; i++ { // 最多等待15秒
			status, _ := iceMaker.GetStatus()
			if status.IsIdle {
				log.Println("[INFO] Ice making process complete!")
				break
			}
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
		fmt.Println() // 换行
	}
	log.Println("[INFO] Checking status after making ice...")
	printIceMakerStatus(iceMaker)

	// --- 5. 执行取冰任务 (DISPENSE_ICE) - 成功案例 ---
	log.Println("\n[ACTION] Executing task: DISPENSE_ICE (150g)...")
	dispenseTask := drivers.Task{
		Command: "DISPENSE_ICE",
		Params:  map[string]interface{}{"amount_grams": 150},
	}
	if err := iceMaker.ExecuteTask(dispenseTask); err != nil {
		log.Printf("[ERROR] Failed to dispense 150g of ice: %v", err)
	} else {
		log.Println("[SUCCESS] DISPENSE_ICE command sent. Waiting for completion (will take ~2s)...")
		time.Sleep(3 * time.Second) // 等待时间比模拟器的处理时间稍长
	}
	log.Println("[INFO] Checking status after dispensing 150g of ice...")
	printIceMakerStatus(iceMaker)

	// --- 6. 执行取冰任务 - 超量案例 ---
	log.Println("\n[ACTION] Executing task: DISPENSE_ICE (5000g)...")
	dispenseOverTask := drivers.Task{
		Command: "DISPENSE_ICE",
		Params:  map[string]interface{}{"amount_grams": 5000},
	}
	if err := iceMaker.ExecuteTask(dispenseOverTask); err != nil {
		log.Printf("[ERROR] Failed to dispense 5000g of ice: %v", err)
	} else {
		log.Println("[SUCCESS] DISPENSE_ICE (overdraft) command sent. Waiting for completion...")
		time.Sleep(3 * time.Second)
	}
	log.Println("[INFO] Checking status after overdrafting ice (expecting 0g)...")
	printIceMakerStatus(iceMaker)

	// --- 7. 尝试一个无效指令 ---
	log.Println("\n[ACTION] Executing an invalid task...")
	invalidTask := drivers.Task{Command: "DO_NOTHING"}
	if err := iceMaker.ExecuteTask(invalidTask); err != nil {
		log.Printf("[SUCCESS] Task failed as expected! Reason: %v", err)
	} else {
		log.Println("[ERROR] This task should have failed, but it succeeded.")
	}

	log.Println("\n--- Verification Program Finished ---")
}

package main

import (
    "bytes"
    "fmt"
    "net/http"
    "io"
)

// demo 按顺序调用统一端点，演示四台设备的核心操作
func main() {
    post(`{"device":"grinder","action":"grind"}`)
    post(`{"device":"coffeemachine","action":"make","coffee_type":"LATTE"}`)
    post(`{"device":"ice_maker","action":"produce","ice_amount":500}`)
    post(`{"device":"ice_maker","action":"dispense","ice_amount":200}`)
    post(`{"device":"delivery_robots","action":"deliver","coffee_type":"LATTE","need_ice":true,"table_number":8}`)
}

func post(s string) {
    resp, err := http.Post("http://localhost:9090/cmd", "application/json", bytes.NewBufferString(s))
    if err != nil {
        fmt.Println("err", err)
        return
    }
    defer resp.Body.Close()
    b, _ := io.ReadAll(resp.Body)
    fmt.Println(resp.StatusCode, string(b))
}
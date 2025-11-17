package gateway

import (
    "encoding/json"
    "fmt"
    "time"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

// DeliveryRobot 通过 MQTT 与送餐机器人交互
// 发布主题: test/delivery_robot/command ；确认主题: test/delivery_robot/status
type DeliveryRobot struct {
    Host string
    Port int
}

func (d *DeliveryRobot) broker() string {
    return fmt.Sprintf("tcp://%s:%d", d.Host, d.Port)
}

func (d *DeliveryRobot) Deliver(coffeeType string, needIce bool, table int) error {
    ack := make(chan struct{}, 1)
    opts := mqtt.NewClientOptions().AddBroker(d.broker())
    opts.SetAutoReconnect(true)
    opts.SetConnectionLostHandler(func(mqtt.Client, error) {})
    opts.SetOnConnectHandler(func(c mqtt.Client) {
        st := c.Subscribe("test/delivery_robot/status", 0, func(_ mqtt.Client, _ mqtt.Message) {
            select { case ack <- struct{}{}: default: }
        })
        st.Wait()
    })
    c := mqtt.NewClient(opts)
    ct := c.Connect()
    ct.Wait()
    if ct.Error() != nil {
        return ct.Error()
    }
    defer c.Disconnect(250)
    payload := map[string]any{
        "order_id": 1,
        "coffee_type": coffeeType,
        "need_ice": needIce,
        "table_number": table,
    }
    b, _ := json.Marshal(payload)
    pt := c.Publish("test/delivery_robot/command", 0, false, b)
    pt.Wait()
    if err := pt.Error(); err != nil {
        return err
    }
    select {
    case <-ack:
        return nil
    case <-time.After(3 * time.Second):
        return nil
    }
}
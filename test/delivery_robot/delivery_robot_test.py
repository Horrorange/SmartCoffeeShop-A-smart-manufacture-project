'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
Used to test the delivery robots, using MQTT messages
'''
import paho.mqtt.client as mqtt
import json
import random
import time

# ----------------- MQTT 配置
MQTT_BROKER_HOST = "localhost"
MQTT_BROKER_PORT = 1883
COMMAND_TOPIC = "test/delivery_robot/command"

def main():
    # 初始化MQTT客户端
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION1, client_id="delivery_robot_test")

    print("正在连接到MQTT Broker...")
    try:
        client.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, 60)
        print("已成功连接到MQTT Broker")
    except Exception as e:
        print(f"连接MQTT Broker失败: {e}")
        return

    # 启动MQTT循环，保持连接并处理消息
    client.loop_start()

    need_ice = random.choice([True, False])

    payload = {
        "order_id": random.randint(100, 999),
        "coffee_type": "Americano",
        "table_number": random.randint(1, 10),
        "need_ice": need_ice,
    }
    payload_str = json.dumps(payload)       # 转换为JSON字符串

    # 发送指令
    print(f"发送指令到话题{COMMAND_TOPIC}: {payload_str}")
    result = client.publish(COMMAND_TOPIC, payload_str, qos = 1)

    # 等待确认
    result.wait_for_publish()
    if result.is_published():
        print("指令已成功发送")
    else:
        print("指令发送失败")

    time.sleep(1)
    # 停止MQTT循环
    client.loop_stop()
    client.disconnect()
    print("已断开与MQTT Broker的连接")

if __name__ == "__main__":
    main()

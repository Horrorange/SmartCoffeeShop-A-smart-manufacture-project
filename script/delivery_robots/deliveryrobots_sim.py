'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
Used to simulate the delivery robots, using MQTT messages
'''

import paho.mqtt.client as mqtt
import time
import json
import logging
from colorlog import ColoredFormatter
import random

logger = logging.getLogger("deliveryrobots_sim")
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
# 设置格式
formatter = ColoredFormatter(
    "%(log_color)s%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    log_colors={
        'DEBUG':    'cyan',
        'INFO':     'green',
        'WARNING':  'yellow',
        'ERROR':    'red',
        'CRITICAL': 'red,bg_white',
    },
)
handler.setFormatter(formatter)

if not logger.handlers:
    logger.addHandler(handler)

MQTT_BROKER_HOST = "0.0.0.0"
MQTT_BROKER_PORT = 1883
COMMAND_TOPIC = "test/delivery_robot/command"
STATUS_TOPIC = "test/delivery_robot/status"

def simulate_delivery(order_details: dict):
    """
    模拟配送，根据订单详情更新配送状态
    输入：order_details (dict) - 订单详情，包含订单ID、咖啡类型、是否需要冰、取餐点
    输入举例：{"order_id": 001, "coffee_type": "MATCHA LATTE", "need_ice": True, "table_number": 1}
    """
    order_id = order_details.get("order_id", "N/A")
    coffee_type = order_details.get("coffee_type", "Unknown Coffee")
    needs_ice = order_details.get("need_ice", False)
    destination_table = order_details.get("table_number", "Bar")

    ice_status_str = "(加冰)" if needs_ice else "(不加冰)"
    display_coffee_name = f"{coffee_type} {ice_status_str}"

    logger.info(f"收到任务：订单 [{order_id}] 配送 {display_coffee_name} 到 {destination_table}")

    logger.info(f"- 订单 [{order_id}] - 正在前往取餐点 ...")
    time.sleep(random.randint(2,4))
    logger.info(f"- 订单 [{order_id}] - 已取到 {display_coffee_name}")
    logger.info(f"- 订单 [{order_id}] - 正在前往 {destination_table} 号桌 ...")
    time.sleep(random.randint(3,5))
    logger.info(f"- 订单 [{order_id}] - 已送达 {display_coffee_name} 到 {destination_table} 号桌")
    logger.info(f"- 订单 [{order_id}] - 配送完成, 正在返回...")
    time.sleep(random.randint(2,5))
    logger.info(f"- 订单 [{order_id}] - 已返回, 进入待命状态")
    return display_coffee_name

def on_connect(client, userdata, flags, rc):
    if rc == 0:
        logger.info("已成功连接到MQTT代理")
        client.subscribe(COMMAND_TOPIC, qos=1)
    else:
        logger.error(f"连接失败, 错误码: {rc}")

def on_message(client, userdata, msg):
    try:
        payload_str = msg.payload.decode('utf-8')
        logger.info(f"从话题 {msg.topic} 收到消息: {payload_str}")

        order_details = json.loads(payload_str)
        coffee_name = simulate_delivery(order_details)
        if coffee_name:
            result = "DELIVERY_COMPLETE"

        status_payload = json.dumps({
            "order_id": order_details.get("order_id", "N/A"),
            "status": result,
            "coffee_name": coffee_name,
        })
        client.publish(STATUS_TOPIC, status_payload)
    
    except json.JSONDecodeError:
        logger.error(f"从话题 {msg.topic} 收到的消息不是有效的JSON格式: {payload_str}")
    except Exception as e:
        logger.error(f"处理消息时发生错误: {e}")

def main():
    client = mqtt.Client(client_id="delivery_robot_sim")
    client.on_connect = on_connect
    client.on_message = on_message

    logger.info("正在连接到MQTT Broker...")
    try:
        client.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, 60)
        logger.info("已成功连接到MQTT Broker")
    except Exception as e:
        logger.error(f"连接MQTT Broker失败: {e}")
        return
        
    client.loop_forever()

if __name__ == "__main__":
    main()
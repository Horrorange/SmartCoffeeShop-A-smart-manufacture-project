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

# ---------------- 设置logger格式
logger = logging.getLogger("deliveryrobots_sim")
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
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

# ---------------- 配置服务器
MQTT_BROKER_HOST = "0.0.0.0"                    # 监听所有网络接口
MQTT_BROKER_PORT = 1883                         # MQTT默认端口
COMMAND_TOPIC = "test/delivery_robot/command"   # 命令话题，用于接收订单指令
STATUS_TOPIC = "test/delivery_robot/status"     # 状态话题，用于发送配送状态


# -----------------------------------------------------
# MQTT 通信逻辑
# 命令话题输入 ： {"order_id": <订单号>, "coffee_type": <咖啡类型>, "need_ice": <是否需要冰>, "table_number": <桌号>}
# 状态话题输出 ： {"order_id": <订单号>, "status": <配送状态>, "table_number": <桌号>}
# 输入和输出都使用 json格式
# -----------------------------------------------------

# ---------------- 模拟配送
def simulate_delivery(order_details: dict):
    """
    模拟配送，根据订单详情更新配送状态
    输入：order_details (dict) - 订单详情，包含订单ID、咖啡类型、是否需要冰、取餐点
    输入举例：{"order_id": 001, "coffee_type": "MATCHA LATTE", "need_ice": True, "table_number": 1}
    """
    # 解析输入
    order_id = order_details.get("order_id", "N/A")
    coffee_type = order_details.get("coffee_type", "Unknown Coffee")
    needs_ice = order_details.get("need_ice", False)
    destination_table = order_details.get("table_number", "Bar")

    # 冰块逻辑
    ice_status_str = "(加冰)" if needs_ice else "(不加冰)"
    display_coffee_name = f"{coffee_type} {ice_status_str}"

    # 开始模拟配送过程
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
    '''
    连接到MQTT代理时的回调函数
    当客户端成功连接到MQTT代理时调用
    rc=0,代表连接成功
    rc=1,代表连接失败, 服务器拒绝连接
    '''
    if rc == 0:
        logger.info("已成功连接到MQTT代理")
        client.subscribe(COMMAND_TOPIC, qos=1)
    else:
        logger.error(f"连接失败, 错误码: {rc}")

def on_message(client, userdata, msg):
    '''
    当客户端收到MQTT消息时的回调函数
    当客户端订阅的话题收到消息时调用
    '''
    try:
        # 从消息中解码收到的信息
        payload_str = msg.payload.decode('utf-8')
        logger.info(f"从话题 {msg.topic} 收到消息: {payload_str}")

        # 解析订单详情，开始配送流程
        order_details = json.loads(payload_str)
        coffee_name = simulate_delivery(order_details)
        result = "DELIVERY_COMPLETE"

        # 准备返回的数据
        status_payload = json.dumps({
            "order_id": order_details.get("order_id", "N/A"),
            "status": result,
            "coffee_name": coffee_name,
        })

        # 发送配送状态到状态话题
        client.publish(STATUS_TOPIC, status_payload)
        logger.info(f"已发送配送状态到话题 {STATUS_TOPIC}: {status_payload}")
    
    except json.JSONDecodeError:
        logger.error(f"从话题 {msg.topic} 收到的消息不是有效的JSON格式: {payload_str}")
    except Exception as e:
        logger.error(f"处理消息时发生错误: {e}")

def main():
    # 初始化MQTT客户端
    client = mqtt.Client(client_id="delivery_robot_sim")
    client.on_connect = on_connect  # 连接成功回调
    client.on_message = on_message  # 收到消息回调

    logger.info("正在连接到MQTT Broker...")
    # 连接到MQTT代理
    try:
        client.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, 60)
        logger.info("已成功连接到MQTT Broker")
    except Exception as e:
        logger.error(f"连接MQTT Broker失败: {e}")
        return
    
    # 保持连接并处理消息
    client.loop_forever()

if __name__ == "__main__":
    main()
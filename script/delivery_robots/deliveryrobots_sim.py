'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
Used to simulate the delivery robots, using MQTT messages
'''
import paho.mqtt.client as mqtt
import time
import json
import logging
import os
from colorlog import ColoredFormatter
import random

# ---------------- 设置logger格式
logger = logging.getLogger("delivery_robot")
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

# ---------------- 配置服务器
# 在容器网络中应使用服务名访问Broker，支持环境变量覆盖
# 本地运行默认使用 localhost，避免未设置环境变量时的连接失败
MQTT_BROKER_HOST = os.getenv("MQTT_HOST", "localhost")
MQTT_BROKER_PORT = int(os.getenv("MQTT_PORT", "1883"))
COMMAND_TOPIC = "test/delivery_robot/command"   # 命令话题，用于接收订单指令
STATUS_TOPIC = "test/delivery_robot/status"     # 状态话题，用于发送配送状态


# -----------------------------------------------------
# MQTT 通信逻辑
# 命令话题输入 ： {"order_id": <订单号>, "coffee_type": <咖啡类型>, "need_ice": <是否需要冰>, "table_number": <桌号>}
# 状态话题输出 ： {"order_id": <订单号>, "status": <配送状态>, "table_number": <桌号>}
# 输入和输出都使用 json格式
# -----------------------------------------------------

# ---------------- 模拟配送
def simulate_delivery(table_number: int):
    """
    模拟配送，根据订单详情更新配送状态
    输入：table_number (int) - 桌号
    输入举例：1
    """
    # 解析输入
    destination_table = f"Table{table_number}"


    # 开始模拟配送过程
    logging.info(f"收到任务：配送到 {destination_table}")
    logging.info(f"- 正在前往取餐点 ...")
    time.sleep(random.randint(2,4))
    logging.info(f"-已取到 咖啡")
    logging.info(f"-正在前往 {destination_table} 号桌 ...")    
    time.sleep(random.randint(3,5))
    logging.info(f"-已送达 咖啡 到 {destination_table} 号桌")
    logging.info(f"-配送完成, 正在返回...")
    time.sleep(random.randint(2,5))
    logging.info(f"已返回, 进入待命状态")
    return "Done"

def on_connect(client, userdata, flags, rc, properties=None):
    '''
    连接到MQTT代理时的回调函数
    当客户端成功连接到MQTT代理时调用
    rc=0,代表连接成功
    rc=1,代表连接失败, 服务器拒绝连接
    '''
    if rc == 0:
        logging.info("已成功连接到MQTT代理")
        client.subscribe(COMMAND_TOPIC, qos=1)
    else:
        logging.error(f"连接失败, 错误码: {rc}")

def on_message(client, userdata, msg):
    '''
    当客户端收到MQTT消息时的回调函数
    当客户端订阅的话题收到消息时调用
    '''
    try:
        # 从消息中解码收到的信息
        payload_str = msg.payload.decode('utf-8')
        logging.info(f"从话题 {msg.topic} 收到消息: {payload_str}")

        # 解析订单详情
        order_details = json.loads(payload_str)

        # 立即发送接收确认，便于上游快速得到ACK
        ack_payload = json.dumps({
            "order_id": order_details.get("order_id", "N/A"),
            "status": "RECEIVED",
            "table_number": order_details.get("table_number", "N/A"),
        })
        client.publish(STATUS_TOPIC, ack_payload, qos=1)
        logging.info(f"已发送接收确认到话题 {STATUS_TOPIC}: {ack_payload}")

        # 开始模拟配送流程并在完成后发送最终状态
        status = simulate_delivery(order_details.get("table_number", 0))
        final_result = "DELIVERY_COMPLETE" if status == "Done" else "DELIVERY_FAILED"
        final_payload = json.dumps({
            "order_id": order_details.get("order_id", "N/A"),
            "status": final_result,
            "table_number": order_details.get("table_number", "N/A"),
        })
        client.publish(STATUS_TOPIC, final_payload, qos=1)
        logging.info(f"已发送配送完成到话题 {STATUS_TOPIC}: {final_payload}")
    
    except json.JSONDecodeError:
        logging.error(f"从话题 {msg.topic} 收到的消息不是有效的JSON格式: {payload_str}")
    except Exception as e:
        logging.error(f"处理消息时发生错误: {e}")

def main():
    # 初始化MQTT客户端
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id="delivery_robot_sim")
    client.on_connect = on_connect  # 连接成功回调
    client.on_message = on_message  # 收到消息回调

    logging.info("正在连接到MQTT Broker...")
    # 连接到MQTT代理，增加重试以等待Broker就绪
    max_attempts = 5
    for attempt in range(1, max_attempts + 1):
        try:
            client.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, 60)
            logging.info("已成功连接到MQTT Broker")
            break
        except Exception as e:
            if attempt == max_attempts:
                logging.error(f"连接MQTT Broker失败: {e}")
                return
            logging.warning(f"连接失败，第{attempt}次重试，原因: {e}")
            time.sleep(2)
    
    # 保持连接并处理消息
    client.loop_forever()

if __name__ == "__main__":
    main()
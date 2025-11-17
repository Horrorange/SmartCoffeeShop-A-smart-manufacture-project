'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
Used to simulate the coffee machine, using custom TCP messages
'''


import socket
import logging
import time
import random
from colorlog import ColoredFormatter
import threading


logger = logging.getLogger("coffeemachine_sim")
logger.setLevel(logging.DEBUG)
# 用于输出，后续可以扩展至网络传输
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

# -------------------- 1. 端口设置
HOST = '0.0.0.0'
PORT = 8888

# -------------------- 2. 库存与食谱设置
MAX_STORAGE = 20 # 最大库存为50
inventory = {
    "MILK": MAX_STORAGE,
    "OAT_MILK": MAX_STORAGE,
    "MATCHA_SAUCE": MAX_STORAGE,
    "CHOCOLATE_SAUCE": MAX_STORAGE,
    "CARAMEL_SYRUP": MAX_STORAGE,
}


# 食谱记录所有种类咖啡所需要消耗的原材料
recipes = {
    "LATTE": {"MILK": 3},
    "FLAT WHITE": {"MILK": 3},
    "CAPPUCCINO": {"MILK": 3},
    "MACCHIATO": {"MILK": 2, "CARAMEL_SYRUP": 1},
    "OAT LATTE": {"OAT_MILK": 3},
    "MOCHA": {"MILK": 2, "CHOCOLATE_SAUCE": 1},
    "MATCHA LATTE": {"MILK": 2, "MATCHA_SAUCE": 1},
    "ESPRESSO":{},
    "AMERICANO": {},
    "LONG BLACK": {},
}
VALID_COFFEES = list(recipes.keys())


def check_and_custom_ingredients(coffee_type):
    """
    检查咖啡的原料是否充足，如果充足则消耗，不充足则返回False和缺少的原材料列表
    """
    recipe = recipes.get(coffee_type)

    required_ingredients = [] # 缺少的原材料列表
    # 首先检查所有原料是否充足
    for ingredient, amount in recipe.items():
        if inventory[ingredient] < amount:
            logger.error(f"原料 {ingredient} 不足，需要 {amount} 单位，当前库存 {inventory[ingredient]} 单位")
            required_ingredients.append(ingredient)
        
        # 如果有缺少的原料，返回False和缺少的原料列表
        if required_ingredients:
            return required_ingredients

        # 如果所有原料都充足，则消耗原料
        for ingredient, amount in recipe.items():
            inventory[ingredient] -= amount

        return []



# ------------------------------
# 自定义报文操作逻辑
# 编码格式： utf-8
# ｜ 指令类型            ｜ 成功返回值                     ｜ 失败返回值 
# ｜ MAKE:COFFEE_TYPE   | DONE:SUCCESS                  | ERROR:INSUFFICIENT_INGREDIENT  ERROR:UNKNOWN_COFFEE_TYPE
# ｜ REFILL:INGREDIENT  | ACK:REFILL_SUCCESS:INGREDIENT | ERROR:UNKNOWN_INGREDIENT
# ｜ REFILL:ALL         | ACK:REFILL_SUCCESS:ALL        | N/A
# ｜ STATUS:INGREDIENTS | ACK:STATUS:INVENTORY:MILK=50  | N/A
# ------------------------------
def handle_client(conn, addr):
    """
    处理客户端连接，接收客户端发送的咖啡类型，检查原料是否充足，充足则制作咖啡，不充足则返回错误信息
    """
    logger.info(f"接收到来自 {addr} 的连接请求")
    with conn:  # 确保连接在处理完成后关闭
        while True: # 持续监听客户端请求
            try:
                data = conn.recv(1024)  # 接收客户端发送的消息
                if not data:
                    logger.warning(f"客户端 {addr} 主动断开连接") # 客户端主动断开连接
                    break

                message = data.decode().strip().upper() # 解码并转换为大写
                logger.debug(f"客户端 {addr} 发送指令: {message}") # 记录接收到的指令
                
                # ----------------- 协议解析
                parts = message.split(":", 1)
                command = parts[0]
                payload = parts[1] if len(parts) > 1 else ""




                if command == "MAKE":
                    coffee_type = payload
                    if coffee_type not in VALID_COFFEES:
                        conn.sendall(b"ERROR:UNKNOWN_COFFEE_TYPE\n")
                        logger.error(f"未知咖啡类型: {coffee_type}")
                        continue
                    else:
                        missing_ingredients = check_and_custom_ingredients(coffee_type)
                        if not missing_ingredients:
                            conn.sendall(b"ACK:MAKE\n")
                            logger.info(f"开始制作 {coffee_type}")

                            # 模拟制作时间
                            time.sleep(random.randint(5,10))

                            conn.sendall(b"DONE:SUCCESS\n")
                            logger.info(f"成功制作 {coffee_type}")
                        else:
                            error_message = f"ERROR:INSUFFICIENT_INGREDIENT:{', '.join(missing_ingredients)}"
                            conn.sendall(error_message.encode('utf-8') + b"\n")
                            logger.error(f"制作 {coffee_type} 失败，缺少原料: {', '.join(missing_ingredients)}")





                elif command == "REFILL":
                    ingredient_to_refill = payload

                    if ingredient_to_refill == "ALL":
                        for ingredient in inventory:
                            inventory[ingredient] = MAX_STORAGE
                        time.sleep(7)
                        logger.info("所有原料已被补充")
                        conn.sendall(b"ACK:REFILL_SUCCESS:ALL\n")
                    
                    elif ingredient_to_refill in inventory:
                        inventory[ingredient_to_refill] = MAX_STORAGE
                        time.sleep(3)
                        logger.info(f"{ingredient_to_refill} 已被补充")
                        conn.sendall(b"ACK:REFILL_SUCCESS:" + ingredient_to_refill.encode('utf-8') + b"\n")
                    else:
                        conn.sendall(b"ERROR:INVALID_INGREDIENT\n")
                        logger.error(f"未知原料: {ingredient_to_refill}")



                elif command == "STATUS" and payload == "INGREDIENTS":
                    status_string = ",".join([f"{ingredient}={amount}" for ingredient, amount in inventory.items()])
                    resp = f"STATUS:INGREDIENTS:{status_string}\n"
                    conn.sendall(resp.encode('utf-8'))
                    logger.info(f"Sent inventory status: {status_string}")
                
                else:
                    conn.sendall(b"ERROR:UNKNOWN_COMMAND\n")
                    logger.error(f"未知指令格式: '{message}'")
            

            except ConnectionResetError:
                logger.error(f"客户端 {addr} 主动断开连接")
                break
            except Exception as e:
                logger.error(f"处理客户端 {addr} 时发生意外错误: {e}")
                break

def run_server():
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
        server_socket.bind((HOST, PORT))
        server_socket.listen()
        logger.info(f"咖啡机器服务器已启动，监听端口 {PORT}")

        while True:
            conn, addr = server_socket.accept()
            # 为每个客户端连接创建一个新线程，使其可以处理多个并发连接
            client_thread = threading.Thread(target=handle_client, args=(conn, addr))
            client_thread.start()

if __name__ == "__main__":
    run_server()


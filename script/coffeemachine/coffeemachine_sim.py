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

# -------------------- 0. 基本设置
# 为进程单独设置对应的logger
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
MAX_STORAGE = 50 # 最大库存为50
inventory = {
    "MILK": MAX_STORAGE,
    "OAT_MILK": MAX_STORAGE,
    "MATCHA_SAUCE": MAX_STORAGE,
    "CHOCOLATE_SAUCE": MAX_STORAGE,
    "CARAMEL_SYRUP": MAX_STORAGE,
}
# 线程锁， 防止多线程访问时出现问题
inventory_lock = threading.Lock()

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
    if not recipe:
        logger.error(f"Unknown coffee type: {coffee_type}")
        return False, None
    
    with inventory_lock:
        required_ingredients = []
        for ingredient, amount in recipe.items():
            if inventory[ingredient] < amount:
                logger.error(f"Not enough {ingredient} for {coffee_type}")
                required_ingredients.append(ingredient)
            if required_ingredients:
                return False, required_ingredients

        for ingredient, amount in recipe.items():
            inventory[ingredient] -= amount

        logger.info(f"Successfully made {coffee_type}, current inventory: {inventory}")
        return True, None

def handle_client(conn, addr):
    """
    处理客户端连接，接收客户端发送的咖啡类型，检查原料是否充足，充足则制作咖啡，不充足则返回错误信息
    """
    logger.info(f"Accepted connection from {addr}.")
    with conn:
        while True:
            try:

                data = conn.recv(1024)
                if not data:
                    logger.warning(f"Client {addr} disconnected.")
                    break

                message = data.decode().strip().upper()
                logger.debug(f"Received command: {message}")
                
                # ----------------- 协议解析
                parts = message.split(":", 1)
                command = parts[0]
                payload = parts[1] if len(parts) > 1 else ""

                if command == "MAKE":
                    coffee_type = payload
                    if coffee_type in VALID_COFFEES:
                        can_make, missing_ingredients = check_and_custom_ingredients(coffee_type)
                        if can_make:
                            conn.sendall(b"ACK:MAKE\n")
                            logger.info(f"Command acknowledged. Starting to make {coffee_type}.")

                            # 模拟制作时间
                            making_time = random.randint(5, 10)
                            time.sleep(making_time)

                            conn.sendall(b"DONE:SUCCESS\n")
                            logger.info(f"Sucessfully made {coffee_type} in {making_time}.")
                        else:
                            error_message = f"ERROR:INSUFFICIENT_INGREDIENT:{', '.join(missing_ingredients)}"
                            conn.sendall(error_message.encode('utf-8') + b"\n")
                            logger.error(f"Failed to make {coffee_type}. Missing ingredients: {', '.join(missing_ingredients)}")
                    else:
                        conn.sendall(b"ERROR:UNKNOWN_COFFEE_TYPE\n")
                        logger.error(f"Invalid coffee type requested: {coffee_type}")

                elif command == "REFILL":
                    ingredient_to_refill = payload
                    with inventory_lock:
                        if ingredient_to_refill == "ALL":
                            for ingredient in inventory:
                                inventory[ingredient] = MAX_STORAGE
                            logger.info("All ingredients have been refilled.")
                            conn.sendall(b"ACK:REFILL_SUCCESS:ALL\n")
                        elif ingredient_to_refill in inventory:
                            inventory[ingredient_to_refill] = MAX_STORAGE
                            logger.info(f"{ingredient_to_refill} has been refilled.")
                            conn.sendall(b"ACK:REFILL_SUCCESS:" + ingredient_to_refill.encode('utf-8') + b"\n")
                        else:
                            conn.sendall(b"ERROR:INVALID_INGREDIENT\n")
                            logger.error(f"Invalid ingredient requested for refill: {ingredient_to_refill}")
                
                elif command == "STATUS" and payload == "INGREDIENTS":
                        with inventory_lock:
                            status_string = ",".join([f"{ingredient}:{amount}" for ingredient, amount in inventory.items()])
                        resp = f"STATUS:INGREDIENTS:{status_string}\n"
                        conn.sendall(resp.encode('utf-8'))
                        logger.info(f"Sent inventory status: {status_string}")
                
                else:
                    conn.sendall(b"ERROR:UNKNOWN_COMMAND\n")
                    logger.error(f"Unknown command format: '{message}'")
            
            except ConnectionResetError:
                logger.error(f"Connection with {addr} was closed by the client.")
                break
            except Exception as e:
                logger.error(f"An unexpected error occurred with client {addr}: {e}")
                break

def run_server():
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
        server_socket.bind((HOST, PORT))
        server_socket.listen()
        logger.info(f"Coffee Machine server listening on {HOST}:{PORT}")

        while True:
            conn, addr = server_socket.accept()
            # 为每个客户端连接创建一个新线程，使其可以处理多个并发连接
            client_thread = threading.Thread(target=handle_client, args=(conn, addr))
            client_thread.start()

if __name__ == "__main__":
    run_server()


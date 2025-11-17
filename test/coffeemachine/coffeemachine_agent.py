'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
智能边缘网关的简单实现
- 连接并控制磨粉机和咖啡机，分别采用Modbus TCP 和 自定义TCP协议
- 实现本地决策，完成订单的制作
'''
# 套接字接口编程
import socket
from pyModbusTCP.client import ModbusClient
import time
import logging
from colorlog import ColoredFormatter

# -------------------- 0. 基本设置
# 为进程单独设置对应的logger
logger = logging.getLogger("coffeemachine_agent")
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

# -------------------- 1. 设备设置
COFFEE_MACHINE_HOST = "localhost" # 连接咖啡机的IP地址
COFFEE_MACHINE_PORT = 8888

# -------------------- 2. 决策菜谱
RECIPES = ["LATTE", "FLAT WHITE", "CAPPUCCINO", "MACCHIATO", "OAT LATTE", "MOCHA", "MATCHA LATTE", "ESPRESSO", "AMERICANO", "LONG BLACK"]

def make_coffee(coffeeType, coffee_maker: socket.socket):
    if coffeeType not in RECIPES:
        return ["FAIL"]
    else:
        coffee_maker.sendall(f"MAKE:{coffeeType}".encode('utf-8'))
        resp = coffee_maker.recv(1024).decode('utf-8').strip()
        if resp.startswith("ACK"):
            resp2 = coffee_maker.recv(1024).decode('utf-8').strip()
            if resp2.startswith("DONE"):
                return ["SUCCESS"]
            else:
                return ["FAIL"]
        elif resp.startswith("ERROR"):
            # 解析格式：ERROR:INSUFFICIENT_INGREDIENT:MILK, CARAMEL_SYRUP
            parts = resp.split(":", 2)
            if len(parts) >= 3 and parts[1] == "INSUFFICIENT_INGREDIENT":
                missing_ingredients_str = parts[2]
                missingIngredients = [ing.strip() for ing in missing_ingredients_str.split(",")]
                refill_ingredients(missingIngredients, coffee_maker)
                return make_coffee(coffeeType, coffee_maker)
            else:
                return ["FAIL"]

def refill_ingredients(ingredients: list, coffee_maker: socket.socket):
    for ingredient in ingredients:
        coffee_maker.sendall(f"REFILL:{ingredient}".encode('utf-8'))
        resp = coffee_maker.recv(1024).decode('utf-8').strip()
        if resp.startswith("ACK"):
            logger.info(f"成功补充原料 {ingredient}")
        else:
            logger.error(f"补充原料 {ingredient} 失败，返回值：{resp}")

def refill_all_ingredients(coffee_maker: socket.socket):
    coffee_maker.sendall("REFILL:ALL".encode('utf-8'))
    resp = coffee_maker.recv(1024).decode('utf-8').strip()
    if resp.startswith("ACK"):
        logger.info(f"成功补充所有原料")
    else:
        logger.error(f"补充所有原料失败，返回值：{resp}")

def check_ingredients(coffee_maker: socket.socket):
    coffee_maker.sendall("STATUS:INGREDIENTS".encode('utf-8'))
    resp = coffee_maker.recv(1024).decode('utf-8').strip()
    if resp.startswith("STATUS:INGREDIENTS"):
        # 解析格式：STATUS:INGREDIENTS:MILK=50,OAT_MILK=50,MATCHA_SAUCE=50,CHOCOLATE_SAUCE=50,CARAMEL_SYRUP=50
        ingredients_part = resp.split(":", 2)[2]  # 获取第三部分
        ingredients = ingredients_part.split(",")
        logger.info(f"当前原料库存：{ingredients}")
    else:
        logger.error(f"检查原料失败，返回值：{resp}")


def main():
    # 只创建与咖啡机的连接
    try:
        coffee_maker_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        coffee_maker_socket.connect((COFFEE_MACHINE_HOST, COFFEE_MACHINE_PORT))
    except Exception as e:
        logger.error(f"创建咖啡机连接失败，错误信息：{e}")
        return
    
    logger.info("成功连接到咖啡机")
    
    # 验证咖啡机功能
    try:
        import random
        
        # 检查原料库存
        logger.info("=== 检查原料库存 ===")
        check_ingredients(coffee_maker_socket)
        time.sleep(1)
        
        # 补充所有原料
        logger.info("=== 补充所有原料 ===")
        refill_all_ingredients(coffee_maker_socket)
        time.sleep(2)
        
        # 随机制作20杯咖啡
        logger.info("=== 开始随机制作20杯咖啡 ===")
        
        for i in range(200):
            # 每5杯检查一次库存
            if i % 5 == 0 and i > 0:
                logger.info(f"=== 第{i}杯后检查库存 ===")
                check_ingredients(coffee_maker_socket)
                time.sleep(1)
            

            
            # 随机选择咖啡类型
            coffee_type = random.choice(RECIPES)
            logger.info(f"--- 第{i+1}杯：制作 {coffee_type} ---")
            
            # 制作咖啡
            result = make_coffee(coffee_type, coffee_maker_socket)
            if result == ["SUCCESS"]:
                logger.info(f"✅ 第{i+1}杯 {coffee_type} 制作成功")
            else:
                logger.error(f"❌ 第{i+1}杯 {coffee_type} 制作失败")
            
            time.sleep(1)  # 每杯之间间隔1秒
        
        # 最终检查库存
        logger.info("=== 最终库存检查 ===")
        check_ingredients(coffee_maker_socket)
    
    finally:
        logger.info("所有任务完成，关闭连接")
        coffee_maker_socket.close()

if __name__ == "__main__":
    main()
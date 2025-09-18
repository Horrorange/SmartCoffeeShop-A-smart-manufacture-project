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
GRINDER_HOST = "localhost" # 连接磨粉机的IP地址
GRINDER_PORT = 502
COFFEE_MACHINE_HOST = "localhost" # 连接咖啡机的IP地址
COFFEE_MACHINE_PORT = 8888

CMD_REG = 0             # 指令寄存器
STATUS_REG = 1          # 状态寄存器 
BEAN_LEVEL_REG = 2      # 豆量寄存器
ERROR_CODE_REG = 3      # 错误代码寄存器

# -------------------- 2. 决策菜谱
RECIPES = {
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

def parse_inventory(status_string: str) -> dict:
    """
    解析咖啡机的状态字符串，返回当前库存的字典表示
    输入： "STATUS:MILK=50,OAT_MILK=50,..."
    输出： {"MILK": 50, "OAT_MILK": 50, ...}
    """
    inventory_dict = {}
    try:
        parts_str = status_string.split(":")[2]
        ingredient_parts = parts_str.split(",")
        for part in ingredient_parts:
            key, val = part.split("=")
            inventory_dict[key.strip()] = int(val.strip())
        return inventory_dict
    except Exception as e:
        logger.error(f"解析咖啡机状态字符串:'{status_string}'失败，错误信息：{e}")
        return {}


def refill_coffee_beans(grinder: ModbusClient) -> bool:
    """
    向磨粉机发送指令，补充咖啡豆
    """
    logger.info("向磨粉机发送指令，补充咖啡豆")
    try:
        grinder.write_single_register(CMD_REG, 2) # 发送指令2，补充咖啡豆
        return True
    except Exception as e:
        logger.error(f"向磨粉机发送补充咖啡豆指令失败，错误信息：{e}")
        return False

def handle_order(coffee_type: str, grinder: ModbusClient, coffee_maker: socket.socket):
    """
    处理订单，根据菜谱检查原料是否充足，充足则发送指令给磨粉机和咖啡机，不充足则返回错误信息
    """
    logger.info(f"--- 处理订单：{coffee_type} ---")

    # ------- 检查配方是否存在
    recipe = RECIPES.get(coffee_type)
    if recipe is None:
        logger.error(f"未知的咖啡类型：{coffee_type}")
        return
    
    # ------- 向咖啡机发送原料查询请求
    logger.info("查询咖啡机当前原料库存")
    try:
        coffee_maker.sendall(b"STATUS:INGREDIENTS")
        resp = coffee_maker.recv(1024).decode('utf-8').strip()
        if not resp.startswith("STATUS:"):
            logger.error("无法获取咖啡机原料库存")
            return
    except socket.timeout:
        logger.error("获取咖啡机原料库存超时，检查咖啡机是否在线")
        return

    # ------- 解析库存状态
    inventory = parse_inventory(resp)
    logger.info(f"当前咖啡机原料库存：{inventory}")

    # ------- 在网关本地检查原料是否充足
    logger.info(f"检查配方 {coffee_type} 是否符合当前库存 {inventory}")
    can_make = True
    missing_ingredient = []

    if not recipe:
        logger.info(f"配方 {coffee_type} 无需任何原料")
    else:
        for ingredient, amount in recipe.items():
            if inventory.get(ingredient, 0) < amount:
                can_make = False
                missing_ingredient.append(ingredient)
    
    # -------- 检查咖啡豆是否充足
    bean_level = grinder.read_holding_registers(BEAN_LEVEL_REG, 1)[0]
    error_code = grinder.read_holding_registers(ERROR_CODE_REG, 1)[0]
    if bean_level < 10 or error_code == 1:
        logger.warning(f"磨豆机咖啡豆不足 (当前豆量: {bean_level}%)，开始自动补充...")
        try:
            refill_coffee_beans(grinder)
        except Exception as e:
            logger.error(f"自动补充咖啡豆失败，错误信息：{e}")
            return
    else:
        logger.info(f"当前咖啡豆数量：{bean_level}%")
    
    # -------- 检查咖啡豆是否补充
    while bean_level < 10 or error_code == 1:
        bean_level = grinder.read_holding_registers(BEAN_LEVEL_REG, 1)[0]
        error_code = grinder.read_holding_registers(ERROR_CODE_REG, 1)[0]
        if bean_level >= 10 and error_code == 0:
            logger.info(f"补充完成，当前咖啡豆数量：{bean_level}%")
            break
        time.sleep(1)  # 等待1秒后再次检查

    # -------- 原料充足，开始制作逻辑
    if can_make:
        logger.info("原料充足，开始进行制作")
        logger.info(f"  -> 向磨粉机发送指令...")
        grinder.write_single_register(CMD_REG, 1) # 发送指令1，开始研磨

        logger.info(f"  -> 等待磨粉机开始工作...")
        # 首先等待磨粉机状态变为"磨粉中"(1)，确保磨粉机收到并开始处理指令
        while True:
            status = grinder.read_holding_registers(STATUS_REG, 1)[0]
            if status == 1:  # 磨粉中
                logger.info("  -> 磨粉机开始磨粉...")
                break
            time.sleep(0.1)  # 短间隔检查，快速响应

        logger.info(f"  -> 等待磨粉机磨粉完成...")
        # 然后等待磨粉机状态变回"空闲"(0)，表示磨粉完成
        while True:
            status = grinder.read_holding_registers(STATUS_REG, 1)[0]
            if status == 0:  # 空闲，磨粉完成
                logger.info("  -> 磨粉完成")
                break
            time.sleep(0.5)  # 适中间隔检查

        logger.info(f"  -> 向咖啡机发送指令...")
        make_command = f"MAKE:{coffee_type}\n".encode('utf-8')
        coffee_maker.sendall(make_command)

        logger.info("  -> 等待咖啡机制作完成...")
        resp = coffee_maker.recv(1024).decode('utf-8').strip()
        logger.info(f"  -> 收到咖啡机确认：{resp}")

        done_response = coffee_maker.recv(1024).decode('utf-8').strip()
        logger.info(f"  -> 收到咖啡机状态：{done_response}")

        if "SUCCESS" in done_response:
            logger.info(f"咖啡 {coffee_type} 制作成功")
        else:
            logger.error(f"咖啡 {coffee_type} 制作失败，错误信息：{done_response}")
    # -------- 原料不足，自动补货逻辑
    else:
        logger.warning(f"咖啡 {coffee_type} 制作失败，缺失原料：{', '.join(missing_ingredient)}")

        logger.info("  -> 进行自动补货...")
        reill_command = f"REFILL:{missing_ingredient}\n".encode("utf-8")
        coffee_maker.sendall(reill_command)

        refill_ack = coffee_maker.recv(1024).decode('utf-8').strip()
        if "SUCCESS" in refill_ack:
            logger.info(f"咖啡 {coffee_type} 自动补货成功，缺失原料：{', '.join(missing_ingredient)}")
            logger.info("请重新下单")
        else:
            logger.error(f"咖啡 {coffee_type} 自动补货失败，错误信息：{refill_ack}")

def main():
    # 创建与磨粉机的连接
    grinder_client = ModbusClient(host=GRINDER_HOST, port=GRINDER_PORT)
    # 创建与咖啡机的连接
    try:
        coffee_maker_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        coffee_maker_socket.connect((COFFEE_MACHINE_HOST, COFFEE_MACHINE_PORT))
    except Exception as e:
        logger.error(f"创建咖啡机连接失败，错误信息：{e}")
        return
    
    logger.info("成功连接到咖啡机与磨粉机")
    # 制作逻辑
    try:
        handle_order("LATTE", grinder_client, coffee_maker_socket)
        time.sleep(2)

        handle_order("ESPRESSO", grinder_client, coffee_maker_socket)
        time.sleep(2)

        handle_order("OAT LATTE", grinder_client, coffee_maker_socket)
    
    finally:
        logger.info("所有任务完成，关闭连接")
        grinder_client.close()
        coffee_maker_socket.close()

if __name__ == "__main__":
    main()
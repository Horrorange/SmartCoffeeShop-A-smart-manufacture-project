'''
Author : Orange horrorange@qq.com
Last-modified: 2025-09-17
ModBus Client test
'''

import time
import argparse
from pyModbusTCP.client import ModbusClient
import logging

# 设置logging
logging.basicConfig(
	level = logging.DEBUG,
	format = "%(asctime)s %(levelname)s %(message)s",
	datefmt = "%Y-%m-%d %H:%M:%S",
)

# 串口号保持一致
CMD_REG = 0         # command register
STATUS_REG = 1      # status register
BEAN_LEVEL_REG = 2  # bean level register

# 设定启动参数
parser = argparse.ArgumentParser(description="ModBus Client Test")
parser.add_argument("--host", type=str, default="localhost", help="ModBus server host IP")
args = parser.parse_args()

# 读取IP地址
SERVER_HOST = args.host
SERVER_PORT = 502

# 连接Modbus服务
client = ModbusClient(host=SERVER_HOST, port=SERVER_PORT, auto_open=False)
connection_result = client.open()

logging.debug(f"Modbus Client connected to {SERVER_HOST}:{SERVER_PORT}")

# ---------------- 1. 读取当前状态
logging.debug("--- Reading Input State ---")
status_map = {0:"Idle", 1:"Grinding", 2:"Error"}
# 从 STATUS_REG 开始读取 2 个寄存器
regs = client.read_holding_registers(STATUS_REG, 2)

if regs:
    status, bean_level  = regs
    logging.debug(f"Grinder status: {status_map.get(status, 'Unknown')}")
    logging.debug(f"Bean level: {bean_level}%")
    
    # 检查磨豆机是否正在工作
    if status == 1:  # 1表示正在磨粉
        logging.warning("磨豆机已经被占用了，正在磨粉中...")
        logging.info("程序停止运行")
        client.close()
        exit()
else:
    logging.error("Failed to read grinder status")
    exit()

# ---------------- 2. 发送启动命令
logging.debug("--- Sending Start Command ---")

is_ok = client.write_single_register(CMD_REG, 1)
if is_ok:
    logging.debug("Command [Start Grinding] sent sucessfully.")
else:
    logging.error("Failed to send command.")
    exit()

# ---------------- 3.监控磨粉状态
# logging.debug("--- Monitoring Grinder ---")
# while True:
#     regs = client.read_holding_registers(STATUS_REG, 2)
#     if regs:
#         status, bean_level  = regs
#         logging.debug(f"Grinder status: {status_map.get(status, 'Unknown')}")
#         logging.debug(f"Bean level: {bean_level}%")
#         if status != 1:
#             logging.debug("Grinding completed.")
#             break
#     else:
#         logging.error("Failed to read grinder status")
#     time.sleep(1)



# ---------------- 4.输出最终状态
logging.debug("--- Reading Final State ---")
regs = client.read_holding_registers(STATUS_REG, 2)
if regs:
    status, bean_level  = regs
    logging.debug(f"Grinder status: {status_map.get(status, 'Unknown')}")
    logging.debug(f"Bean level: {bean_level}%")
else:
    logging.error("Failed to read grinder status")

client.close()

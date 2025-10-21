'''
Author : Orange horrorange@qq.com
Last-modified: 2025-09-17
Used to simulate the grinder, using modbus TCP
'''

import time
import random
from pyModbusTCP.server import ModbusServer
import logging

# 设置logging
logging.basicConfig(
	level = logging.DEBUG,
	format = "%(asctime)s %(levelname)s %(message)s",
	datefmt = "%Y-%m-%d %H:%M:%S",
)

# 寄存器设置
CMD_REG = 0         # command register
STATUS_REG = 1      # status register
BEAN_LEVEL_REG = 2  # bean level register
ERROR_CODE_REG = 3  # error code register

# ----------
# CMD_REG 命令寄存器
# 2: 补充豆子
# 1: 磨粉命令写入
# 0: 处理命令结束，重置为0
# ----------
# STATUS_REG 状态寄存器
# 2: 故障
# 1: 正在工作
# 0: 空闲
# ----------
# BEAN_LEVEL_REG 评估豆量状态
# 范围: 0 ~ 100
# ----------
# ERROR_CODE_REG 故障代码寄存器
# 1: 咖啡豆不足
# 0: 无故障
# ----------

def grind():
    logging.debug("开始磨粉")
    server.data_bank.set_holding_registers(STATUS_REG, [1]) 
    current_bean_level = server.data_bank.get_holding_registers(BEAN_LEVEL_REG, 1)[0]

    # 判断豆量是否大于10
    if current_bean_level < 10:
        logging.error("豆量不足！")
        server.data_bank.set_holding_registers(STATUS_REG, [2])
        server.data_bank.set_holding_registers(ERROR_CODE_REG, [1])
        return
    else:
        logging.debug("豆量充足，开始磨粉")
        current_bean_level = current_bean_level - random.randint(5, 10)
        time.sleep(5)
        server.data_bank.set_holding_registers(BEAN_LEVEL_REG, [current_bean_level])
        logging.debug("磨粉完成，当前豆量: %d%%", current_bean_level)
        server.data_bank.set_holding_registers(STATUS_REG, [0])

def add_bean():
    logging.debug("补充豆子")
    server.data_bank.set_holding_registers(STATUS_REG, [1])
    time.sleep(2)
    server.data_bank.set_holding_registers(STATUS_REG, [0])
    server.data_bank.set_holding_registers(BEAN_LEVEL_REG, [100])
    server.data_bank.set_holding_registers(ERROR_CODE_REG, [0])

    logging.debug("补充豆子完成")

# 创建server，0.0.0.0 表示监听所有IP地址, 502是 ModBus TCP的默认端口
server = ModbusServer(host="0.0.0.0", port=502, no_block = True)
logging.debug("磨粉机开始运行")

try:
    # 启动服务
    server.start()
    logging.debug("磨粉机已启动")

    # 初始化状态
    server.data_bank.set_holding_registers(STATUS_REG, [0])       
    server.data_bank.set_holding_registers(BEAN_LEVEL_REG, [100]) 
    server.data_bank.set_holding_registers(ERROR_CODE_REG, [0])   
    logging.debug("磨粉机初始状态: 空闲, 豆量: 100%, 无故障.")


    while True:
        # 读取CMD_REG的值，判断是否有命令写入，get方法返回的是一个列表
        command = server.data_bank.get_holding_registers(CMD_REG, 1)[0]
        if command == 1:
            grind()
        elif command == 2:
            add_bean()
        # 循环时间
        server.data_bank.set_holding_registers(CMD_REG,[0])
        time.sleep(0.5)

# 异常处理
except Exception as e:
    logging.error(f"错误:{e}")
    server.stop()
    logging.error("服务关闭")
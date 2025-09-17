'''
Author : Orange horrorange@qq.com
Last-modified: 2025-09-17
Used to simulate the grinder, using modbus TCP
'''

import time
import random
from pyModbusTCP.server import ModbusServer, DataBank
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

# 创建server，0.0.0.0 表示监听所有IP地址, 502是 ModBus TCP的默认端口
server = ModbusServer(host="0.0.0.0", port=502, no_block = True)
logging.debug("Modbus Server starting...")

try:
    # 启动服务
    server.start()
    logging.debug("Modbus Server is online")

    # 初始化状态
    server.data_bank.set_holding_registers(STATUS_REG, [0])       
    server.data_bank.set_holding_registers(BEAN_LEVEL_REG, [100]) 
    server.data_bank.set_holding_registers(ERROR_CODE_REG, [0])   
    logging.debug("Grinder initial state: Idle, Bean level: 100%, No Error.")


    while True:
        # 读取CMD_REG的值，判断是否有命令写入，get方法返回的是一个列表
        command = server.data_bank.get_holding_registers(CMD_REG, 1)[0]

        if command == 1:
            logging.debug("Start grinding...")

            # 重置CMD_REG
            server.data_bank.set_holding_registers(CMD_REG, [0])
            # 获取当前豆量
            current_bean_level = server.data_bank.get_holding_registers(BEAN_LEVEL_REG, 1)[0]

            # 报错：咖啡豆不足
            if current_bean_level < 10:
                logging.error("Bean level is too low!")
                server.data_bank.set_holding_registers(STATUS_REG, [2])
                server.data_bank.set_holding_registers(ERROR_CODE_REG, [1])
            else:
                # 正常磨粉
                server.data_bank.set_holding_registers(STATUS_REG, [1])
                server.data_bank.set_holding_registers(ERROR_CODE_REG, [0])
                logging.debug("Grinder status: Grinding")

                # 每次循环的磨粉时间
                time.sleep(5)

                # 更新豆量
                new_bean_level = current_bean_level - random.randint(5, 10)
                server.data_bank.set_holding_registers(BEAN_LEVEL_REG, [new_bean_level])

                # 设置状态为空闲
                server.data_bank.set_holding_registers(STATUS_REG, [0])
                logging.debug(f"Grinding complete. New bean level: {new_bean_level}%.")
                logging.debug("Grinder status: Idle.")

        # 循环时间
        time.sleep(0.5)

# 异常处理
except Exception as e:
    logging.error(f"Error:{e}")
    logging.error("Shutting down the Modbus server...")
    server.stop()
    logging.error("Server is offline")
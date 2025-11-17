'''
Author : Orange horrorange@qq.com
Last-modified: 2025-09-17
ModBus Client test
'''

from re import S
import time
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
ERROR_CODE_REG = 3  # error code register
# 读取IP地址
SERVER_HOST = "localhost"
SERVER_PORT = 502  # 修正端口号，与模拟器保持一致

# 连接Modbus服务
client = ModbusClient(host=SERVER_HOST, port=SERVER_PORT, auto_open=False)
connection_result = client.open()

if not connection_result:
    logging.error(f"无法连接到 {SERVER_HOST}:{SERVER_PORT}")
    exit(1)
    
logging.debug(f"成功连接到 {SERVER_HOST}:{SERVER_PORT}")

def read_status():
    '''
    读取磨豆机状态
    '''
    try:
        status = client.read_holding_registers(STATUS_REG, 2)[0]
        bean_level = client.read_holding_registers(STATUS_REG, 2)[1]
        return status, bean_level
    except Exception as e:
        logging.error(f"读取状态失败: {e}")
        return None, None



def grind():
    '''
    发送磨粉命令
    '''
    status, bean_level = read_status()
    logging.debug("当前豆量: %d%%", bean_level)
    logging.debug("当前状态: %d", status)
    

    if status == 2 or bean_level < 10:
        logging.error("当前状态为故障，正在自动补豆")
        client.write_single_register(CMD_REG, 2)
        time.sleep(0.5)
        status, bean_level = read_status()
        while status != 0:
            time.sleep(0.5)
            status, bean_level = read_status()
        logging.debug("自动补豆完成")

    if status == 1:
        logging.debug("当前状态为正在工作，等待完成")
        status, bean_level = read_status()
        while status != 0:
            time.sleep(0.5)
            status, bean_level = read_status()
        logging.debug("当前工作完成，开始磨粉")

    client.write_single_register(CMD_REG, 1)
    time.sleep(0.5)
    status, bean_level = read_status()
    while status != 0:
        time.sleep(0.5)
        status, bean_level = read_status()
    return status, bean_level


for i in range(30):
    grind()

client.close()

'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-18
Used to test the ice makers, using S7 communication over TCP messages
'''
import snap7
from snap7.server import Server
from snap7.util import set_int, get_int
import time
import logging
from formatter import ColoredFormatter

# ----------------- 日志配置
logger = logging.getLogger("icemaker_sim")
logger.setLevel(logging.DEBUG)
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

# ----------------- 服务器配置
SERVER_HOST = '0.0.0.0'     # 监听所有网络接口
SERVER_PORT = 102           # 西门子PLC默认端口
RACK = 0                    # 机架号
SLOT = 1                    # 插槽号

# --- 数据块DB1的内存布局定义 ---
# 地址 | 类型 | 描述
# -----|------|----------------------
# 0    | INT  | 当前冰块库存 (单位:克)
# 2    | INT  | 设备状态 (0=待机, 1=正在制冰, 2=出冰中, 3=故障)
# 4    | INT  | 网关指令 (1=开始制冰, 2=停止制冰, 3=取冰)
# 6    | INT  | 本次取冰量 (单位:克)

db1_data = bytearray(20)    # 数据块DB1，20字节大小
# ----------------- 初始化DB1数据
set_int(db1_data, 0, 1000)  # 初始化当前冰块库存为1000克
set_int(db1_data, 2, 0)     # 初始化设备状态为待机
set_int(db1_data, 4, 0)     # 初始化网关指令为0
set_int(db1_data, 6, 0)     # 初始化本次取冰量为0克

def main():
    server = Server()
    server.register_db(1, db1_data)

    logger.info(f"S7服务器已启动，监听地址：{SERVER_HOST}:{SERVER_PORT}")
    try:
        server.start(SERVER_HOST, SERVER_PORT, RACK, SLOT)
        logger.info("S7服务器已成功启动, 等待连接...")
        while True:
            current_db_data = server.get_db(1)

            status = get_int(current_db_data, 2)
            command = get_int(current_db_data, 4)

            if command != 0:
                logger.info(f"收到网关指令：{command}")
                if command == 1:    # 开始制冰
                    logger.info("  -> 开始制冰...")
                    set_int(current_db_data, 2, 1)      # 设备状态设为正在制冰
                    server.set_db(1, current_db_data)   # 更新DB1数据
                    time.sleep(10)

                    current_ice = get_int(current_db_data, 0)
                    new_ice = min(current_ice + 1000, 1500)
                    set_int(current_db_data, 0, new_ice)  # 更新当前冰块库存
                    logger.info(f"  -> 制冰完成，当前库存：{new_ice}克")
                
                elif command == 3:  # 取冰
                    dispense_ice = get_int(current_db_data, 6)
                    logger.info(f"  -> 取冰：{dispense_ice}克")
                    set_int(current_db_data, 2, 2)      # 设备状态设为出冰中
                    server.set_db(1, current_db_data)   # 更新DB1数据
                    time.sleep(2)

                    current_ice = get_int(current_db_data, 0)
                    new_ice = max(current_ice - dispense_ice, 0)
                    if new_ice == 0:
                        logger.warning("冰块已经消耗完成！")
                    set_int(current_db_data, 0, new_ice)
                    logger.info(f"  -> 取冰完成，当前库存：{new_ice}克")
                
                # 指令处理完成后，重置指令位
                set_int(current_db_data, 4, 0)
                set_int(current_db_data, 2, 0)
                server.set_db(1, current_db_data)   # 更新DB1数据
            time.sleep(0.5)
    except KeyboardInterrupt:
        logger.debug("S7服务器已被手动停止")
        server.stop()
        logger.debug("S7服务器已成功停止")
    except Exception as e:
        logger.error(f"S7服务器运行时发生错误：{e}")
        server.stop()
        logger.debug("S7服务器已停止")

if __name__ == "__main__":
    main()                    



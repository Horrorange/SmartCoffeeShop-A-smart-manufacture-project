'''
Author: Orange horrorange@qq.com
Last-modified: 2025-09-19
测试制冰机的简单实现
'''
import snap7
from snap7.util import set_int, get_int
import time
import logging
from colorlog import ColoredFormatter

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

# ----------------- 测试配置
SERVER_IP = "127.0.0.1"     # 制冰机模拟服务器IP
SERVER_PORT = 102           # 制冰机模拟服务器端口
RACK = 0                    # 机架号
SLOT = 1                    # 插槽号

def read_current_status(client):
    """读取当前制冰机状态"""
    db_data = client.db_read(1, 0, 8)
    ice_stock = get_int(db_data, 0)
    device_status = get_int(db_data, 2)
    command = get_int(db_data, 4)
    ice_amount = get_int(db_data, 6)
    
    status_text = {0: "待机", 1: "正在制冰", 2: "出冰中", 3: "故障"}
    logger.info(f"当前状态 - 冰块库存: {ice_stock}克, 设备状态: {status_text.get(device_status, '未知')}, 当前指令: {command}, 取冰量设置: {ice_amount}克")
    return ice_stock, device_status, command, ice_amount

def make_ice(client):
    """制冰功能（加冰）"""
    logger.info("=== 开始制冰操作 ===")
    
    # 读取当前状态
    ice_stock, device_status, _, _ = read_current_status(client)
    
    if device_status != 0:
        logger.warning("设备当前不在待机状态，无法开始制冰")
        return False
    
    # 发送制冰指令
    db_data = client.db_read(1, 0, 8)
    set_int(db_data, 4, 1)  # 设置网关指令为制冰
    client.db_write(1, 4, db_data[4:6])
    logger.info("已发送制冰指令，等待制冰完成...")
    
    # 等待制冰完成
    while True:
        time.sleep(1)
        ice_stock, device_status, command, _ = read_current_status(client)
        
        if device_status == 0 and command == 0:  # 制冰完成，回到待机状态
            logger.info("制冰完成！")
            return True
        elif device_status == 3:  # 故障状态
            logger.error("制冰过程中发生故障")
            return False

def dispense_ice(client, amount):
    """取冰功能"""
    logger.info(f"=== 开始取冰操作，取冰量: {amount}克 ===")
    
    # 读取当前状态
    ice_stock, device_status, _, _ = read_current_status(client)
    
    if device_status != 0:
        logger.warning("设备当前不在待机状态，无法取冰")
        return False
    
    if ice_stock < amount:
        logger.warning(f"库存不足！当前库存: {ice_stock}克, 需要: {amount}克")
        return False
    
    # 设置取冰量
    db_data = client.db_read(1, 0, 8)
    set_int(db_data, 6, amount)
    client.db_write(1, 6, db_data[6:8])
    logger.info(f"已设置取冰量: {amount}克")
    
    # 发送取冰指令
    set_int(db_data, 4, 3)  # 设置网关指令为取冰
    client.db_write(1, 4, db_data[4:6])
    logger.info("已发送取冰指令，等待取冰完成...")
    
    # 等待取冰完成
    while True:
        time.sleep(1)
        ice_stock, device_status, command, _ = read_current_status(client)
        
        if device_status == 0 and command == 0:  # 取冰完成，回到待机状态
            logger.info("取冰完成！")
            return True
        elif device_status == 3:  # 故障状态
            logger.error("取冰过程中发生故障")
            return False

def show_menu():
    """显示操作菜单"""
    print("\n" + "="*50)
    print("🧊 制冰机控制客户端")
    print("="*50)
    print("1. 查看当前状态")
    print("2. 制冰（加冰）")
    print("3. 取冰")
    print("4. 退出")
    print("="*50)

def main():
    client = snap7.client.Client()

    try:
        logger.info(f"尝试连接制冰机模拟服务器 {SERVER_IP}:{SERVER_PORT}")
        client.connect(SERVER_IP, RACK, SLOT)
        logger.info("成功连接制冰机模拟服务器")

        # 读取初始状态
        read_current_status(client)
        
        # 交互式菜单
        while True:
            show_menu()
            try:
                choice = input("请选择操作 (1-4): ").strip()
                
                if choice == "1":
                    # 查看当前状态
                    read_current_status(client)
                    
                elif choice == "2":
                    # 制冰
                    if make_ice(client):
                        logger.info("制冰操作成功完成")
                    else:
                        logger.error("制冰操作失败")
                        
                elif choice == "3":
                    # 取冰
                    try:
                        amount = int(input("请输入取冰量（克）: ").strip())
                        if amount <= 0:
                            logger.warning("取冰量必须大于0")
                            continue
                        
                        if dispense_ice(client, amount):
                            logger.info("取冰操作成功完成")
                        else:
                            logger.error("取冰操作失败")
                    except ValueError:
                        logger.error("请输入有效的数字")
                        
                elif choice == "4":
                    # 退出
                    logger.info("退出程序")
                    break
                    
                else:
                    print("无效选择，请输入1-4")
                    
            except KeyboardInterrupt:
                logger.info("用户中断操作")
                break
            except Exception as e:
                logger.error(f"操作过程中发生错误: {e}")
                
    except RuntimeError as e:
        logger.error(f"连接制冰机模拟服务器失败: {e}")
    except Exception as e:
        logger.error(f"程序运行过程中发生错误: {e}")
    finally:
        if client.get_connected():
            client.disconnect()
            logger.info("已断开与制冰机模拟服务器的连接")

if __name__ == "__main__":
    main()

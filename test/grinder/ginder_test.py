import time
from pyModbusTCP.client import ModbusClient
import logging

# 设置日志
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

class GrinderTester:
    def __init__(self, host='localhost', port=502):
        self.client = ModbusClient(host=host, port=port)
        self.registers = {
            'CMD': 0,        # 命令寄存器
            'STATUS': 1,     # 状态寄存器  
            'BEAN_LEVEL': 2, # 豆量寄存器
            'ERROR_CODE': 3  # 错误代码寄存器
        }
        
    def connect(self):
        """连接到Modbus服务器"""
        if self.client.open():
            logging.info("成功连接到磨粉机模拟器")
            return True
        else:
            logging.error("无法连接到磨粉机模拟器")
            return False
            
    def read_registers(self):
        """读取所有寄存器状态"""
        try:
            registers = self.client.read_holding_registers(0, 4)
            if registers:
                return {
                    'command': registers[0],
                    'status': registers[1],
                    'bean_level': registers[2],
                    'error_code': registers[3]
                }
            else:
                logging.error("读取寄存器失败")
                return None
        except Exception as e:
            logging.error(f"读取寄存器时出错: {e}")
            return None
            
    def send_command(self, command):
        """发送命令到磨粉机"""
        try:
            if self.client.write_single_register(self.registers['CMD'], command):
                logging.info(f"命令 {command} 发送成功")
                return True
            else:
                logging.error(f"命令 {command} 发送失败")
                return False
        except Exception as e:
            logging.error(f"发送命令时出错: {e}")
            return False
            
    def wait_for_status(self, target_status, timeout=10):
        """等待设备达到特定状态"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            status = self.read_registers()
            if status and status['status'] == target_status:
                return True
            time.sleep(0.5)
        logging.warning(f"等待状态 {target_status} 超时")
        return False
        
    def print_status(self):
        """打印当前状态"""
        status = self.read_registers()
        if status:
            status_map = {0: "空闲", 1: "工作中", 2: "故障"}
            error_map = {0: "无故障", 1: "咖啡豆不足"}
            
            logging.info("=== 磨粉机状态 ===")
            logging.info(f"命令寄存器: {status['command']}")
            logging.info(f"状态: {status['status']} ({status_map.get(status['status'], '未知')})")
            logging.info(f"豆量: {status['bean_level']}%")
            logging.info(f"错误代码: {status['error_code']} ({error_map.get(status['error_code'], '未知')})")
            logging.info("=================")
        return status

    def test_normal_grinding(self):
        """测试正常磨粉流程"""
        logging.info("\n" + "="*50)
        logging.info("开始测试正常磨粉流程")
        logging.info("="*50)
        
        # 1. 检查初始状态
        self.print_status()
        
        # 2. 发送磨粉命令
        if not self.send_command(1):
            return False
            
        # 3. 等待设备开始工作
        if self.wait_for_status(1):
            logging.info("磨粉机开始工作")
            self.print_status()
        else:
            logging.error("磨粉机未按预期开始工作")
            return False
            
        # 4. 等待工作完成
        time.sleep(6)  # 等待磨粉完成
        if self.wait_for_status(0, 15):
            logging.info("磨粉完成")
            self.print_status()
            return True
        else:
            logging.error("磨粉机未按预期完成工作")
            return False
            
    def test_low_bean_level(self):
        """测试豆量不足的情况"""
        logging.info("\n" + "="*50)
        logging.info("开始测试豆量不足情况")
        logging.info("="*50)
        
        # 持续磨粉直到豆量不足
        bean_level = 100
        grinding_count = 0
        
        while bean_level > 10 and grinding_count < 15:  # 安全限制
            logging.info(f"\n--- 第 {grinding_count + 1} 次磨粉 ---")
            
            if not self.send_command(1):
                return False
                
            # 等待磨粉完成
            time.sleep(6)
            
            status = self.print_status()
            if not status:
                return False
                
            bean_level = status['bean_level']
            grinding_count += 1
            
            if status['error_code'] == 1:
                logging.warning("检测到豆量不足错误")
                break
        
        # 验证错误状态
        status = self.print_status()
        if status and status['error_code'] == 1 and status['status'] == 2:
            logging.info("✓ 豆量不足测试通过")
            return True
        else:
            logging.error("✗ 豆量不足测试失败")
            return False
            
    def test_refill_beans(self):
        """测试补充豆子功能"""
        logging.info("\n" + "="*50)
        logging.info("开始测试补充豆子功能")
        logging.info("="*50)
        
        # 发送补充豆子命令
        if not self.send_command(2):
            return False
            
        # 等待补充完成
        time.sleep(3)
        
        status = self.print_status()
        if status and status['bean_level'] == 100 and status['error_code'] == 0:
            logging.info("✓ 补充豆子测试通过")
            return True
        else:
            logging.error("✗ 补充豆子测试失败")
            return False
            
    def run_all_tests(self):
        """运行所有测试"""
        if not self.connect():
            return False
            
        tests_passed = 0
        total_tests = 3
        
        try:
            # 测试1: 正常磨粉
            if self.test_normal_grinding():
                tests_passed += 1
                
            # 测试2: 豆量不足
            if self.test_low_bean_level():
                tests_passed += 1
                
            # 测试3: 补充豆子
            if self.test_refill_beans():
                tests_passed += 1
                
        except Exception as e:
            logging.error(f"测试过程中出现异常: {e}")
            
        # 输出测试结果
        logging.info("\n" + "="*50)
        logging.info(f"测试完成: {tests_passed}/{total_tests} 通过")
        logging.info("="*50)
        
        return tests_passed == total_tests

def main():
    """主函数"""
    tester = GrinderTester(host='localhost', port=502)
    
    if tester.run_all_tests():
        logging.info("🎉 所有测试通过！")
    else:
        logging.error("❌ 部分测试失败！")

if __name__ == "__main__":
    main()

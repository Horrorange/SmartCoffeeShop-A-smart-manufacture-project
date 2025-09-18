# 🤝 贡献指南 (Contributing Guide)

感谢您对 Smart Coffee Shop 项目的关注！我们欢迎所有形式的贡献，无论是代码、文档、测试还是反馈建议。

## 📋 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境设置](#开发环境设置)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [问题报告](#问题报告)
- [功能请求](#功能请求)
- [文档贡献](#文档贡献)
- [测试指南](#测试指南)

## 🌟 行为准则

参与本项目即表示您同意遵守我们的行为准则：

- **尊重他人**：对所有参与者保持友善和尊重
- **包容性**：欢迎不同背景和经验水平的贡献者
- **建设性**：提供建设性的反馈和建议
- **专业性**：保持专业的沟通方式
- **学习态度**：保持开放的学习和分享态度

## 🚀 如何贡献

### 贡献类型

我们欢迎以下类型的贡献：

1. **🐛 Bug 修复**：修复现有功能中的问题
2. **✨ 新功能**：添加新的功能或改进现有功能
3. **📚 文档改进**：改进文档、添加示例或教程
4. **🧪 测试**：添加或改进测试用例
5. **🎨 代码优化**：性能优化、代码重构
6. **🌐 国际化**：添加多语言支持
7. **📦 依赖管理**：更新或优化项目依赖

### 贡献流程

1. **Fork 项目**到您的 GitHub 账户
2. **创建分支**用于您的更改
3. **进行开发**并测试您的更改
4. **提交代码**并推送到您的 Fork
5. **创建 Pull Request**到主仓库

## 🛠️ 开发环境设置

### 环境要求

- Python 3.7+
- Git
- 推荐使用虚拟环境

### 安装步骤

```bash
# 1. Fork 并克隆项目
git clone https://github.com/YOUR_USERNAME/SmartCoffeeShop-A-smart-manufacture-project.git
cd SmartCoffeeShop-A-smart-manufacture-project

# 2. 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/macOS
# 或
venv\Scripts\activate     # Windows

# 3. 安装依赖
pip install -r requirements.txt

# 4. 安装开发依赖（可选）
pip install -e ".[dev]"

# 5. 验证安装
python script/grinder/grinder_sim.py --help
```

### 开发工具推荐

- **IDE**: VS Code, PyCharm, 或您喜欢的编辑器
- **代码格式化**: Black
- **代码检查**: Flake8
- **类型检查**: MyPy
- **测试**: Pytest

## 📝 代码规范

### Python 代码风格

我们遵循 [PEP 8](https://pep8.org/) 代码风格指南：

```bash
# 使用 Black 格式化代码
black script/ test/

# 使用 Flake8 检查代码
flake8 script/ test/

# 使用 isort 排序导入
isort script/ test/
```

### 代码质量要求

1. **可读性**：代码应该清晰易懂
2. **注释**：复杂逻辑需要适当注释
3. **文档字符串**：所有公共函数和类需要文档字符串
4. **错误处理**：适当的异常处理
5. **测试覆盖**：新功能需要相应的测试

### 示例代码风格

```python
def handle_grinder_command(command: int, client: ModbusClient) -> bool:
    """
    处理磨粉机命令
    
    Args:
        command: 命令代码 (0=停止, 1=开始磨粉)
        client: Modbus客户端实例
        
    Returns:
        bool: 命令执行是否成功
        
    Raises:
        ConnectionError: 当无法连接到磨粉机时
    """
    try:
        # 发送命令到磨粉机
        success = client.write_single_register(CMD_REG, command)
        if not success:
            logger.error(f"Failed to send command {command} to grinder")
            return False
            
        logger.info(f"Successfully sent command {command} to grinder")
        return True
        
    except Exception as e:
        logger.error(f"Error sending command to grinder: {e}")
        raise ConnectionError(f"Failed to communicate with grinder: {e}")
```

## 📤 提交规范

### 提交消息格式

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 提交类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式化（不影响功能）
- `refactor`: 代码重构
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

### 提交示例

```bash
# 新功能
feat(grinder): add automatic bean level monitoring

# Bug 修复
fix(coffee-machine): resolve inventory synchronization issue

# 文档更新
docs: update installation instructions for Windows

# 测试
test(agent): add unit tests for order processing logic
```

## 🔄 Pull Request 流程

### 创建 Pull Request

1. **确保分支是最新的**：
   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **运行测试**：
   ```bash
   python -m pytest test/
   ```

3. **检查代码质量**：
   ```bash
   black --check script/ test/
   flake8 script/ test/
   ```

4. **创建 Pull Request**并填写模板

### Pull Request 模板

```markdown
## 📝 变更描述
简要描述您的更改内容

## 🎯 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 代码重构
- [ ] 测试改进

## 🧪 测试
- [ ] 已添加新的测试用例
- [ ] 所有现有测试通过
- [ ] 手动测试通过

## 📋 检查清单
- [ ] 代码遵循项目规范
- [ ] 已更新相关文档
- [ ] 提交消息符合规范
- [ ] 已解决所有冲突

## 🔗 相关 Issue
Closes #(issue number)
```

### 代码审查

所有 Pull Request 都需要经过代码审查：

1. **自动检查**：CI/CD 流水线会自动运行测试
2. **人工审查**：至少一位维护者会审查您的代码
3. **反馈处理**：根据反馈进行必要的修改
4. **合并**：审查通过后会被合并到主分支

## 🐛 问题报告

### 报告 Bug

发现 Bug 时，请创建详细的 Issue：

1. **使用 Bug 报告模板**
2. **提供复现步骤**
3. **包含错误信息和日志**
4. **说明预期行为**
5. **提供环境信息**

### Bug 报告模板

```markdown
## 🐛 Bug 描述
简要描述遇到的问题

## 🔄 复现步骤
1. 启动磨粉机模拟器
2. 运行客户端测试
3. 观察错误

## 💭 预期行为
描述您期望发生的情况

## 📸 实际行为
描述实际发生的情况

## 🖥️ 环境信息
- OS: [e.g. macOS 12.0]
- Python: [e.g. 3.9.0]
- 项目版本: [e.g. 1.0.0]

## 📋 附加信息
添加任何其他有助于解决问题的信息
```

## 💡 功能请求

### 提出新功能

1. **检查现有 Issue**：确保功能未被提出
2. **使用功能请求模板**
3. **详细描述用例**
4. **考虑实现复杂度**

### 功能请求模板

```markdown
## 🚀 功能描述
简要描述您希望添加的功能

## 💭 动机和用例
解释为什么需要这个功能，它解决了什么问题

## 📋 详细设计
详细描述功能的工作方式

## 🔄 替代方案
描述您考虑过的其他解决方案

## 📎 附加信息
添加任何其他相关信息或截图
```

## 📚 文档贡献

### 文档类型

- **README**: 项目概述和快速开始
- **API 文档**: 函数和类的详细说明
- **教程**: 分步指导和示例
- **FAQ**: 常见问题解答

### 文档规范

1. **清晰简洁**：使用简单明了的语言
2. **结构化**：使用标题和列表组织内容
3. **示例丰富**：提供实际的代码示例
4. **保持更新**：确保文档与代码同步

## 🧪 测试指南

### 测试类型

1. **单元测试**：测试单个函数或方法
2. **集成测试**：测试组件间的交互
3. **端到端测试**：测试完整的工作流程

### 运行测试

```bash
# 运行所有测试
python -m pytest

# 运行特定测试文件
python -m pytest test/test_grinder.py

# 运行带覆盖率的测试
python -m pytest --cov=script

# 运行特定标记的测试
python -m pytest -m "not slow"
```

### 编写测试

```python
import pytest
from script.grinder.grinder_sim import GrinderSimulator

class TestGrinderSimulator:
    def setup_method(self):
        """每个测试方法前的设置"""
        self.grinder = GrinderSimulator()
    
    def test_initial_state(self):
        """测试初始状态"""
        assert self.grinder.status == 0  # 空闲状态
        assert self.grinder.bean_level == 100  # 满豆
    
    def test_start_grinding(self):
        """测试开始磨粉"""
        result = self.grinder.start_grinding()
        assert result is True
        assert self.grinder.status == 1  # 磨粉中
    
    @pytest.mark.parametrize("bean_level", [0, 5, 10])
    def test_low_bean_level(self, bean_level):
        """测试低豆量情况"""
        self.grinder.bean_level = bean_level
        result = self.grinder.start_grinding()
        assert result is False
```

## 🏷️ 版本发布

### 版本号规范

我们使用 [语义化版本](https://semver.org/)：

- **主版本号**：不兼容的 API 修改
- **次版本号**：向下兼容的功能性新增
- **修订号**：向下兼容的问题修正

### 发布流程

1. **更新版本号**：在 `setup.py` 中更新版本
2. **更新 CHANGELOG**：记录所有变更
3. **创建 Release**：在 GitHub 上创建新的 Release
4. **发布到 PyPI**：（如果适用）

## 🆘 获取帮助

如果您在贡献过程中遇到问题，可以通过以下方式获取帮助：

1. **查看文档**：首先查看项目文档和 FAQ
2. **搜索 Issue**：查看是否有类似的问题已被讨论
3. **创建 Issue**：如果找不到答案，创建新的 Issue
4. **联系维护者**：通过邮件 horrorange@qq.com 联系

## 🙏 致谢

感谢所有为 Smart Coffee Shop 项目做出贡献的开发者！您的贡献让这个项目变得更好。

### 贡献者列表

- **Orange** - 项目创建者和主要维护者

---

再次感谢您的贡献！让我们一起构建更好的智能咖啡店系统！ ☕️✨
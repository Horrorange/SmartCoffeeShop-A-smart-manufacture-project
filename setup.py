#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Smart Coffee Shop - Setup Script
一个基于工业物联网协议的智能咖啡店制造系统

Author: Orange
Email: horrorange@qq.com
Last-modified: 2025-09-18
License: MIT
"""

from setuptools import setup, find_packages
import os

# 读取README文件作为长描述
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), "README.md")
    try:
        with open(readme_path, "r", encoding="utf-8") as fh:
            return fh.read()
    except FileNotFoundError:
        return "一个基于工业物联网协议的智能咖啡店制造系统，模拟真实的咖啡制作流程。"

# 读取requirements文件
def read_requirements():
    requirements_path = os.path.join(os.path.dirname(__file__), "requirements.txt")
    try:
        with open(requirements_path, "r", encoding="utf-8") as fh:
            requirements = []
            for line in fh:
                line = line.strip()
                # 跳过注释和空行
                if line and not line.startswith("#"):
                    requirements.append(line)
            return requirements
    except FileNotFoundError:
        return ["pyModbusTCP>=0.2.0", "colorlog>=6.0.0"]

# 项目版本
VERSION = "1.0.0"

setup(
    # 基本信息
    name="smart-coffee-shop",
    version=VERSION,
    author="Orange",
    author_email="horrorange@qq.com",
    
    # 项目描述
    description="一个基于工业物联网协议的智能咖啡店制造系统",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    
    # 项目链接
    url="https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project",
    project_urls={
        "Bug Reports": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project/issues",
        "Source": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project",
        "Documentation": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project/blob/main/README.md",
        "Changelog": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project/releases",
    },
    
    # 包配置
    packages=find_packages(include=["script", "script.*", "test", "test.*"]),
    include_package_data=True,
    zip_safe=False,
    
    # 分类器
    classifiers=[
        # 开发状态
        "Development Status :: 4 - Beta",
        
        # 目标受众
        "Intended Audience :: Developers",
        "Intended Audience :: Education",
        "Intended Audience :: Manufacturing",
        "Intended Audience :: Science/Research",
        
        # 许可证
        "License :: OSI Approved :: MIT License",
        
        # 操作系统
        "Operating System :: OS Independent",
        "Operating System :: POSIX :: Linux",
        "Operating System :: Microsoft :: Windows",
        "Operating System :: MacOS",
        
        # Python版本支持
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Python :: 3 :: Only",
        
        # 主题分类
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Hardware",
        "Topic :: Communications",
        "Topic :: Internet",
        "Topic :: Scientific/Engineering",
        "Topic :: System :: Networking",
        "Topic :: Software Development :: Testing :: Mocking",
        "Topic :: Education",
    ],
    
    # Python版本要求
    python_requires=">=3.7",
    
    # 依赖管理
    install_requires=read_requirements(),
    
    # 可选依赖
    extras_require={
        "dev": [
            "pytest>=6.0.0",
            "pytest-cov>=2.0.0",
            "black>=21.0.0",
            "flake8>=3.8.0",
            "isort>=5.0.0",
            "mypy>=0.900",
        ],
        "docs": [
            "sphinx>=4.0.0",
            "sphinx-rtd-theme>=1.0.0",
            "myst-parser>=0.15.0",
        ],
        "testing": [
            "pytest>=6.0.0",
            "pytest-cov>=2.0.0",
            "pytest-mock>=3.0.0",
            "coverage>=5.0.0",
        ],
    },
    
    # 控制台脚本入口点
    entry_points={
        "console_scripts": [
            "grinder-sim=script.grinder.grinder_sim:main",
            "coffeemachine-sim=script.coffeemachine.coffeemachine_sim:main",
            "grinder-client=test.grinder.client_test:main",
            "coffee-agent=test.coffeemachine.coffeemachine_agent:main",
        ],
    },
    
    # 关键词
    keywords=[
        "modbus", "tcp", "coffee", "grinder", "simulation", "iot", 
        "industrial", "manufacturing", "smart", "automation", "edge-computing",
        "industry-4.0", "protocol", "communication", "embedded", "hardware"
    ],
    
    # 数据文件
    package_data={
        "": ["*.md", "*.txt", "*.yml", "*.yaml", "*.json"],
    },
    
    # 平台特定选项
    options={
        "bdist_wheel": {
            "universal": False,  # 不是通用wheel，因为可能有平台特定的依赖
        }
    },
)
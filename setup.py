#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Smart Coffee Shop - Setup Script
Author: Orange horrorange@qq.com
Last-modified: 2025-09-17
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("requirements.txt", "r", encoding="utf-8") as fh:
    requirements = [line.strip() for line in fh if line.strip() and not line.startswith("#")]

setup(
    name="smart-coffee-shop",
    version="1.0.0",
    author="Orange",
    author_email="horrorange@qq.com",
    description="一个基于Modbus TCP协议的智能咖啡磨豆机模拟系统",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Intended Audience :: Education",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Hardware",
        "Topic :: Communications",
    ],
    python_requires=">=3.7",
    install_requires=requirements,
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "black>=21.0",
            "flake8>=3.8",
        ],
    },
    entry_points={
        "console_scripts": [
            "grinder-sim=script.grinder.grinder_sim:main",
            "grinder-client=script.grinder.client_test:main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
    keywords="modbus tcp coffee grinder simulation iot industrial",
    project_urls={
        "Bug Reports": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project/issues",
        "Source": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project",
        "Documentation": "https://github.com/yourusername/SmartCoffeeShop-A-smart-manufacture-project/blob/main/README.md",
    },
)
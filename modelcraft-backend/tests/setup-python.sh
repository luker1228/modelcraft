#!/bin/bash

# ModelCraft Go - Python 环境初始化脚本
# 使用 pyenv 管理 Python 版本

echo "🚀 开始设置 ModelCraft Go Python 环境..."

# 检查 pyenv 是否已安装
if ! command -v pyenv &> /dev/null; then
    echo "❌ pyenv 未安装，正在安装..."
    
    # 安装 pyenv
    curl https://pyenv.run | bash
    
    # 添加 pyenv 到 shell 配置
    echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
    echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
    echo 'eval "$(pyenv init -)"' >> ~/.bashrc
    
    # 重新加载配置
    source ~/.bashrc
    echo "✅ pyenv 安装完成"
else
    echo "✅ pyenv 已安装"
fi

# 检查是否有所需的 Python 版本
REQUIRED_PYTHON="3.9.18"
if pyenv versions | grep -q "$REQUIRED_PYTHON"; then
    echo "✅ Python $REQUIRED_PYTHON 已安装"
else
    echo "📥 正在安装 Python $REQUIRED_PYTHON..."
    
    # 安装依赖（Ubuntu/Debian）
    if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y make build-essential libssl-dev zlib1g-dev \
            libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm \
            libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev \
            libffi-dev liblzma-dev
    fi
    
    # 安装 Python
    pyenv install $REQUIRED_PYTHON
    echo "✅ Python $REQUIRED_PYTHON 安装完成"
fi

# 设置项目 Python 版本
echo "🔧 设置项目 Python 版本为 $REQUIRED_PYTHON..."
pyenv local $REQUIRED_PYTHON

# 创建虚拟环境
echo "📦 创建虚拟环境..."
python -m venv .venv

# 激活虚拟环境
echo "🔌 激活虚拟环境..."
source .venv/bin/activate

# 安装项目依赖
echo "📚 安装项目依赖..."
pip install --upgrade pip
pip install -e .

# 安装开发依赖
echo "🔧 安装开发依赖..."
pip install -e .[dev]

# 验证安装
echo "✅ 验证安装..."
python --version
pip list | grep modelcraft-go-tests

echo ""
echo "🎉 Python 环境设置完成！"
echo ""
echo "📋 使用说明："
echo "   激活虚拟环境: source .venv/bin/activate"
echo "   运行测试: pytest"
echo "   格式化代码: black tests/"
echo "   检查代码质量: flake8 tests/"
echo ""
echo "💡 提示：每次进入项目目录时，pyenv 会自动切换到正确的 Python 版本"
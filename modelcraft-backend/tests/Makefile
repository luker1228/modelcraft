# ModelCraft Go - 测试专用 Makefile
# 所有Python测试相关的命令都在这里管理

.PHONY: help test clean-data setup-python format lint

# 默认目标：自动检查Python环境并执行所有测试
default:
	@echo "🚀 自动执行所有测试..."
	if [ -d ".venv" ]; then \
		echo "✅ Python环境已就绪"; \
		$(MAKE) test; \
	else \
		echo "🐍 设置Python环境..."; \
		$(MAKE) setup-python; \
		echo "✅ Python环境设置完成"; \
		$(MAKE) test; \
	fi

# 帮助信息
help:
	@echo "📋 ModelCraft Go 测试命令"
	@echo ""
	@echo "🐍 核心测试命令:"
	@echo "  make                    - 自动检查环境并执行所有测试"
	@echo "  make test               - 运行所有测试（Python、集成、自动化）"
	@echo ""
	@echo "🔧 环境管理:"
	@echo "  make setup-python       - 设置 Python 环境"
	@echo "  make install-deps       - 安装 Python 依赖"
	@echo "  make activate           - 显示虚拟环境激活命令"
	@echo "  make clean-data         - 清理数据库数据"
	@echo ""
	@echo "🎨 代码质量:"
	@echo "  make format             - 格式化 Python 代码"
	@echo "  make lint               - 检查 Python 代码质量"
	@echo "  make python-test-html   - 生成 HTML 测试报告"
	@echo ""
	@echo "🚀 自动化测试:"
	@echo "  make auto-test          - 运行自动化测试（默认环境）"
	@echo "  make auto-test-env      - 运行指定环境的自动化测试"

# ========================
# 🐍 测试命令
# ========================

# 运行所有测试（合并了Python、集成、自动化测试）
test:
	@echo "🧪 运行所有测试..."
	if [ -d ".venv" ]; then \
		echo "� 运行Python单元测试..."; \
		source .venv/bin/activate && pytest; \
		echo "🔗 运行集成测试..."; \
		source .venv/bin/activate && pytest automated/ -v -m integration; \
		echo "🤖 运行自动化测试..."; \
		./run_smart_tests.sh local; \
		echo "✅ 所有测试完成"; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# ========================
# 🔧 Python 环境管理
# ========================

# 设置 Python 环境
setup-python:
	@echo "🚀 设置 Python 环境..."
	chmod +x ./setup-python.sh
	./setup-python.sh

# 安装 Python 依赖
install-deps:
	@echo "📚 安装 Python 依赖..."
	if [ -d ".venv" ]; then \
		source .venv/bin/activate && pip install -e .. && pip install -e ..[dev]; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# 激活虚拟环境
activate:
	@echo "🔌 激活虚拟环境..."
	if [ -d ".venv" ]; then \
		echo "运行以下命令激活虚拟环境:"; \
		echo "source .venv/bin/activate"; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# 清理数据库数据
clean-data:
	@echo "🧹 清理数据库数据..."
	if [ -d ".venv" ]; then \
		source .venv/bin/activate && python cleanup_test_data.py; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# ========================
# 🎨 代码质量
# ========================

# 格式化 Python 代码
format:
	@echo "🎨 格式化 Python 代码..."
	if [ -d ".venv" ]; then \
		source .venv/bin/activate && black .; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# 检查代码质量
lint:
	@echo "🔍 检查代码质量..."
	if [ -d ".venv" ]; then \
		source .venv/bin/activate && flake8 .; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# 运行 Python 测试并生成 HTML 报告
python-test-html:
	@echo "📊 运行 Python 测试并生成 HTML 报告..."
	if [ -d ".venv" ]; then \
		source .venv/bin/activate && pytest --html=reports/report.html; \
		echo "✅ 报告生成在 reports/report.html"; \
	else \
		echo "❌ 请先运行 'make setup-python' 创建虚拟环境"; \
	fi

# ========================
# 🚀 自动化测试
# ========================

# 运行自动化测试（默认环境）
auto-test:
	@echo "🧪 运行自动化测试（默认环境：local）..."
	./run_smart_tests.sh local

# 运行指定环境的自动化测试
auto-test-env:
	@echo "🧪 运行自动化测试（环境：$(ENV)）..."
	./run_smart_tests.sh $(ENV) $(if $(CLEANUP),$(if $(filter no,$(CLEANUP)),--no-cleanup,),)

# ========================
# 🧹 清理命令
# ========================

# 清理所有测试相关文件
clean:
	@echo "🧹 清理所有测试文件..."
	rm -rf .pytest_cache/
	rm -rf __pycache__/
	rm -rf automated/__pycache__/
	rm -rf design/__pycache__/
	rm -rf runtime/__pycache__/
	rm -rf reports/
	mkdir -p reports
	@echo "✅ 清理完成"
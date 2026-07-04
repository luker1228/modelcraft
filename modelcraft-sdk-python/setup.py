from setuptools import setup, find_packages

setup(
    name="modelcraft",
    version="0.1.0",
    packages=find_packages(),
    install_requires=["requests>=2.28"],
    python_requires=">=3.9",
    description="ModelCraft Python SDK — RLS-aware data API client",
    author="ModelCraft",
)

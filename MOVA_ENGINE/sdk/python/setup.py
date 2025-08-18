from setuptools import find_packages, setup

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="mova-engine-sdk",
    version="1.0.0",
    author="MOVA Engine Team",
    author_email="team@mova-engine.com",
    description="Python SDK for MOVA Automation Engine",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/mova-engine/mova-engine",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Systems Administration",
        "Topic :: Software Development :: Libraries :: Application Frameworks",
    ],
    python_requires=">=3.8",
    install_requires=[
        "requests>=2.28.0",
        "pydantic>=2.0.0",
        "typing-extensions>=4.0.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "pytest-mock>=3.10.0",
            "responses>=0.23.0",
            "black>=23.0.0",
            "flake8>=6.0.0",
            "mypy>=1.0.0",
            "types-requests>=2.28.0",
        ],
    },
    keywords="mova automation workflow engine python sdk",
    project_urls={
        "Bug Reports": "https://github.com/mova-engine/mova-engine/issues",
        "Source": "https://github.com/mova-engine/mova-engine",
        "Documentation": "https://docs.mova-engine.com",
    },
)

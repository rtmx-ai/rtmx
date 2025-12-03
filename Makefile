# RTMX Development Makefile
# Requirements Traceability Matrix toolkit

.PHONY: help install dev test lint format typecheck clean build publish

PYTHON := python3
VENV := .venv
PIP := $(VENV)/bin/pip
PYTEST := $(VENV)/bin/pytest
RUFF := $(VENV)/bin/ruff
MYPY := $(VENV)/bin/mypy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Installation targets

$(VENV)/bin/activate:
	$(PYTHON) -m venv $(VENV)
	$(PIP) install --upgrade pip

install: $(VENV)/bin/activate ## Install package in production mode
	$(PIP) install -e .

dev: $(VENV)/bin/activate ## Install package with dev dependencies
	$(PIP) install -e ".[dev]"

# Testing targets

test: ## Run tests
	$(PYTEST) tests/ -v

test-cov: ## Run tests with coverage
	$(PYTEST) tests/ -v --cov=rtmx --cov-report=term-missing --cov-report=html

test-fast: ## Run tests without coverage
	$(PYTEST) tests/ -v --no-cov

# Code quality targets

lint: ## Run linter
	$(RUFF) check src/ tests/

lint-fix: ## Run linter and fix issues
	$(RUFF) check --fix src/ tests/

format: ## Format code
	$(RUFF) format src/ tests/

format-check: ## Check code formatting
	$(RUFF) format --check src/ tests/

typecheck: ## Run type checker
	$(MYPY) src/rtmx/

check: lint format-check typecheck ## Run all checks

# Build targets

clean: ## Clean build artifacts
	rm -rf build/ dist/ *.egg-info src/*.egg-info
	rm -rf .pytest_cache/ .mypy_cache/ .ruff_cache/
	rm -rf htmlcov/ .coverage coverage.xml
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete

build: clean ## Build package
	$(VENV)/bin/python -m build

publish: build ## Publish to PyPI
	$(VENV)/bin/twine upload dist/*

publish-test: build ## Publish to TestPyPI
	$(VENV)/bin/twine upload --repository testpypi dist/*

# RTM self-test targets (dogfooding)

rtm: ## Show RTM status (summary)
	$(VENV)/bin/rtmx status

rtm-v: ## Show RTM status (categories)
	$(VENV)/bin/rtmx status -v

rtm-vv: ## Show RTM status (subcategories)
	$(VENV)/bin/rtmx status -vv

rtm-vvv: ## Show RTM status (all requirements)
	$(VENV)/bin/rtmx status -vvv

backlog: ## Show backlog
	$(VENV)/bin/rtmx backlog

cycles: ## Check for dependency cycles
	$(VENV)/bin/rtmx cycles

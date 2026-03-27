.PHONY: install lint fmt test cover build clean

install:  ## Install package with dev dependencies
	pip install -e ".[dev]"

lint:  ## Run all linters
	ruff check .
	ruff format --check .
	mypy src/

fmt:  ## Auto-format code
	ruff check --fix .
	ruff format .

test:  ## Run tests
	pytest

cover:  ## Run tests with coverage (90% threshold)
	pytest --cov --cov-report=term-missing
	coverage report --fail-under=90

build:  ## Build package
	python -m build

clean:  ## Remove build artifacts
	rm -rf dist/ build/ *.egg-info src/*.egg-info coverage.xml .coverage .mypy_cache .pytest_cache .ruff_cache

help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

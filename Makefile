.PHONY: install
install: ## Install the environment
	@echo "🚀 Creating virtual environment using pyenv and PDM"
	@pdm install

.PHONY: check
check: ## Run code quality tools.
	@echo "🚀 Checking pdm lock file consistency with 'pyproject.toml': Running pdm lock --check"
	@pdm lock --check
	@echo "🚀 Linting code: Running pre-commit"
	@pdm run pre-commit run -a
	@echo "🚀 Linting with ruff"
	@pdm run ruff check tests httpie_websockets.py --config pyproject.toml
	@echo "🚀 Checking for obsolete dependencies: Running deptry"
	@pdm run deptry .

.PHONY: format
format: ## Format code with ruff and isort
	@echo "🚀 Formatting code: Running ruff"
	@pdm run ruff check --fix . --config pyproject.toml
	@echo "🚀 Formatting code: Running isort"
	@pdm run isort . --settings-path pyproject.toml

.PHONY: test
test: ## Test the code with pytest.
	@echo "🚀 Testing code: Running pytest"
	@pdm run pytest --cov --cov-config=pyproject.toml --cov-report=xml tests

.PHONY: build
build: clean-build ## Build wheel file
	@echo "🚀 Creating wheel file"
	@pdm build

.PHONY: clean-build
clean-build: ## clean build artifacts
	@rm -rf dist

.PHONY: publish
publish: ## publish a release to pypi.
	@echo "🚀 Publishing."
	@pdm publish --username __token__

.PHONY: publish-test
publish-test: ## publish a release to testpypi
	@echo "🚀 Publishing to testpypi."
	@pdm publish -r testpypi --username __token__

.PHONY: build-and-publish
build-and-publish: build publish ## Build and publish.

.PHONY: docs-test
docs-test: ## Test if documentation can be built without warnings or errors
	@mkdocs build -s

.PHONY: docs
docs: ## Build and serve the documentation
	@mkdocs serve

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

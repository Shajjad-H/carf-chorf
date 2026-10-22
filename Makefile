# =========================
# Help
# =========================
.PHONY: help
help: ## Show this help
	@echo
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo

# =========================
# Development
# =========================
.PHONY: dev
dev: ## Run the server with hot reload (air)
	@air


# =========================
# Development
# =========================
.PHONY: dev_dash
dev_dash: ## Run the server with hot reload (air)
	@air -c .air.dash.toml 
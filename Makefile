# 🏦 Barclays-Grade Transaction Monitoring System
# Makefile for development, testing, and deployment

.PHONY: help install test build deploy clean load-test demo
.DEFAULT_GOAL := help

# Configuration
PROJECT_NAME := barclays-tx-monitoring
DOCKER_REGISTRY := ghcr.io
VERSION := $(shell git describe --tags --always --dirty)
GO_VERSION := $(shell go version | awk '{print $$3}')

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)🏦 Barclays-Grade Transaction Monitoring System$(NC)"
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Commands
install: ## Install dependencies and tools
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/segmentio/kafka-go@latest
	go install github.com/redis/go-redis/v9@latest
	@echo "$(GREEN)Dependencies installed successfully!$(NC)"

dev: ## Start development environment
	@echo "$(GREEN)Starting development environment...$(NC)"
	cd infra/local && docker-compose up -d
	cd apps/ingestion-service && go run main.go

dev-stop: ## Stop development environment
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	cd infra/local && docker-compose down

# Testing Commands
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	cd apps/ingestion-service && go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests completed!$(NC)"

test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	cd apps/ingestion-service && go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=apps/ingestion-service/coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	cd apps/ingestion-service && golangci-lint run
	@echo "$(GREEN)Linting completed!$(NC)"

security-scan: ## Run security scans
	@echo "$(GREEN)Running security scans...$(NC)"
	trivy fs . --format table --exit-code 1
	gosec ./...
	@echo "$(GREEN)Security scans completed!$(NC)"

# Build Commands
build: ## Build all services
	@echo "$(GREEN)Building services...$(NC)"
	cd apps/ingestion-service && go build -o bin/ingestion-service main.go
	cd apps/processing-service && go build -o bin/processing-service main.go
	cd apps/storage-service && go build -o bin/storage-service main.go
	cd apps/alert-service && go build -o bin/alert-service main.go
	@echo "$(GREEN)Build completed!$(NC)"

build-docker: ## Build Docker images
	@echo "$(GREEN)Building Docker images...$(NC)"
	docker build -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)/ingestion-service:$(VERSION) apps/ingestion-service/
	docker build -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)/processing-service:$(VERSION) apps/processing-service/
	docker build -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)/storage-service:$(VERSION) apps/storage-service/
	docker build -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)/alert-service:$(VERSION) apps/alert-service/
	@echo "$(GREEN)Docker images built successfully!$(NC)"

# Load Testing Commands
load-test: ## Run k6 load tests
	@echo "$(GREEN)Running load tests...$(NC)"
	cd apps/ingestion-service && k6 run load-test.js
	@echo "$(GREEN)Load tests completed!$(NC)"

load-test-stress: ## Run stress tests (higher load)
	@echo "$(GREEN)Running stress tests...$(NC)"
	cd apps/ingestion-service && k6 run --env STAGE=stress load-test.js
	@echo "$(GREEN)Stress tests completed!$(NC)"

# Infrastructure Commands
infra-init: ## Initialize Terraform infrastructure
	@echo "$(GREEN)Initializing Terraform...$(NC)"
	cd infra/terraform && terraform init
	@echo "$(GREEN)Terraform initialized!$(NC)"

infra-plan: ## Plan Terraform changes
	@echo "$(GREEN)Planning Terraform changes...$(NC)"
	cd infra/terraform && terraform plan -var-file=staging.tfvars
	@echo "$(GREEN)Terraform plan completed!$(NC)"

infra-apply: ## Apply Terraform changes
	@echo "$(GREEN)Applying Terraform changes...$(NC)"
	cd infra/terraform && terraform apply -var-file=staging.tfvars -auto-approve
	@echo "$(GREEN)Terraform apply completed!$(NC)"

infra-destroy: ## Destroy Terraform infrastructure
	@echo "$(RED)Destroying infrastructure...$(NC)"
	cd infra/terraform && terraform destroy -var-file=staging.tfvars -auto-approve
	@echo "$(YELLOW)Infrastructure destroyed!$(NC)"

# Kubernetes Commands
k8s-deploy: ## Deploy to Kubernetes
	@echo "$(GREEN)Deploying to Kubernetes...$(NC)"
	cd infra/helm/ingestion-service && helm upgrade --install ingestion-service . \
		--namespace staging --create-namespace \
		--set image.tag=$(VERSION)
	@echo "$(GREEN)Deployment completed!$(NC)"

k8s-status: ## Check Kubernetes deployment status
	@echo "$(GREEN)Checking deployment status...$(NC)"
	kubectl get pods -n staging
	kubectl get services -n staging
	kubectl get ingress -n staging

k8s-logs: ## View service logs
	@echo "$(GREEN)Viewing service logs...$(NC)"
	kubectl logs -f deployment/ingestion-service -n staging

# Kong API Gateway Commands
kong-start: ## Start Kong API Gateway
	@echo "$(GREEN)Starting Kong API Gateway...$(NC)"
	cd infra/kong && docker-compose up -d
	@echo "$(GREEN)Kong started! Access at http://localhost:8000$(NC)"

kong-stop: ## Stop Kong API Gateway
	@echo "$(YELLOW)Stopping Kong API Gateway...$(NC)"
	cd infra/kong && docker-compose down

kong-status: ## Check Kong status
	@echo "$(GREEN)Checking Kong status...$(NC)"
	curl -s http://localhost:8001/status | jq .

# Demo Commands
demo: ## Run complete system demo
	@echo "$(GREEN)🚀 Starting Barclays-Grade System Demo$(NC)"
	@echo "$(YELLOW)1. Starting infrastructure...$(NC)"
	$(MAKE) infra-start
	@echo "$(YELLOW)2. Starting services...$(NC)"
	$(MAKE) k8s-deploy
	@echo "$(YELLOW)3. Running load tests...$(NC)"
	$(MAKE) load-test
	@echo "$(YELLOW)4. Showing metrics...$(NC)"
	$(MAKE) show-metrics
	@echo "$(GREEN)✅ Demo completed successfully!$(NC)"

demo-cleanup: ## Clean up demo environment
	@echo "$(YELLOW)Cleaning up demo environment...$(NC)"
	$(MAKE) k8s-cleanup
	$(MAKE) infra-stop
	@echo "$(GREEN)Demo environment cleaned up!$(NC)"

# Utility Commands
show-metrics: ## Show system metrics
	@echo "$(GREEN)System Metrics:$(NC)"
	@echo "$(YELLOW)Health Check:$(NC)"
	curl -s http://localhost:8080/health | jq .
	@echo "$(YELLOW)Prometheus Metrics:$(NC)"
	curl -s http://localhost:9090/metrics | head -20

check-deps: ## Check system dependencies
	@echo "$(GREEN)Checking dependencies...$(NC)"
	@echo "Go: $(GO_VERSION)"
	@echo "Docker: $(shell docker --version)"
	@echo "Kubernetes: $(shell kubectl version --client --short 2>/dev/null || echo 'Not installed')"
	@echo "Helm: $(shell helm version --short 2>/dev/null || echo 'Not installed')"
	@echo "Terraform: $(shell terraform version -json 2>/dev/null | jq -r .terraform_version || echo 'Not installed')"

# Cleanup Commands
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf apps/*/bin/
	rm -rf apps/*/coverage.out
	rm -rf coverage.html
	@echo "$(GREEN)Cleanup completed!$(NC)"

clean-all: ## Clean everything including Docker
	@echo "$(RED)Cleaning everything...$(NC)"
	$(MAKE) clean
	docker system prune -f
	docker volume prune -f
	@echo "$(GREEN)Complete cleanup finished!$(NC)"

# Documentation Commands
docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	cd apps/ingestion-service && godoc -http=:6060 &
	@echo "$(GREEN)Documentation available at http://localhost:6060$(NC)"

# CI/CD Commands
ci-test: ## Run CI pipeline tests
	@echo "$(GREEN)Running CI pipeline tests...$(NC)"
	$(MAKE) install
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) security-scan
	$(MAKE) build
	@echo "$(GREEN)CI tests completed successfully!$(NC)"

# Production Commands
prod-deploy: ## Deploy to production
	@echo "$(RED)🚨 Deploying to PRODUCTION...$(NC)"
	@read -p "Are you sure you want to deploy to production? (y/N): " confirm && [ "$$confirm" = "y" ]
	cd infra/helm/ingestion-service && helm upgrade --install ingestion-service . \
		--namespace production --create-namespace \
		--set image.tag=$(VERSION) \
		--values production-values.yaml
	@echo "$(GREEN)Production deployment completed!$(NC)"

# Helpers
.PHONY: check-env
check-env: ## Check environment variables
	@echo "$(GREEN)Environment Check:$(NC)"
	@echo "AWS_REGION: $(AWS_REGION)"
	@echo "KUBECONFIG: $(KUBECONFIG)"
	@echo "DOCKER_REGISTRY: $(DOCKER_REGISTRY)"

# Default targets for common workflows
all: install test build ## Install, test, and build everything
dev-setup: install dev ## Setup development environment
prod-setup: install infra-init infra-apply k8s-deploy ## Setup production environment

# Show version info
version: ## Show version information
	@echo "$(GREEN)Version Information:$(NC)"
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Docker Registry: $(DOCKER_REGISTRY)"

.PHONY: test coverage race test-e2e build compose-up compose-down compose-logs k8s-apply k8s-delete

# Executa todos os testes unitários
test:
	go test -v ./...

# Executa testes com detector de race conditions
race:
	go test -race -v ./internal/...

# Executa teste ponta a ponta automatizado contra a stack Docker
test-e2e:
	bash scripts/test_e2e.sh

# Calcula cobertura de código com relatório
coverage:
	go test -coverprofile=coverage.out ./internal/domain ./internal/application ./internal/httpapi ./internal/httpapi/handler ./internal/httpapi/middleware ./internal/adapter/memory ./internal/adapter/postgres ./internal/adapter/storage
	go tool cover -func=coverage.out
	@echo "Gerando coverage.html..."
	go tool cover -html=coverage.out -o coverage.html

# Compila binários locais
build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

# Inicia ambiente completo via Docker Compose
compose-up:
	docker compose up --build -d

# Para containers
compose-down:
	docker compose down -v

# Acompanha logs do compose
compose-logs:
	docker compose logs -f

# Aplica manifestos no Kubernetes
k8s-apply:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/config.yaml
	kubectl apply -f k8s/infra.yaml
	kubectl apply -f k8s/apps.yaml

# Remove recursos do Kubernetes
k8s-delete:
	kubectl delete -f k8s/apps.yaml --ignore-not-found
	kubectl delete -f k8s/infra.yaml --ignore-not-found
	kubectl delete -f k8s/config.yaml --ignore-not-found
	kubectl delete -f k8s/namespace.yaml --ignore-not-found

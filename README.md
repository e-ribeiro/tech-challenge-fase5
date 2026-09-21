# FIAP X — Sistema Distribuído de Processamento de Vídeos (Fase 5 - Hackathon)

[![CI](https://github.com/e-ribeiro/tech-challenge-fase5/actions/workflows/ci.yml/badge.svg)](https://github.com/e-ribeiro/tech-challenge-fase5/actions/workflows/ci.yml)
[![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=e-ribeiro_tech-challenge-fase5&metric=alert_status)](https://sonarcloud.io/summary/overall?id=e-ribeiro_tech-challenge-fase5)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=e-ribeiro_tech-challenge-fase5&metric=coverage)](https://sonarcloud.io/summary/overall?id=e-ribeiro_tech-challenge-fase5)

Este projeto é a evolução e refatoração completa da prova de conceito da **FIAP X** para uma **arquitetura de microsserviços orientada a eventos**, escalável, resiliente a picos e protegida por autenticação, atendendo rigorosamente a todos os requisitos do Hackathon da Fase 5 da Pós-Tech.

---

## 1. Atendimento aos Requisitos

### Requisitos Funcionais
- ✅ **Processamento Simultâneo de Vídeos**: Arquitetura assíncrona desacoplada em que múltiplos workers processam vídeos em paralelo.
- ✅ **Resiliência a Picos (Zero Data Loss)**: Buffer de mensageria com RabbitMQ (filas duráveis, confirmação manual `ack`, Dead-Letter Queue `DLQ` e retries). A API aceita requisições imediatamente (`202 Accepted`) sem perda mesmo sob carga extrema.
- ✅ **Proteção por Usuário e Senha**: Autenticação com senhas criptografadas via `bcrypt` e emissão de tokens seguros `JWT`.
- ✅ **Listagem e Status por Usuário**: O usuário autenticado visualiza apenas os seus vídeos com status em tempo real (`PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`), data de envio, contagem de frames e botão de download do `.ZIP`.
- ✅ **Notificação em Caso de Erro**: Se um vídeo corrompido for enviado, o worker identifica a falha, atualiza o status para `FAILED` com a causa do erro e dispara um e-mail com detalhes para o endereço do usuário (verificável via MailHog).

### Requisitos Técnicos
- ✅ **Persistência de Dados**: **PostgreSQL 16** com migrations versionadas + **MinIO / AWS S3** para armazenamento de objetos (vídeos e zips).
- ✅ **Arquitetura Escalável (Hexagonal / Clean Architecture)**: Padrão Ports & Adapters em Go, com clara separação entre domínio, aplicação e adaptadores de infraestrutura.
- ✅ **Testes Automatizados e Qualidade**: Cobertura de código **superior a 80%** (`83.8%` nos pacotes core), testes unitários com `go-sqlmock`, testes de integração e verificações de concorrência com detector `-race`.
- ✅ **CI/CD Automatizado**: Workflow do **GitHub Actions** validando formatação, `go vet`, testes, bloqueio de cobertura < 80% e compilação de imagens Docker.
- ✅ **Containers e Orquestração**: Dockerfiles multi-stage otimizados, ambiente unificado no `docker-compose.yml` e manifestos **Kubernetes com HPA** (Horizontal Pod Autoscaler) escalando de 2 a 10 réplicas.

---

## 2. Desenho de Arquitetura

```mermaid
flowchart LR
    Client[Investidor / Frontend] -->|REST + JWT| API[fiapx-video-api]
    API -->|Grava Vídeo| S3[MinIO S3 Storage]
    API -->|Jobs & Usuários| DB[(PostgreSQL 16)]
    API -->|Publica video.process| Broker[RabbitMQ Broker + DLQ]

    Broker -->|Consome Tarefa| Worker[fiapx-video-worker\n(FFmpeg + HPA)]
    Worker -->|Baixa Vídeo & Sobe ZIP| S3
    Worker -->|Atualiza Status| DB
    Worker -->|Emite video.failed / completed| Broker

    Broker -->|Consome Notificação| Notifier[Notification Service]
    Notifier -->|Envia E-mail de Erro| MailHog[Servidor SMTP / MailHog]
```

Detalhes completos e ADRs: [docs/architecture.md](docs/architecture.md).

---

## 3. Execução Rápida (1 Comando)

### Pré-requisito: Docker Compose
Para subir todo o ecossistema localmente (API, Worker, PostgreSQL, RabbitMQ, MinIO e MailHog):

```bash
docker compose up --build
```

### URLs de Acesso:
- **Interface Web do FIAP X**: [http://localhost:8080](http://localhost:8080)
- **API Health Check**: [http://localhost:8080/health/live](http://localhost:8080/health/live)
- **Painel RabbitMQ**: [http://localhost:15672](http://localhost:15672) (usuário: `guest`, senha: `guest`)
- **Console MinIO S3**: [http://localhost:9001](http://localhost:9001) (usuário: `minioadmin`, senha: `minioadmin`)
- **Webmail MailHog (Notificações de Erro)**: [http://localhost:8025](http://localhost:8025)

---

## 4. Testes e Cobertura de Código

Execução dos testes com checagem de concorrência e geração de relatório de cobertura:

```bash
make test
make race
make coverage
```

O relatório HTML de cobertura é gerado em `coverage.html`. A cobertura atual dos pacotes centrais de domínio, aplicação e adaptadores atinge **83.8%**, cumprindo com segurança o requisito da banca.

---

## 5. Kubernetes & Escalabilidade Horizontal (HPA)

Os manifestos Kubernetes estão organizados em `k8s/`:

```bash
# Aplicar todos os recursos no cluster (API, Worker, Postgres, RabbitMQ, MinIO e HPAs):
make k8s-apply
```

O HPA do Worker monitora o uso de CPU e dimensiona automaticamente as réplicas entre **2 e 10 pods** conforme a demanda de extração de frames cresce.

---

## 6. Contratos de API

A documentação OpenAPI 3.0 completa está disponível em [`api/openapi.yaml`](api/openapi.yaml).

| Método | Endpoint | Protegido? | Descrição |
|---|---|:---:|---|
| `POST` | `/api/v1/auth/register` | Não | Cadastra novo investidor/usuário |
| `POST` | `/api/v1/auth/login` | Não | Autentica e retorna token JWT |
| `POST` | `/api/v1/videos/upload` | Sim (JWT) | Envia vídeo e enfileira para processamento |
| `GET` | `/api/v1/videos` | Sim (JWT) | Lista o status dos vídeos do usuário logado |
| `GET` | `/api/v1/videos/:id` | Sim (JWT) | Consulta detalhes e status de um vídeo |
| `GET` | `/api/v1/videos/:id/download` | Sim (JWT) | Realiza download do arquivo `.ZIP` com os frames |
| `GET` | `/health/live` | Não | Liveness probe do Kubernetes |
| `GET` | `/health/ready` | Não | Readiness probe do Kubernetes |

---

## 7. Roteiro da Apresentação em Vídeo (Máx. 10 Minutos)

Para a gravação do vídeo de entrega do Hackathon:

1. **Minutos 0:00 - 2:00 | Introdução e Contexto**:
   - Apresentação do desafio da FIAP X.
   - Demonstração do código legado (`cmd/legacy/main.go`) e explicação de por que era inviável (síncrono, sem fila, sem banco, sem autenticação, perda de dados sob carga).
2. **Minutos 2:00 - 4:00 | Arquitetura Escolhida e Boas Práticas**:
   - Apresentação do diagrama C4 e do padrão Hexagonal (Ports & Adapters) em Go.
   - Explicação do desacoplamento da ingestão (API) e do processamento (Worker escalável via RabbitMQ com confirmação manual `ack`).
   - Armazenamento desacoplado em MinIO S3 e PostgreSQL.
3. **Minutos 4:00 - 7:30 | Demonstração Prática do Sistema Funcionando**:
   - Acesso à interface web ([http://localhost:8080](http://localhost:8080)).
   - Criação de conta de usuário e login com obtenção de JWT.
   - Envio de múltiplos vídeos simultâneos comprovando concorrência e upload assíncrono imediato (`202 Accepted`).
   - Acompanhamento do status mudando de `PENDING` para `PROCESSING` e `COMPLETED`.
   - Download e descompactação do arquivo `.ZIP` gerado com as imagens extraídas a 1 fps.
   - **Simulação de Erro e Notificação**: Envio de arquivo corrompido / inválido, exibição do status `FAILED` e abertura do MailHog ([http://localhost:8025](http://localhost:8025)) demonstrando o e-mail de alerta disparado para o investidor.
4. **Minutos 7:30 - 9:00 | Qualidade de Software, CI/CD e Kubernetes**:
   - Demonstração dos testes unitários passando com `make coverage` (> 80%).
   - Demonstração do pipeline no GitHub Actions aprovando o quality gate.
   - Apresentação dos manifestos Kubernetes e do HPA configurado para autoscaling sob picos.
5. **Minutos 9:00 - 10:00 | Conclusão**:
   - Fechamento reforçando o atendimento integral a todos os requisitos funcionais e técnicos da FIAP X.

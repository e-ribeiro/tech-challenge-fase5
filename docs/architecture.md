# Documentação de Arquitetura — FIAP X Video Processing System

## 1. Visão Geral

Este documento descreve a arquitetura distribuída desenvolvida para a **FIAP X** para substituir a prova de conceito monolítica e síncrona original por um sistema baseado em microsserviços desacoplados, resiliente a picos de tráfego, escalável horizontalmente e seguro.

---

## 2. Diagrama de Arquitetura C4 (Nível Container)

```mermaid
flowchart TB
    subgraph Users["Atores"]
        Client["Investidor / Usuário Final\n(Browser / API Client)"]
    end

    subgraph FIAPX_Cluster["Cluster Kubernetes / Docker Compose"]
        subgraph IngressLayer["Borda e Ingress"]
            Ingress["Ingress Controller / API Gateway\n[Porta 8080]"]
        end

        subgraph CoreServices["Microsserviços"]
            API["fiapx-video-api\n(Auth, Ingestão, Status, Download)\n[Golang Gin]"]
            Worker["fiapx-video-worker\n(Consumidor AMQP, FFmpeg, ZIP)\n[Golang + FFmpeg (HPA 2-10 pods)]"]
            Notification["fiapx-notification-service\n(Consumidor de eventos e despachante)\n[Golang Worker]"]
        end

        subgraph Messaging["Broker de Mensageria"]
            RabbitMQ["RabbitMQ Broker\n(video.process.queue + DLQ)\n(video.notification.queue)"]
        end

        subgraph Persistence["Camada de Dados & Objetos"]
            PostgreSQL[("PostgreSQL 16\n(Tabelas: users, video_jobs)")]
            MinIO[("MinIO S3 Object Storage\n(Buckets: fiapx-videos)")]
        end

        subgraph ExternalMock["Serviços de Comunicação"]
            MailHog["MailHog SMTP Mock\n[Portas 1025 / 8025]"]
        end
    end

    Client -->|HTTPS REST| Ingress
    Ingress --> API
    API -->|Validação & Hash| PostgreSQL
    API -->|Grava Vídeo Bruto| MinIO
    API -->|Publica Job (ack)| RabbitMQ

    RabbitMQ -->|Consome Tarefas (prefetch 1)| Worker
    Worker -->|Baixa Vídeo| MinIO
    Worker -->|Processa 1 fps & Compacta| Worker
    Worker -->|Salva ZIP de Frames| MinIO
    Worker -->|Atualiza Status COMPLETED/FAILED| PostgreSQL
    Worker -->|Publica video.failed / video.completed| RabbitMQ

    RabbitMQ -->|Consome Notificações| Notification
    Notification -->|Dispara E-mail com Erro| MailHog
```

---

## 3. Registro de Decisões Arquiteturais (ADRs)

### ADR 001: Arquitetura Orientada a Eventos e Desacoplamento Assíncrono
- **Contexto**: A POC original realizava a conversão do vídeo durante o request HTTP. Vídeos longos causavam congelamento do servidor, concorrência nula e perda de requisições sob picos.
- **Decisão**: A API responde imediatamente `202 Accepted` ao receber o vídeo, salvando o arquivo bruto no Object Storage e enviando uma mensagem para a fila `video.process.queue` do RabbitMQ. Os workers consomem de forma assíncrona.
- **Consequências**: Zero perda de requisições durante picos. A fila absorve qualquer sobrecarga sem degradar a API.

### ADR 002: RabbitMQ com Confirmação Manual (Ack) e Dead-Letter Queue (DLQ)
- **Contexto**: Se um pod de processamento falhar durante a extração de frames (falha de hardware, OOM ou reinicialização), o trabalho não pode ser perdido. Se o vídeo estiver corrompido, não pode travar o loop de consumo.
- **Decisão**: Consumo configurado com `auto-ack = false`. A confirmação (`Ack`) só é emitida após o arquivo ZIP ser persistido no MinIO e gravado no PostgreSQL. Falhas irrecuperáveis são enviadas para a Dead-Letter Queue (`video.process.dlq`).
- **Consequências**: Garantia de entrega *at-least-once* e isolamento de mensagens venenosas.

### ADR 003: Armazenamento em Object Storage (MinIO / S3) em Vez de Disco Efêmero
- **Contexto**: Em ambientes orquestrados (Kubernetes), os pods são efêmeros e descartáveis. Gravar vídeos e zips no sistema de arquivos local do container gera perda de dados ao reiniciar ou escalar pods.
- **Decisão**: Uso da API S3 via MinIO no ambiente local e AWS S3 em produção. O pod apenas mantém diretórios temporários na memória/disco durante o processamento do job em andamento.
- **Consequências**: Qualquer réplica de API pode atender o download do ZIP processado por qualquer réplica de Worker.

### ADR 004: Autenticação JWT com bcrypt e Isolamento Multi-inquilino
- **Contexto**: O sistema base permitia que qualquer usuário listasse e baixasse qualquer vídeo.
- **Decisão**: Implementação de autenticação com credenciais (e-mail e senha com hash bcrypt). Cada job é vinculado ao `user_id`. Endpoints de listagem e download validam estritamente o proprietário do recurso.
- **Consequências**: Segurança em conformidade com os requisitos dos investidores.

---

## 4. Dicionário de Status do Job de Vídeo

| Status | Descrição |
|---|---|
| `PENDING` | Vídeo recebido pela API, salvo no storage e enfileirado na mensageria. |
| `PROCESSING` | Worker iniciou a tarefa, baixou o vídeo e está executando o FFmpeg. |
| `COMPLETED` | Frames extraídos com sucesso, arquivo `.zip` gerado e disponível para download. |
| `FAILED` | Falha na decodificação ou conversão. Mensagem de erro registrada e notificação enviada por e-mail ao usuário. |

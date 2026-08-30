# postech-tc3-lambda-auth

**Function serverless de autenticação por CPF** do sistema de gestão de oficinas mecânicas.

Tech Challenge Fase 3 — FIAP Pós Tech SOAT.

## Propósito

Recebe um CPF, valida o número, consulta a existência e o status do cliente no banco gerenciado e devolve um **JWT** para consumo das APIs protegidas atrás do API Gateway.

## Tecnologias

- Go 1.23 rodando em AWS Lambda (runtime `provided.al2023`, arm64)
- `golang-jwt/v5` para emissão do token
- `lib/pq` para acesso ao PostgreSQL gerenciado
- Logs estruturados em JSON via `log/slog`
- CI/CD: GitHub Actions com OIDC

## Arquitetura

```
  POST /auth  { "cpf": "529.982.247-25" }
        │
        ▼
 ┌──────────────┐
 │ API Gateway  │  (postech-tc3-infra-k8s)
 └──────┬───────┘
        ▼
 ┌───────────────────────────────────────────┐
 │ Lambda auth                               │
 │  1. cpf.Validate     → 400 se inválido    │
 │  2. FindByDocument   → 404 / 403          │
 │  3. token.Issue      → 200 { JWT }        │
 └──────────────────┬────────────────────────┘
                    ▼
              RDS PostgreSQL
          (postech-tc3-infra-database)
```

## Estrutura

```
cmd/lambda/main.go        entrypoint, orquestra validação → consulta → token
internal/cpf/             validação de CPF (dígitos verificadores)
internal/customer/        consulta do cliente e checagem de status
internal/token/           emissão do JWT
```

## Respostas

| Situação | HTTP | Corpo |
|---|---|---|
| CPF válido e cliente ativo | 200 | `access_token`, `token_type`, `expires_at` |
| Payload malformado | 400 | `payload invalido` |
| CPF inválido | 400 | `cpf invalido` |
| Cliente inexistente | 404 | `cliente nao encontrado` |
| Cliente inativo | 403 | `cliente inativo` |
| Falha interna | 500 | `erro interno` |

## Execução local

```bash
cp .env.example .env
make tidy
make test
```

Empacotar o artefato de deploy:

```bash
make package    # gera function.zip
```

## Variáveis de ambiente

| Variável | Descrição |
|---|---|
| `DATABASE_URL` | string de conexão do RDS (vinda do Secrets Manager) |
| `JWT_SECRET` | segredo de assinatura HS256 |
| `JWT_TTL` | validade do token (padrão `15m`) |

## Deploy

| Evento | Ação |
|---|---|
| Pull Request | `go mod tidy`, `go vet`, testes com race e cobertura |
| Push em `homolog` | build + `update-function-code` em staging |
| Push em `main` | build + `update-function-code` em produção |

Secret necessário: `AWS_ROLE_ARN`.

## Pendências

- `go.sum` ainda não versionado — rode `make tidy` e commite o arquivo gerado.
- Nome da tabela/colunas em `internal/customer` precisa ser conferido contra o schema real do `postech-tc3-app`.

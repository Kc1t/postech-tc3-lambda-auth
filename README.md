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
internal/requester/       consulta em `requesters` e checagem de status
internal/token/           emissão do JWT
infra/                    Terraform da function (role, VPC, log group, código)
```

## Contrato do token

O JWT é validado pelo middleware `Auth` do [`postech-tc3-app`](https://github.com/Kc1t/postech-tc3-app), que exige `sub` e `role` não vazios. As claims emitidas são:

| Claim | Valor |
|---|---|
| `sub` | id do solicitante em `requesters` |
| `role` | `client` — vocabulário do app (`admin` \| `client`) |
| `email`, `name`, `document` | dados do solicitante |
| `iat`, `exp` | emissão e expiração (`JWT_TTL`) |

Assinatura HS256 com o **mesmo** `JWT_SECRET` da aplicação. Se os segredos divergirem, o app rejeita o token com 401.

## Respostas

| Situação | HTTP | Corpo |
|---|---|---|
| CPF válido e cliente ativo | 200 | `access_token`, `token_type`, `expires_at` |
| Payload malformado | 400 | `payload invalido` |
| CPF inválido | 400 | `cpf invalido` |
| Cliente inexistente | 404 | `cliente nao encontrado` |
| Cliente inativo | 403 | `cliente inativo` |
| Falha interna | 500 | `erro interno` |

O `x-correlation-id` do request é propagado para os logs; na ausência dele, usa o request id do API Gateway.

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
| Pull Request | tidy, `gofmt`, `go vet`, testes com race e cobertura, `terraform validate` |
| Push em `homolog` | `make package` + `terraform apply` em staging |
| Push em `main` | `make package` + `terraform apply` em produção |

O Terraform é dono do código da function: o `source_code_hash` do `function.zip` dispara a atualização no `apply`, sem passo separado de `update-function-code`.

Secrets necessários: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, `TF_STATE_BUCKET`, `POSTGRES_DSN`, `JWT_SECRET`.

## Infraestrutura

A function roda **dentro da VPC** para alcançar o RDS, que não é publicamente acessível. Como só precisa falar com o banco, não há NAT Gateway: o `DATABASE_URL` chega por variável de ambiente em vez de leitura do Secrets Manager em runtime, o que evita saída para a internet e os $32,85/mês do NAT.

A execution role é a `LabRole` existente (`var.lab_role_name`), e não uma role criada pelo Terraform — o AWS Academy Learner Lab não permite criar roles IAM.

Após o primeiro apply, leve o output `invoke_arn` para a variável `lambda_authorizer_invoke_arn` do [`postech-tc3-infra-k8s`](https://github.com/Kc1t/postech-tc3-infra-k8s) para ligar o authorizer no API Gateway.

## Pendências

- **`go.sum` ainda não versionado.** Rode `make tidy` e commite o arquivo gerado — sem ele o job de teste falha no `git diff --exit-code`.
- `vpc_id` está com placeholder em `infra/envs/*.tfvars`.

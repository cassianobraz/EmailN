# EmailN

EmailN e uma API em Go para criacao, gerenciamento e envio de campanhas de e-mail. O projeto combina uma API HTTP protegida por JWT, autenticacao centralizada no Keycloak, persistencia em PostgreSQL via GORM e envio de mensagens por SMTP.

A proposta e manter o fluxo de campanhas simples e seguro: somente usuarios autenticados conseguem acessar as rotas, o e-mail do usuario autenticado e extraido do token JWT e usado como autor da campanha, e o envio pode ser disparado pela API ou processado por um worker.

## Principais recursos

- API REST escrita em Go.
- Rotas HTTP com `go-chi`.
- Autenticacao OIDC/JWT integrada ao Keycloak.
- Validacao de token por `coreos/go-oidc`.
- Extracao de claims do JWT para identificar o usuario autenticado.
- Persistencia de campanhas e contatos em PostgreSQL.
- Mapeamento e migracao automatica com GORM.
- Envio de e-mails via SMTP usando `gomail`.
- Worker independente para reprocessar campanhas iniciadas.
- Validacao de entrada com `go-playground/validator`.
- Ambiente local com Docker Compose para PostgreSQL e Keycloak.

## Stack

- Go `1.26.2`
- PostgreSQL `17`
- Keycloak `21.1.1`
- GORM
- go-chi
- go-oidc
- gomail
- Docker Compose

## Arquitetura

O projeto segue uma separacao simples por responsabilidade:

```text
cmd/
  api/                 Entrada da API HTTP
  worker/              Processo em loop para envio/reprocessamento
  initializers/        Carregamento de variaveis de ambiente

internal/
  contract/            DTOs de entrada e saida
  domain/campaign/     Entidade, regras de negocio e contrato do repositorio
  endpoints/           Handlers HTTP, middleware de auth e tratamento de erro
  infrastructure/
    database/          Conexao PostgreSQL e implementacao GORM do repositorio
    mail/              Integracao SMTP para envio de e-mail
  internal-errors/     Erros padronizados e validacao
```

### Fluxo principal

1. O usuario autentica no Keycloak e recebe um `access_token`.
2. A API recebe requisicoes com `Authorization: Bearer <token>`.
3. O middleware `Auth` valida o token no provider OIDC configurado em `URLKEYCLOAK`.
4. A claim `email` do JWT e colocada no contexto da requisicao.
5. Ao criar uma campanha, esse e-mail e salvo como `CreatedBy`.
6. A campanha nasce com status `Pending`.
7. Ao iniciar a campanha, a API marca como `Started` e dispara o envio em uma goroutine.
8. O envio SMTP atualiza a campanha para `Done` em caso de sucesso ou `Fail` em caso de erro.
9. O worker tambem busca campanhas `Started` paradas ha pelo menos 1 minuto e tenta processa-las.

## Modelo de dominio

### Campaign

Uma campanha representa uma mensagem enviada para uma lista de contatos.

Campos principais:

- `ID`: identificador gerado com `xid`.
- `Name`: nome/assunto da campanha.
- `Content`: corpo HTML do e-mail.
- `Contacts`: lista de destinatarios.
- `Status`: estado atual da campanha.
- `CreatedBy`: e-mail do usuario autenticado que criou a campanha.
- `CreatedOn` e `UpdatedOn`: controle temporal.

### Contact

Cada contato armazena o e-mail de destino associado a uma campanha.

### Status possiveis

- `Pending`: campanha criada e ainda nao enviada.
- `Started`: campanha em processamento.
- `Done`: envio concluido com sucesso.
- `Fail`: erro no envio.
- `Canceled`: campanha cancelada.
- `Deleted`: campanha removida logicamente no dominio.

## Regras de negocio

- `Name` deve ter entre 5 e 24 caracteres.
- `Content` deve ter entre 5 e 1024 caracteres.
- A campanha deve ter pelo menos 1 contato.
- Cada contato deve ter um e-mail valido.
- `CreatedBy` deve ser um e-mail valido vindo do token JWT.
- Apenas campanhas com status `Pending` podem ser iniciadas, canceladas ou removidas.
- O envio atualiza o status para `Done` ou `Fail`.

## Variaveis de ambiente

Crie um arquivo `.env` na raiz do projeto.

```env
DATABASE=host=localhost user=emailn_dev password=d4#rt6 dbname=emailn_dev port=5432 sslmode=disable
URLKEYCLOAK=http://localhost:8080/realms/<realm>

EMAIL_SMTP=smtp.seuprovedor.com
EMAIL_USER=seu-email@dominio.com
EMAIL_PASWORD=sua-senha-ou-app-password
```

Observacao: a variavel `EMAIL_PASWORD` esta com esse nome no codigo atual. Se ela for renomeada para `EMAIL_PASSWORD`, tambem sera necessario ajustar `internal/infrastructure/mail/send_mail.go`.

## Subindo dependencias locais

O `docker-compose.yaml` sobe PostgreSQL e Keycloak:

```bash
docker compose up -d
```

Servicos expostos:

- PostgreSQL: `localhost:5432`
- Keycloak: `http://localhost:8080`

Credenciais padrao do Keycloak no ambiente local:

- Usuario: `admin`
- Senha: `admin`

No Keycloak, configure um realm, um client com o identificador `emailn` e usuarios com a claim `email`, pois a API valida o token usando esse client e depende dessa claim para preencher o criador da campanha.

## Executando a API

```bash
go run ./cmd/api
```

A API sobe em:

```text
http://localhost:3000
```

## Executando o worker

```bash
go run ./cmd/worker
```

O worker executa continuamente a cada 10 segundos. Ele busca campanhas com status `Started` que estejam nessa situacao ha pelo menos 1 minuto e tenta concluir o envio.

## Autenticacao e permissoes

Todas as rotas registradas em `/campaigns` usam o middleware `Auth`.

O fluxo de seguranca e:

1. Receber o header `Authorization`.
2. Remover o prefixo `Bearer`.
3. Carregar o provider OIDC usando `URLKEYCLOAK`.
4. Validar o token para o client `emailn`.
5. Rejeitar a requisicao com `401 Unauthorized` quando o token estiver ausente ou invalido.
6. Ler a claim `email` do JWT.
7. Armazenar o e-mail no contexto da requisicao.

Exemplo de header:

```http
Authorization: Bearer eyJhbGciOi...
```

## Rotas da API

Base URL:

```text
http://localhost:3000/campaigns
```

### Criar campanha

```http
POST /campaigns/
Authorization: Bearer <access_token>
Content-Type: application/json
```

Body:

```json
{
  "name": "Oferta julho",
  "content": "<strong>Campanha promocional</strong>",
  "emails": ["cliente@dominio.com"]
}
```

Resposta `201 Created`:

```json
{
  "id": "campaign_id"
}
```

### Buscar campanha por ID

```http
GET /campaigns/{id}
Authorization: Bearer <access_token>
```

Resposta `200 OK`:

```json
{
  "ID": "campaign_id",
  "Name": "Oferta julho",
  "Content": "<strong>Campanha promocional</strong>",
  "Status": "Pending",
  "AmountOfEmailsToSend": 1,
  "CreatedBy": ""
}
```

### Iniciar envio

```http
PATCH /campaigns/start/{id}
Authorization: Bearer <access_token>
```

Resposta esperada:

```text
200 OK
```

### Remover campanha

```http
DELETE /campaigns/delete/{id}
Authorization: Bearer <access_token>
```

Resposta esperada no codigo atual:

```text
404 Not Found
```

Observacao: o handler remove a campanha, mas retorna `http.StatusNotFound`. Para uma API mais consistente, o ideal seria retornar `204 No Content` quando a remocao ocorrer com sucesso.

## Tratamento de erros

Os handlers retornam erros por meio de `HandlerError`, que centraliza a resposta HTTP:

- `500 Internal Server Error` para erros internos.
- `404 Not Found` quando o registro nao e encontrado.
- `400 Bad Request` para erros de validacao ou regra de negocio.
- `401 Unauthorized` para token ausente ou invalido no middleware de autenticacao.

Formato comum:

```json
{
  "error": "mensagem do erro"
}
```

## Envio de e-mail

O envio fica em `internal/infrastructure/mail/send_mail.go`.

Ele usa:

- `EMAIL_SMTP` como host SMTP.
- Porta `587`.
- `EMAIL_USER` como remetente e usuario de autenticacao.
- `EMAIL_PASWORD` como senha.
- `campaign.Name` como assunto.
- `campaign.Content` como corpo HTML.
- `campaign.Contacts` como destinatarios.

## Banco de dados

A conexao e criada em `internal/infrastructure/database/new_db.go`.

O projeto usa a variavel `DATABASE` como DSN do PostgreSQL e executa `AutoMigrate` para:

- `Campaign`
- `Contact`

## Exemplos HTTP

O arquivo `server.http` contem exemplos de chamadas para:

- Login no Keycloak.
- Criacao de campanha.
- Busca por ID.
- Cancelamento.
- Remocao.
- Inicio de envio.

Alguns exemplos do arquivo precisam ser alinhados com as rotas registradas atualmente na API. Por exemplo, no codigo a rota de inicio de envio e `PATCH /campaigns/start/{id}`.

## Pontos de melhoria tecnica

Este projeto ja possui uma base funcional, mas alguns ajustes deixariam a implementacao mais robusta:

- Registrar a rota de cancelamento (`CampaignCancelPatch`) no router da API.
- Ajustar o retorno de `CampaignDelete` para `204 No Content` em caso de sucesso.
- Corrigir `EMAIL_PASWORD` para `EMAIL_PASSWORD`.
- Validar erro de `render.DecodeJSON` ao receber payloads invalidos.
- Evitar criar o provider OIDC a cada request; ele pode ser inicializado uma vez e reutilizado.
- Tratar com mais seguranca o parse de claims do JWT antes do type assertion.
- Adicionar testes unitarios para dominio e service.
- Adicionar testes de integracao para endpoints protegidos por JWT.
- Remover ou mover o `main.go` da raiz, que hoje parece ser um arquivo auxiliar de teste de validacao.

## Comandos uteis

```bash
go mod tidy
go run ./cmd/api
go run ./cmd/worker
go test ./...
docker compose up -d
docker compose down
```

## Resumo

EmailN e um servico backend em Go para campanhas de e-mail com autenticacao JWT via Keycloak, persistencia em PostgreSQL e envio SMTP. A estrutura separa dominio, endpoints e infraestrutura, o que facilita evoluir o projeto com novas regras, testes, provedores de e-mail e melhorias de seguranca.

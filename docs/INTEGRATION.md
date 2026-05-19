# Guia de Integração — AkiraWhats API

API REST para criar e gerenciar instâncias do WhatsApp, enviar mensagens e receber eventos via webhook.

- **Base URL (dev):** `http://localhost:8080`
- **Auth:** JWT Bearer no header `Authorization: Bearer <token>` (validade 7 dias)
- **OpenAPI:** disponível em `GET /docs` (UI) e [`docs/openapi.yaml`](openapi.yaml)

---

## 1. Autenticação

### Registrar usuário

```http
POST /api/auth/register
Content-Type: application/json

{
  "first_name": "João",
  "last_name": "Silva",
  "email": "joao@exemplo.com",
  "password": "minhasenha123"
}
```

**Respostas:**
- `201 Created` → `{ "token": "...", "user": { ... } }`
- `409 Conflict` → e-mail já cadastrado
- `429 Too Many Requests` → rate-limit (3 burst, 5/min por IP)

### Login

```http
POST /api/auth/login
Content-Type: application/json

{ "email": "joao@exemplo.com", "password": "minhasenha123" }
```

**Respostas:**
- `200 OK` → `{ "token": "...", "user": { ... } }`
- `401 Unauthorized` → credenciais inválidas
- `429 Too Many Requests` → rate-limit (5 burst, 10/min por IP)

Guarde o `token` e envie em todas as chamadas subsequentes:

```http
Authorization: Bearer <token>
```

---

## 2. Fluxo completo de uma instância

```
┌──────────┐   POST /api/instance          ┌──────────────┐
│  Cliente │ ───────────────────────────►  │  AkiraWhats  │
└──────────┘   { id, webhookUrl? }         └──────────────┘
      │                                            │
      │   status: "qr"                             │
      │ ◄──────────────────────────────────────────┤
      │                                            │
      │   GET /api/instance/:id/qr                 │
      │ ──────────────────────────────────────────►│
      │   { qr: "2@..." }                          │
      │ ◄──────────────────────────────────────────┤
      │                                            │
      │   [usuário escaneia com WhatsApp]          │
      │                                            │
      │   GET /api/instance/:id                    │
      │ ──────────────────────────────────────────►│
      │   { status: "connected", phone: "55..." }  │
      │ ◄──────────────────────────────────────────┤
      │                                            │
      │   POST /api/instance/:id/send/text         │
      │ ──────────────────────────────────────────►│
```

### 2.1 Criar instância

```http
POST /api/instance
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "minha-instancia-1",
  "webhookUrl": "https://meuapp.com/webhooks/whatsapp"   // opcional
}
```

`webhookUrl` pode ser configurado depois — veja [§5](#5-webhooks).

**Resposta:** `201 Created`
```json
{
  "id": "minha-instancia-1",
  "status": "qr",
  "qr": "2@AbCd...",
  "webhookUrl": "https://meuapp.com/webhooks/whatsapp"
}
```

Conflito: `409 Conflict` se o `id` já existir.

### 2.2 Obter QR code

Polling a cada 2s até `status` virar `connected`:

```http
GET /api/instance/:id/qr
```

- `200 OK` → `{ "qr": "2@..." }` (string para renderizar como QR)
- `204 No Content` → não há QR (já conectado ou expirado)

> 💡 Renderize a string como QR code (ex.: `qrcode.react`).

### 2.3 Consultar status

```http
GET /api/instance/:id
```

```json
{
  "id": "minha-instancia-1",
  "status": "connected",
  "qr": "",
  "phone": "5511999998888",
  "webhookUrl": "https://meuapp.com/webhooks/whatsapp"
}
```

**Status possíveis:**

| Status | Significado |
|---|---|
| `connecting` | Estabelecendo conexão |
| `qr` | Aguardando escaneamento do QR |
| `connected` | Conectado e autenticado |
| `disconnected` | Caiu — reconexão automática em andamento |
| `logged_out` | Sessão encerrada (precisa novo QR) |

### 2.4 Listar instâncias do usuário

```http
GET /api/instance
```

Retorna apenas as instâncias do dono do token.

### 2.5 Deletar instância

```http
DELETE /api/instance/:id
```

Faz logout no WhatsApp e remove a sessão persistida.

---

## 3. Envio de mensagens

### 3.1 Texto

```http
POST /api/instance/:id/send/text
Authorization: Bearer <token>
Content-Type: application/json

{
  "to": "5511999998888",
  "message": "Olá!"
}
```

Formato de `to`:
- **Contato:** `5511999998888` (DDI + DDD + número, só dígitos) — a API completa com `@s.whatsapp.net`
- **Grupo:** `123456789-987654321@g.us` (JID completo, obtido em `GET /instance/:id/groups`)

**Resposta:** `200 OK`
```json
{ "id": "BAE5...", "timestamp": "2026-05-19T13:45:22Z" }
```

Erros:
- `503 Service Unavailable` → instância não está `connected`
- `400 Bad Request` → JID inválido

### 3.2 Imagem

```http
POST /api/instance/:id/send/image
Authorization: Bearer <token>
Content-Type: multipart/form-data

to=5511999998888
caption=Veja isso     (opcional)
file=@foto.jpg
```

- Limite: **5 MB**
- Content-Type detectado do upload (default: `image/jpeg`)

**Resposta:** `200 OK` (mesmo formato do texto).

### 3.3 Listar grupos

```http
GET /api/instance/:id/groups
```

```json
[
  { "jid": "123456-789@g.us", "name": "Família" },
  { "jid": "987654-321@g.us", "name": "Trabalho" }
]
```

### 3.4 Histórico de mensagens

Últimas 50 mensagens recebidas/enviadas pela instância (persistidas no Mongo):

```http
GET /api/instance/:id/messages
```

```json
[
  {
    "instance_id": "minha-instancia-1",
    "from": "5511999998888@s.whatsapp.net",
    "body": "Olá",
    "timestamp": "2026-05-19T13:30:00Z",
    "status": "received"
  },
  {
    "instance_id": "minha-instancia-1",
    "msg_id": "BAE5...",
    "to": "5511999998888",
    "body": "Oi!",
    "timestamp": "2026-05-19T13:30:05Z",
    "status": "read"
  }
]
```

`status` de mensagens enviadas: `sent` → `delivered` → `read`.

---

## 5. Webhooks

### 5.1 Configurar

Na criação da instância (campo `webhookUrl`) ou via:

```http
POST /api/instance/:id/webhook
Authorization: Bearer <token>
Content-Type: application/json

{ "url": "https://meuapp.com/webhooks/whatsapp" }
```

Enviar `{ "url": "" }` desativa.

### 5.2 Payload

A AkiraWhats faz `POST` para sua URL **em cada mensagem recebida**:

```http
POST https://meuapp.com/webhooks/whatsapp
Content-Type: application/json

{
  "instance": "minha-instancia-1",
  "from": "5511999998888@s.whatsapp.net",
  "message": "Olá!",
  "timestamp": "2026-05-19T13:45:22Z"
}
```

### 5.3 Garantias e limites

- **Retry:** 3 tentativas com backoff exponencial (1s, 2s, 4s). Timeout 10s por tentativa.
- 5xx → retentado. 4xx → considerado entregue (sem retry).
- Seu endpoint **deve responder em até 10s** com 2xx, ou será tratado como falha.
- **At-least-once:** em raras condições (retry após timeout que chegou ao destino) você pode receber a mesma mensagem mais de uma vez — deduplique pelo par `(instance, timestamp, from, message)` ou implemente sua própria chave de idempotência.

### 5.4 Limitações conhecidas

- O webhook **só dispara para mensagens recebidas**. Não há webhook para recibos de entrega/leitura, desconexões da instância, ou mensagens enviadas — para esses, consulte `GET /instance/:id/messages` ou `GET /instance/:id`.
- O campo `message` traz apenas o texto simples (`Conversation`). Mensagens estendidas (com formatação, citação ou links) podem chegar com `message` vazio — nesses casos consulte `GET /instance/:id/messages` para o conteúdo persistido.
- **Não há assinatura HMAC** ainda — não há como o consumidor verificar criptograficamente que a request veio da AkiraWhats. Mitigação temporária: use uma URL com path secreto (ex.: `/webhooks/whatsapp/<token-aleatorio>`) e valide-o do lado consumidor.

---

## 6. Gerenciamento da conta

| Método | Path | Descrição |
|---|---|---|
| `GET` | `/api/user/me` | Dados do usuário autenticado |
| `PUT` | `/api/user/me` | Atualizar nome/email |
| `PUT` | `/api/user/me/password` | Trocar senha (`current_password`, `new_password`) |
| `DELETE` | `/api/user/me` | Excluir a própria conta |
| `GET` | `/api/user` | Listar todos (**admin only**) |

---

## 7. Códigos de erro

Todas as respostas de erro têm o formato:

```json
{ "error": "mensagem descritiva" }
```

| Código | Quando |
|---|---|
| `400` | JSON inválido, campo obrigatório faltando, JID malformado |
| `401` | Token ausente/inválido/expirado, credenciais erradas |
| `403` | Operação requer role `admin` |
| `404` | Instância/usuário não encontrado ou não é do dono |
| `409` | E-mail já cadastrado, ID de instância duplicado |
| `413` | Upload de imagem > 5 MB |
| `429` | Rate limit excedido em `/auth/*` |
| `500` | Erro interno (mensagem genérica — detalhes nos logs do servidor) |
| `503` | Instância não está `connected` para o envio |

---

## 8. Exemplo end-to-end (cURL)

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@akirawhats.com","password":"admin123"}' \
  | jq -r .token)

# 2. Criar instância com webhook
curl -X POST http://localhost:8080/api/instance \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":"prod1","webhookUrl":"https://meuapp.com/wa"}'

# 3. Pegar QR (repetir até conectar)
curl http://localhost:8080/api/instance/prod1/qr \
  -H "Authorization: Bearer $TOKEN"

# 4. Conferir status
curl http://localhost:8080/api/instance/prod1 \
  -H "Authorization: Bearer $TOKEN"

# 5. Enviar texto
curl -X POST http://localhost:8080/api/instance/prod1/send/text \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"to":"5511999998888","message":"Oi do AkiraWhats!"}'

# 6. Enviar imagem
curl -X POST http://localhost:8080/api/instance/prod1/send/image \
  -H "Authorization: Bearer $TOKEN" \
  -F "to=5511999998888" \
  -F "caption=Veja isso" \
  -F "file=@./foto.jpg"
```

---

## 9. CORS

A API valida `CORS_ORIGINS` (env, comma-separated). Em produção, configure com a(s) origem(ns) do seu frontend:

```
CORS_ORIGINS=https://app.minhaempresa.com,https://staging.minhaempresa.com
```

Em dev a variável pode ser omitida (permite qualquer origem com log de aviso).

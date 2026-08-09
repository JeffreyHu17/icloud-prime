# iCloud Prime API

Base URL:

```text
http://127.0.0.1:8081
```

All API responses use this shape:

```json
{
  "success": true,
  "data": {}
}
```

Errors use:

```json
{
  "success": false,
  "message": "error message"
}
```

Do not send real Cookie values or App-specific passwords in public bug reports.

## Accounts

### List Accounts

```http
GET /api/accounts
```

Sensitive fields are redacted from the response.

### Add Account

```http
POST /api/accounts
Content-Type: application/json
```

Body:

```json
{
  "name": "Main account",
  "host": "icloud.com",
  "cookies": "X-APPLE-WEBAUTH-TOKEN=value; X-APPLE-WEBAUTH-USER=value",
  "proxy": ""
}
```

Fields:

- `name`: required display name.
- `host`: optional, usually `icloud.com` or `icloud.com.cn`.
- `cookies`: optional Cookie input as header string or JSON object string.
- `proxy`: optional HTTP/SOCKS5 proxy URL.

### Delete Account

```http
DELETE /api/accounts/:id
```

### Set App-Specific Password

```http
POST /api/accounts/:id/password
Content-Type: application/json
```

Body:

```json
{
  "icloud_email": "your_email@icloud.com",
  "app_password": "xxxx-xxxx-xxxx-xxxx"
}
```

The App-specific password is used for IMAP mail reading.

### Update Cookies

```http
PUT /api/accounts/:id/cookies
Content-Type: application/json
```

Body:

```json
{
  "cookies": {
    "X-APPLE-WEBAUTH-TOKEN": "value",
    "X-APPLE-WEBAUTH-USER": "value",
    "X-APPLE-DS-WEB-SESSION-TOKEN": "value"
  }
}
```

### Login Account

```http
POST /api/accounts/:id/login
Content-Type: application/json
```

Body:

```json
{
  "password": "apple-id-password",
  "otp_code": "123456"
}
```

`otp_code` is optional when 2FA is not required.

## Hide My Email Aliases

### Create Alias

```http
POST /api/create
Content-Type: application/json
```

Body:

```json
{
  "account_id": "acc_1",
  "label": "Example site"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "email": "alias@icloud.com",
    "label": "Example site",
    "created_at": "2026-01-01T00:00:00Z",
    "account_id": "acc_1"
  }
}
```

### List Aliases

```http
GET /api/aliases?account_id=acc_1
```

### Deactivate Alias

```http
POST /api/aliases/:id/deactivate
Content-Type: application/json
```

Body:

```json
{
  "account_id": "acc_1"
}
```

### Reactivate Alias

```http
POST /api/aliases/:id/reactivate
Content-Type: application/json
```

Body:

```json
{
  "account_id": "acc_1"
}
```

### Delete Alias

```http
DELETE /api/aliases/:id
Content-Type: application/json
```

Body:

```json
{
  "account_id": "acc_1"
}
```

## Mail

### Read Inbox

```http
GET /api/inbox?account_id=acc_1&alias=alias@icloud.com&limit=20&days=7
```

Query parameters:

- `account_id`: required.
- `alias`: optional; when present, filters messages sent to that alias.
- `folder`: optional; defaults to `inbox`.
- `limit`: optional; defaults to `20`.
- `days`: optional; defaults to `7` for IMAP mode.

Read order:

1. IMAP through App-specific password.
2. Web API through Cookie fallback.

Response includes `method` with `imap` or `web_api`.

### List Mailboxes

```http
GET /api/mailboxes?account_id=acc_1
```

Requires an App-specific password.

## System

### Reload Account Configuration

```http
POST /api/reload
```

Reloads `data/accounts.json` without restarting the process.

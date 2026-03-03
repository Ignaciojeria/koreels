# Desarrollo local

## Servicios con Docker

Levantar Postgres y Dex (IdP local para OIDC):

```bash
docker-compose up -d
```

- **Postgres**: `localhost:5432`, base `ledger`, usuario/contraseña `postgres/postgres`.
- **Dex**: `http://localhost:5556/dex` (issuer para OIDC). Discovery: `http://localhost:5556/dex/.well-known/openid-configuration`.

## BFF con OIDC (Dex local)

Para que el BFF valide JWTs con Dex, configura en `.env` (o en el entorno):

```env
OIDC_ISSUER=http://localhost:5556/dex
OIDC_CLIENT_ID=ledger-bff
```

Opcional (si el IdP exige un audience distinto del client id):

```env
OIDC_AUDIENCE=ledger-bff
```

Arranca el BFF:

```bash
go run ./cmd/bff
```

Las rutas protegidas (ej. `GET /me/balance`) requieren cabecera `Authorization: Bearer <token>`.

## Obtener un token de Dex (local)

Usuario de prueba en `dex-config.yaml`: **dev** / **password** (bcrypt del ejemplo).

### Opción 1: Resource owner password grant (curl)

Con el cliente `ledger-bff` y secret `ledger-secret`:

```bash
curl -s -X POST http://localhost:5556/dex/token \
  -d "grant_type=password" \
  -d "username=dev" \
  -d "password=password" \
  -d "client_id=ledger-bff" \
  -d "client_secret=ledger-secret" \
  -d "scope=openid"
```

La respuesta incluye `id_token` (JWT). Úsalo en el BFF:

```bash
export TOKEN="<id_token de la respuesta>"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/me/balance
```

### Opción 2: Flujo authorization code (navegador)

1. Abre en el navegador (sustituye `STATE` por un valor aleatorio):
   `http://localhost:5556/dex/auth?client_id=ledger-bff&redirect_uri=http://localhost:8080/callback&response_type=code&scope=openid&state=STATE`
2. Inicia sesión con **dev** / **password**.
3. Dex redirige a `http://localhost:8080/callback?code=...&state=STATE`. Intercambia el `code` por tokens en el endpoint `/dex/token` (client_id + client_secret) y usa el `id_token` como Bearer en el BFF.

## Postman

Importa la colección `doc/postman/ledger-service-local.json` en Postman. Variables de colección:

- **base_url**: `http://localhost:8081` (o el puerto de tu `.env`).
- **token**: el `id_token` obtenido de Dex (solo para las peticiones BFF).
- **accountId**, **transactionId**: UUIDs de ejemplo; créalos con POST /accounts o reutiliza los que devuelve la API.

Para las peticiones BFF, obtén antes un token con el curl de Dex y pégalo en la variable `token`. **Primer uso:** llama primero a **POST /me/account** (crea tu cuenta o devuelve la existente de forma idempotente); luego **GET /me/balance** para ver el saldo.

## Sin OIDC (cmd/ledger, cmd/api)

Si no defines `OIDC_ISSUER`, el middleware de auth es no-op. `cmd/ledger` y `cmd/api` no requieren token; el BFF sin variables OIDC tampoco validará JWT (pero `/me/balance` seguirá exigiendo identidad en context, que solo tendrás si algún middleware la inyecta; en ese caso devolverá 401 sin token válido).

## Producción

Mismo código; cambia las variables de entorno al IdP real:

- `OIDC_ISSUER=https://tu-tenant.auth0.com/` (o tu IdP)
- `OIDC_CLIENT_ID=<client id registrado en el IdP>`
- `OIDC_AUDIENCE` si el IdP lo exige

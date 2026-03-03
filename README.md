# ledger-service

[![Go Report Card](https://goreportcard.com/badge/github.com/Ignaciojeria/ledger-service)](https://goreportcard.com/report/github.com/Ignaciojeria/ledger-service)
[![codecov](https://codecov.io/gh/Ignaciojeria/ledger-service/graph/badge.svg)](https://codecov.io/gh/Ignaciojeria/ledger-service)

## Entry points

| Carpeta    | Uso | Expone API ledger |
|------------|-----|-------------------|
| **ledger/** | Ledger como servicio independiente. Solo la API del libro mayor. | Sí |
| **api/**    | Todas las APIs juntas (monolito). Por ahora igual que ledger; luego se pueden sumar más módulos. | Sí |
| **bff/**    | BFF: rutas orientadas al cliente (ej. `/me/balance`). No importa el adapter HTTP del ledger. | No |

## Cómo correr

```bash
# Ledger solo (servicio independiente)
go run ./cmd/ledger

# Todas las APIs (monolito)
go run ./cmd/api

# BFF (sin exponer /accounts, /transactions del ledger)
go run ./cmd/bff
```

El BFF usa el ledger en memoria (use cases) y solo registra sus propias rutas; al no importar `internal/ledger/adapter/in/fuegoapi`, las rutas del ledger no se exponen.

## BFF y OIDC

Para que el BFF exija autenticación JWT, define `OIDC_ISSUER` y `OIDC_CLIENT_ID` (por ejemplo con Dex en local). Sin ellas el middleware de auth es no-op; las rutas como `/me/balance` requieren identidad en el context y devolverán 401 si no hay token válido. Ver [doc/dev.md](./doc/dev.md) para levantar Dex y obtener un token.

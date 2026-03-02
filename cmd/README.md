# Entry points

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

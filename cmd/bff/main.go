// BFF: expone solo rutas orientadas al cliente (ej. /me/balance).
// No importa internal/ledger/adapter/in/fuegoapi, así que el API del ledger no se expone.
// El ledger se usa en memoria vía los use cases.
package main

import (
	"log"
	"os"
	"strings"

	ledger "koreels"
	_ "koreels/internal/bff"
	_ "koreels/internal/ledger/adapter/out/postgres"
	_ "koreels/internal/ledger/application/usecase"
	_ "koreels/internal/shared/configuration"
	_ "koreels/internal/shared/infrastructure/httpserver"
	_ "koreels/internal/shared/infrastructure/httpserver/middleware"
	_ "koreels/internal/shared/infrastructure/observability"
	_ "koreels/internal/shared/infrastructure/postgresql"

	"github.com/Ignaciojeria/ioc"
)

func main() {
	os.Setenv("VERSION", strings.TrimSpace(ledger.Version))

	if err := ioc.LoadDependencies(); err != nil {
		log.Fatal("Failed to load dependencies:", err)
	}
}

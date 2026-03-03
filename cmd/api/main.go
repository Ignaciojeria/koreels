// Todas las APIs juntas (monolito): ledger API + en el futuro otros módulos.
// Mismo comportamiento que cmd/ledger por ahora.
package main

import (
	"log"
	"os"
	"strings"

	ledger "koreels"
	_ "koreels/internal/ledger/adapter/in/fuegoapi"
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

// Ledger como servicio independiente: expone la API del ledger (POST/GET accounts, transactions).
package main

import (
	"log"
	"os"
	"strings"

	ledger "ledger-service"
	_ "ledger-service/internal/ledger/adapter/in/fuegoapi"
	_ "ledger-service/internal/ledger/adapter/out/postgres"
	_ "ledger-service/internal/ledger/application/usecase"
	_ "ledger-service/internal/shared/configuration"
	_ "ledger-service/internal/shared/infrastructure/httpserver"
	_ "ledger-service/internal/shared/infrastructure/httpserver/middleware"
	_ "ledger-service/internal/shared/infrastructure/observability"
	_ "ledger-service/internal/shared/infrastructure/postgresql"

	"github.com/Ignaciojeria/ioc"
)

func main() {
	os.Setenv("VERSION", strings.TrimSpace(ledger.Version))

	if err := ioc.LoadDependencies(); err != nil {
		log.Fatal("Failed to load dependencies:", err)
	}
}

// Reelgen como servicio independiente.
package main

import (
	"log"
	"os"
	"strings"

	reelgen "koreels"
	_ "koreels/internal/reelgen/adapter/in/fuegoapi"
	_ "koreels/internal/reelgen/adapter/out/qwenapi"
	_ "koreels/internal/reelgen/application/usecase"
	_ "koreels/internal/shared/configuration"
	_ "koreels/internal/shared/infrastructure/eventbus"
	_ "koreels/internal/shared/infrastructure/httpserver"
	_ "koreels/internal/shared/infrastructure/httpserver/middleware"
	_ "koreels/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
)

func main() {
	os.Setenv("VERSION", strings.TrimSpace(reelgen.Version))

	if err := ioc.LoadDependencies(); err != nil {
		log.Fatal("Failed to load dependencies:", err)
	}
}

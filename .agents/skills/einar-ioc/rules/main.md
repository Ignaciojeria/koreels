# main

> Application entry point - loads IoC dependencies

**Blank imports:** Each package that registers with `ioc.Register` MUST be imported in `cmd/api/main.go` via `_ "archetype/path/to/package"` (not in `version.go`, which is package archetype and only embeds Version). When adding a new component, the agent must add its blank import so the package loads and the IoC receives the constructor. Example:

```go
import (
    "log"
    "os"
    "strings"
    "archetype"
    _ "archetype/app/adapter/in/fuegoapi"
    _ "archetype/app/adapter/in/eventbus"
    _ "archetype/app/adapter/out/postgres"
    _ "archetype/app/application/usecase"
    _ "archetype/app/shared/infrastructure/httpserver"
    _ "archetype/app/shared/infrastructure/postgresql"
    _ "archetype/app/shared/infrastructure/observability"
    "github.com/Ignaciojeria/ioc"
)
```

## cmd/api/main.go

```go
package main

import (
	"log"
	"os"
	"strings"

	"archetype"

	"github.com/Ignaciojeria/ioc"
)

func main() {
	os.Setenv("VERSION", strings.TrimSpace(archetype.Version))

	if err := ioc.LoadDependencies(); err != nil {
		log.Fatal("Failed to load dependencies:", err)
	}
}
```

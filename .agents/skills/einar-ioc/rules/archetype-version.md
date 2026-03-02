# archetype-version

> Root package - embedded Version from .version file

## version.go

```go
package archetype

import (
	_ "embed"
)

//go:embed .version
var Version string
```

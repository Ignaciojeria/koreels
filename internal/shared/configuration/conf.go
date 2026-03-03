package configuration

import (
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewConf)

type Conf struct {
	PORT         string `env:"PORT" envDefault:"8080"`
	PROJECT_NAME string `env:"PROJECT_NAME"`
	VERSION      string `env:"VERSION"`
	DATABASE_URL string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/ledger?sslmode=disable"`

	// OIDC: cuando OIDC_ISSUER está vacío el middleware de auth es no-op (cmd/ledger, cmd/api).
	// En cmd/bff se define en env para activar validación JWT.
	OIDC_ISSUER    string `env:"OIDC_ISSUER"`     // ej. http://localhost:5556/dex (local) o https://tenant.auth0.com/
	OIDC_CLIENT_ID string `env:"OIDC_CLIENT_ID"`  // client id del cliente OIDC
	OIDC_AUDIENCE  string `env:"OIDC_AUDIENCE"`   // opcional; si no se define el middleware usa OIDC_CLIENT_ID como aud
}

func NewConf() (Conf, error) {
	return Parse[Conf]()
}

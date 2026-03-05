package eventbus

import (
	"log"
	"time"

	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var _ = ioc.Register(NewNatsClient)

// NatsClient encapsulates an in-memory NATS server and the connection to it.
type NatsClient struct {
	EmbeddedServer *server.Server
	Connection     *nats.Conn
}

func NewNatsClient(conf configuration.Conf) (*NatsClient, error) {
	if conf.EVENT_BROKER != "nats" {
		return nil, nil
	}

	opts := &server.Options{
		Port: -1,
	}

	embedded, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	go embedded.Start()

	if !embedded.ReadyForConnections(5 * time.Second) {
		return nil, err
	}

	log.Printf("Embedded NATS Server running on %s", embedded.ClientURL())

	conn, err := nats.Connect(embedded.ClientURL())
	if err != nil {
		return nil, err
	}

	return &NatsClient{
		EmbeddedServer: embedded,
		Connection:     conn,
	}, nil
}

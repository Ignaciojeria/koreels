# einar-template

> Einar CLI config - defines how generators create and rename .go files

**How it works:** For each controller, repository, consumer, publisher, etc., Einar CLI generates a **new file** in the structure indicated by `destination_dir`, with a name derived from the template's `source_file` and the operation/entity name. The content is built from the template with `replace_holders` applied (PascalCase/SnakeCase). `ioc_discovery: true` means the CLI adds the package to blank imports in `cmd/api/main.go`.

## .einar.template.json

```json
{
    "installations_base": [],
    "base_template": {
        "folders": [],
        "files": [
            {
                "source_file": "app/shared/configuration/conf.go",
                "destination_file": "app/shared/configuration/conf.go"
            },
            {
                "source_file": "app/shared/configuration/conf_test.go",
                "destination_file": "app/shared/configuration/conf_test.go"
            },
            {
                "source_file": "app/shared/configuration/parse.go",
                "destination_file": "app/shared/configuration/parse.go"
            },
            {
                "source_file": "cmd/api/main.go",
                "destination_file": "cmd/api/main.go"
            },
            {
                "source_file": ".version",
                "destination_file": ".version"
            },
            {
                "source_file": ".environment",
                "destination_file": ".env"
            },
            {
                "source_file": ".gitignore",
                "destination_file": ".gitignore"
            },
            {
                "source_file": "version.go",
                "destination_file": "version.go"
            },
            {
                "source_file": "README.md",
                "destination_file": "README.md"
            }
        ]
    },
    "component_commands": [
        {
            "kind": "get-controller",
            "adapter_type": "inbound",
            "command": "einar generate get-controller ${operation_name}",
            "depends_on": [
                "fuego"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/fuegoapi/get.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateGet",
                            "append_at_start": "New",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/fuegoapi/get_test.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateGet",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplateGet",
                            "append_at_start": "TestNew",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "post-controller",
            "adapter_type": "inbound",
            "command": "einar generate post-controller ${operation_name}",
            "depends_on": [
                "fuego"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/fuegoapi/post.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePost",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePostRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePostResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/fuegoapi/post_test.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePost",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplatePost",
                            "append_at_start": "TestNew",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePostRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePostResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "put-controller",
            "adapter_type": "inbound",
            "command": "einar generate put-controller ${operation_name}",
            "depends_on": [
                "fuego"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/fuegoapi/put.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePut",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePutRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePutResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/fuegoapi/put_test.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePut",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplatePut",
                            "append_at_start": "TestNew",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePutRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePutResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "patch-controller",
            "adapter_type": "inbound",
            "command": "einar generate patch-controller ${operation_name}",
            "depends_on": [
                "fuego"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/fuegoapi/patch.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePatch",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePatchRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePatchResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/fuegoapi/patch_test.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePatch",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplatePatch",
                            "append_at_start": "TestNew",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePatchRequest",
                            "append_at_start": "",
                            "append_at_end": "Request"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePatchResponse",
                            "append_at_start": "",
                            "append_at_end": "Response"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "delete-controller",
            "adapter_type": "inbound",
            "command": "einar generate delete-controller ${operation_name}",
            "depends_on": [
                "fuego"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/fuegoapi/delete.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateDelete",
                            "append_at_start": "New",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/fuegoapi/delete_test.go",
                    "destination_dir": "app/adapter/in/fuegoapi",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateDelete",
                            "append_at_start": "New",
                            "append_at_end": ""
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplateDelete",
                            "append_at_start": "TestNew",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "postgres-repository",
            "adapter_type": "outbound",
            "command": "einar generate postgres-repository ${operation_name}",
            "depends_on": [
                "postgresql"
            ],
            "files": [
                {
                    "source_file": "app/adapter/out/postgres/postgres_repository.go",
                    "destination_dir": "app/adapter/out/postgres",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateRepository",
                            "append_at_start": "New",
                            "append_at_end": "Repository"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateRepository",
                            "append_at_start": "",
                            "append_at_end": "Repository"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateStruct",
                            "append_at_start": "",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/out/postgres/postgres_repository_test.go",
                    "destination_dir": "app/adapter/out/postgres",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateRepository",
                            "append_at_start": "New",
                            "append_at_end": "Repository"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplateRepository",
                            "append_at_start": "TestNew",
                            "append_at_end": "Repository"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateRepository",
                            "append_at_start": "",
                            "append_at_end": "Repository"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "pubsub-consumer",
            "adapter_type": "inbound",
            "command": "einar generate pubsub-consumer ${operation_name}",
            "depends_on": [
                "gcp-pubsub"
            ],
            "files": [
                {
                    "source_file": "app/adapter/in/eventbus/consumer.go",
                    "destination_dir": "app/adapter/in/eventbus",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateConsumer",
                            "append_at_start": "New",
                            "append_at_end": "Consumer"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateConsumer",
                            "append_at_start": "",
                            "append_at_end": "Consumer"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateMessage",
                            "append_at_start": "",
                            "append_at_end": "Message"
                        },
                        {
                            "kind": "SnakeCase",
                            "name": "template_topic_or_hook",
                            "append_at_start": "",
                            "append_at_end": ""
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/in/eventbus/consumer_test.go",
                    "destination_dir": "app/adapter/in/eventbus",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplateConsumer",
                            "append_at_start": "New",
                            "append_at_end": "Consumer"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplateConsumer",
                            "append_at_start": "TestNew",
                            "append_at_end": "Consumer"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateConsumer",
                            "append_at_start": "",
                            "append_at_end": "Consumer"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplateMessage",
                            "append_at_start": "",
                            "append_at_end": "Message"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        },
        {
            "kind": "publisher",
            "adapter_type": "outbound",
            "command": "einar generate publisher ${publisher_name}",
            "depends_on": [
                "gcp-pubsub"
            ],
            "files": [
                {
                    "source_file": "app/adapter/out/eventbus/publisher.go",
                    "destination_dir": "app/adapter/out/eventbus",
                    "ioc_discovery": true,
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePublisher",
                            "append_at_start": "New",
                            "append_at_end": "Publisher"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePublisher",
                            "append_at_start": "",
                            "append_at_end": "Publisher"
                        }
                    ],
                    "literal_replacements": []
                },
                {
                    "source_file": "app/adapter/out/eventbus/publisher_test.go",
                    "destination_dir": "app/adapter/out/eventbus",
                    "ioc_discovery": true,
                    "append_at_end": "_test",
                    "replace_holders": [
                        {
                            "kind": "PascalCase",
                            "name": "NewTemplatePublisher",
                            "append_at_start": "New",
                            "append_at_end": "Publisher"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TestNewTemplatePublisher",
                            "append_at_start": "TestNew",
                            "append_at_end": "Publisher"
                        },
                        {
                            "kind": "PascalCase",
                            "name": "TemplatePublisher",
                            "append_at_start": "",
                            "append_at_end": "Publisher"
                        }
                    ],
                    "literal_replacements": []
                }
            ]
        }
    ],
    "installation_commands": [
        {
            "name": "fuego",
            "unique": "http-server",
            "depends_on": [
                "observability"
            ],
            "files": [
                {
                    "source_file": "app/shared/infrastructure/httpserver/server.go",
                    "destination_dir": "app/shared/infrastructure/httpserver",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/httpserver/server_test.go",
                    "destination_dir": "app/shared/infrastructure/httpserver",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/httpserver/middleware/request_logger.go",
                    "destination_dir": "app/shared/infrastructure/httpserver/middleware",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/httpserver/middleware/request_logger_test.go",
                    "destination_dir": "app/shared/infrastructure/httpserver/middleware",
                    "ioc_discovery": true
                }
            ],
            "libraries": [
                "github.com/go-fuego/fuego",
                "github.com/hellofresh/health-go/v5"
            ]
        },
        {
            "name": "gcp-pubsub",
            "depends_on": [
                "observability"
            ],
            "files": [
                {
                    "source_file": "app/shared/infrastructure/eventbus/strategy.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": false
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/gcp_client.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/gcp_publisher.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/gcp_subscriber.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                }
            ],
            "libraries": [
                "cloud.google.com/go/pubsub",
                "github.com/cloudevents/sdk-go/v2"
            ]
        },
        {
            "name": "observability",
            "depends_on": [],
            "files": [
                {
                    "source_file": "app/shared/infrastructure/observability/observability.go",
                    "destination_dir": "app/shared/infrastructure/observability",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/observability/observability_test.go",
                    "destination_dir": "app/shared/infrastructure/observability",
                    "ioc_discovery": true
                }
            ],
            "libraries": [
                "go.opentelemetry.io/otel",
                "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp",
                "go.opentelemetry.io/otel/propagation",
                "go.opentelemetry.io/otel/sdk/resource",
                "go.opentelemetry.io/otel/sdk/trace",
                "go.opentelemetry.io/otel/semconv/v1.26.0",
                "go.opentelemetry.io/otel/trace",
                "go.opentelemetry.io/otel/trace/noop"
            ]
        },
        {
            "name": "nats-in-memory",
            "depends_on": [
                "observability"
            ],
            "files": [
                {
                    "source_file": "app/shared/infrastructure/eventbus/strategy.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": false
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/nats_client.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/nats_publisher.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                },
                {
                    "source_file": "app/shared/infrastructure/eventbus/nats_subscriber.go",
                    "destination_dir": "app/shared/infrastructure/eventbus",
                    "ioc_discovery": true
                }
            ],
            "libraries": [
                "github.com/nats-io/nats.go",
                "github.com/nats-io/nats-server/v2",
                "github.com/cloudevents/sdk-go/v2"
            ]
        }
    ]
}
```

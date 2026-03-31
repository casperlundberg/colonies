# Plugin System

Backend implementations for ColonyOS live in this directory. Each plugin registers
itself via Go's `init()` function when imported. The default binary (`cmd/main.go`)
imports all plugins. Custom builds can import a subset.

## Default Plugins

| Directory        | Description                                                        |
|------------------|--------------------------------------------------------------------|
| `postgresql/`    | PostgreSQL/TimescaleDB database backend                            |
| `embedded/`      | Embedded key-value database (no external dependencies, suitable for development and edge) |
| `gin/`           | Gin HTTP framework adapter with WebSocket support                  |
| `localfs/`       | Local filesystem object storage                                    |
| `s3/`            | S3-compatible object storage (MinIO, AWS S3, etc.)                 |
| `etcd/`          | etcd-based clustering, leader election, and relay server           |
| `prometheus/`    | Prometheus metrics exporter                                        |

## How Plugins Work

Each plugin has a `register.go` file that calls a `Register()` function in the
corresponding interface package during `init()`. Example from postgresql:

```go
package postgresql

import "github.com/colonyos/colonies/pkg/database"

func init() {
    database.Register("postgresql", func(config database.DatabaseConfig) (database.Database, error) {
        db := CreatePQDatabase(config.Host, config.Port, config.User, config.Password, config.Name, config.Prefix, config.TimescaleDB)
        err := db.Connect()
        return db, err
    })
}
```

When a plugin package is imported (even as a blank import `_ "..."`), Go
executes its `init()` function, which registers a factory with the interface
package. At runtime the server calls `Create()` on the interface package to
instantiate the selected backend by name.

## Writing an External Plugin

1. Create a new Go module.
2. Import the interface package (e.g., `github.com/colonyos/colonies/pkg/database`).
3. Implement the interface.
4. Add a `register.go` with an `init()` function that calls `Register()`.
5. Users import your plugin with a blank import in their `main.go`.

Minimal external plugin example:

```go
// In github.com/someone/colonies-mysql/register.go
package mysql

import "github.com/colonyos/colonies/pkg/database"

func init() {
    database.Register("mysql", func(config database.DatabaseConfig) (database.Database, error) {
        return NewMySQLDatabase(config)
    })
}
```

## Building a Custom Binary

Create a `main.go` that imports only the plugins you need:

```go
package main

import (
    "github.com/colonyos/colonies/pkg/app"
    "github.com/colonyos/colonies/pkg/build"

    _ "github.com/colonyos/colonies/plugin/embedded"
    _ "github.com/colonyos/colonies/plugin/gin"
    _ "github.com/colonyos/colonies/plugin/localfs"
    _ "github.com/someone/colonies-mysql"  // external plugin
)

func main() {
    build.BuildVersion = "custom"
    app.Execute()
}
```

Then build:

```
go mod init my-colonies && go mod tidy && go build -o colonies .
```

## Interface Packages

Each interface package defines the contract a plugin must satisfy and contains a
`registry.go` with the `Register()` and `Create()` functions.

| Interface    | Location                          |
|--------------|-----------------------------------|
| Database     | `pkg/database/database.go`        |
| HTTP Backend | `pkg/backends/interfaces.go`      |
| File Storage | `pkg/fs/objectstore.go`           |
| Clustering   | `pkg/cluster/interfaces.go`       |
| Monitoring   | `pkg/monitoring/interfaces.go`    |

package main

import (
	"github.com/colonyos/colonies/internal/cli"
	"github.com/colonyos/colonies/pkg/build"

	// Default plugins -- remove or add imports to customize your build.
	// Each plugin registers itself via init() when imported.
	_ "github.com/colonyos/colonies/plugin/embedded"
	_ "github.com/colonyos/colonies/plugin/gin"
	_ "github.com/colonyos/colonies/plugin/localfs"
	_ "github.com/colonyos/colonies/plugin/postgresql"
	_ "github.com/colonyos/colonies/plugin/prometheus"
	_ "github.com/colonyos/colonies/plugin/s3"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime
	cli.Execute()
}

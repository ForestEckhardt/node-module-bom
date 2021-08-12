package main

import (
	"os"

	nodemodulebom "github.com/paketo-buildpacks/node-module-bom"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {

	packit.Run(
		nodemodulebom.Detect(),
		nodemodulebom.Build(
			postal.NewService(cargo.NewTransport()),
			nil,
			chronos.DefaultClock,
			scribe.NewEmitter(os.Stdout),
		),
	)
}

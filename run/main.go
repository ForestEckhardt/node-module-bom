package main

import (
	nodemodulebom "github.com/paketo-buildpacks/node-module-bom"
	"github.com/paketo-buildpacks/packit"
)

func main() {

	packit.Run(nodemodulebom.Detect(), nodemodulebom.Build())
}

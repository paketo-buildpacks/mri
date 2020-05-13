package main

import (
	"os"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-community/mri/mri"
)

func main() {
	logEmitter := mri.NewLogEmitter(os.Stdout)
	entryResolver := mri.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := mri.NewPlanRefinery()
	clock := mri.NewClock(time.Now)
	gem := pexec.NewExecutable("gem")

	packit.Build(mri.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock, gem))
}

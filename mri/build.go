package mri

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan
}

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

func Build(entries EntryResolver, dependencies DependencyManager, planRefinery BuildPlanRefinery, logger LogEmitter, clock Clock, gem Executable) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving MRI version")

		entry := entries.Resolve(context.Plan.Entries)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, entry.Version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		mriLayer, err := context.Layers.Get(MRI)
		if err != nil {
			return packit.BuildResult{}, err
		}

		mriLayer.Launch = entry.Metadata["launch"] == true
		mriLayer.Build = entry.Metadata["build"] == true
		mriLayer.Cache = entry.Metadata["build"] == true

		bom := planRefinery.BillOfMaterial(postal.Dependency{
			ID:      dependency.ID,
			Name:    dependency.Name,
			SHA256:  dependency.SHA256,
			Stacks:  dependency.Stacks,
			URI:     dependency.URI,
			Version: dependency.Version,
		})

		cachedSHA, ok := mriLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", mriLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{mriLayer},
			}, nil
		}

		logger.Process("Executing build process")

		err = mriLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		mriLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		logger.Subprocess("Installing MRI %s", dependency.Version)
		then := clock.Now()
		err = dependencies.Install(dependency, context.CNBPath, mriLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Action("Completed in %s", time.Since(then).Round(time.Millisecond))
		logger.Break()

		os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(mriLayer.Path, "bin"), os.Getenv("PATH")))

		buffer := bytes.NewBuffer(nil)
		err = gem.Execute(pexec.Execution{
			Args:   []string{"env", "path"},
			Stdout: buffer,
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		mriLayer.SharedEnv.Override("GEM_PATH", strings.TrimSpace(buffer.String()))

		logger.Environment(mriLayer.SharedEnv)

		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{mriLayer},
		}, nil
	}
}

package mri

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

func Build(entries EntryResolver, dependencies DependencyManager, logger scribe.Emitter, clock chronos.Clock, gem Executable) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving MRI version")

		entry, allEntries := entries.Resolve("mri", context.Plan.Entries, []interface{}{"BP_MRI_VERSION", "buildpack.yml"})
		logger.Candidates(allEntries)

		// NOTE: this is to override that the dependency is called "ruby" in the
		// buildpack.toml. We can remove this once we update our own dependencies
		// and can name it however we like.
		entry.Name = "ruby"
		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// NOTE: this is to override that the dependency is called "ruby" in the
		// buildpack.toml. We can remove this once we update our own dependencies
		// and can name it however we like.
		dependency.ID = "mri"
		dependency.Name = "MRI"

		logger.SelectedDependency(entry, dependency, clock.Now())

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information.")
			logger.Break()
		}

		mriLayer, err := context.Layers.Get(MRI)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entries.MergeLayerTypes("mri", context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		cachedSHA, ok := mriLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", mriLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Layers: []packit.Layer{mriLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		mriLayer, err = mriLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		mriLayer.Launch, mriLayer.Build, mriLayer.Cache = launch, build, build

		logger.Subprocess("Installing MRI %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, mriLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		mriLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

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

		logger.EnvironmentVariables(mriLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{mriLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}

package mri

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface Executable --output fakes/executable.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

type Executable interface {
	Execute(pexec.Execution) error
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func Build(
	entries EntryResolver,
	dependencies DependencyManager,
	gem Executable,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
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

		logger.Debug.Process("Getting the layer associated with MRI:")
		mriLayer, err := context.Layers.Get(MRI)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(mriLayer.Path)
		logger.Debug.Break()

		legacySBOM := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entries.MergeLayerTypes("mri", context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = legacySBOM
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = legacySBOM
		}

		cachedSHA, ok := mriLayer.Metadata[DepKey].(string)

		if ok && cargo.Checksum(dependency.Checksum).MatchString(cachedSHA) {
			logger.Process("Reusing cached layer %s", mriLayer.Path)
			logger.Break()

			mriLayer.Launch, mriLayer.Build, mriLayer.Cache = launch, build, build

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
			logger.Debug.Subprocess("Installation path: %s", mriLayer.Path)
			logger.Debug.Subprocess("Source URI: %s", dependency.URI)
			return dependencies.Deliver(dependency, context.CNBPath, mriLayer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(mriLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, mriLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		mriLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		mriLayer.Metadata = map[string]interface{}{
			DepKey: dependency.Checksum,
		}

		logger.Debug.Process("Adding %s to the $PATH", filepath.Join(mriLayer.Path, "bin"))
		logger.Debug.Break()
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

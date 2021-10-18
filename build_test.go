package mri_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/mri/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		cnbDir    string
		clock     chronos.Clock
		timeStamp time.Time
		buffer    *bytes.Buffer

		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		gem               *fakes.Executable

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
  id = "org.some-org.some-buildpack"
  name = "Some Buildpack"
  version = "some-version"

[metadata]
  [metadata.default-versions]
    mri = "2.5.x"

  [[metadata.dependencies]]
    id = "some-dep"
    name = "Some Dep"
    sha256 = "some-sha"
    stacks = ["some-stack"]
    uri = "some-uri"
    version = "some-dep-version"
`), 0600)
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "mri",
			Metadata: map[string]interface{}{
				"version-source": "buildpack.yml",
				"version":        "2.5.x",
				"launch":         true,
			},
		}

		entryResolver.MergeLayerTypesCall.Returns.Launch = true
		entryResolver.MergeLayerTypesCall.Returns.Build = false

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{ID: "ruby", Name: "Ruby"}
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "mri",
				Metadata: packit.BOMMetadata{
					Version: "mri-dependency-version",
					Checksum: packit.BOMChecksum{
						Algorithm: packit.SHA256,
						Hash:      "mri-dependency-sha",
					},
					URI: "mri-dependency-uri",
				},
			},
		}

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)
		gem = &fakes.Executable{}
		gem.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "/some/mri/gems/path")
			return nil
		}

		build = mri.Build(entryResolver, dependencyManager, logEmitter, clock, gem)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs mri", func() {
		result, err := build(packit.BuildContext{
			CNBPath: cnbDir,
			Stack:   "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "0.1.2",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "mri",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "2.5.x",
							"launch":         true,
						},
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name: "mri",
					Path: filepath.Join(layersDir, "mri"),
					SharedEnv: packit.Environment{
						"GEM_PATH.override": "/some/mri/gems/path",
					},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           true,
					Cache:            false,
					Metadata: map[string]interface{}{
						mri.DepKey: "",
						"built_at": timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
			Launch: packit.LaunchMetadata{
				BOM: []packit.BOMEntry{
					{
						Name: "mri",
						Metadata: packit.BOMMetadata{
							Version: "mri-dependency-version",
							Checksum: packit.BOMChecksum{
								Algorithm: packit.SHA256,
								Hash:      "mri-dependency-sha",
							},
							URI: "mri-dependency-uri",
						},
					},
				},
			},
		}))

		Expect(filepath.Join(layersDir, "mri")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "mri",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "2.5.x",
					"launch":         true,
				},
			},
		}))

		Expect(entryResolver.MergeLayerTypesCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "mri",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "2.5.x",
					"launch":         true,
				},
			},
		}))
		Expect(entryResolver.MergeLayerTypesCall.Receives.String).To(Equal("mri"))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("ruby"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("2.5.x"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{{ID: "mri", Name: "MRI"}}))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{ID: "mri", Name: "MRI"}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "mri")))

		Expect(gem.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"env", "path"}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.1.2"))
		Expect(buffer.String()).To(ContainSubstring("Resolving MRI version"))
		Expect(buffer.String()).To(ContainSubstring("Selected MRI version (using buildpack.yml): "))
		Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v1.0.0."))
		Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information."))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Configuring build environment"))
		Expect(buffer.String()).To(ContainSubstring("Configuring launch environment"))
	})

	context("when the build plan entry does not include a version", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "mri",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			}
		})

		it("picks the newest version", func() {
			result, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Stack:   "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "mri",
							Metadata: map[string]interface{}{
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name: "mri",
						Path: filepath.Join(layersDir, "mri"),
						SharedEnv: packit.Environment{
							"GEM_PATH.override": "/some/mri/gems/path",
						},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            false,
						Launch:           true,
						Cache:            false,
						Metadata: map[string]interface{}{
							mri.DepKey: "",
							"built_at": timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "mri",
							Metadata: packit.BOMMetadata{
								Version: "mri-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "mri-dependency-sha",
								},
								URI: "mri-dependency-uri",
							},
						},
					},
				},
			}))
		})
	})

	context("when the build plan entry version source is from $BP_MRI_VERSION", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "mri",
				Metadata: map[string]interface{}{
					"version-source": "BP_MRI_VERSION",
					"version":        "2.6.x",
					"launch":         true,
				},
			}
		})

		it("returns a result that installs mri with BP_MRI_VERSION", func() {
			result, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Stack:   "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "mri",
							Metadata: map[string]interface{}{
								"version-source": "BP_MRI_VERSION",
								"version":        "2.6.x",
								"launch":         true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name: "mri",
						Path: filepath.Join(layersDir, "mri"),
						SharedEnv: packit.Environment{
							"GEM_PATH.override": "/some/mri/gems/path",
						},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            false,
						Launch:           true,
						Cache:            false,
						Metadata: map[string]interface{}{
							mri.DepKey: "",
							"built_at": timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "mri",
							Metadata: packit.BOMMetadata{
								Version: "mri-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "mri-dependency-sha",
								},
								URI: "mri-dependency-uri",
							},
						},
					},
				},
			}))

			Expect(filepath.Join(layersDir, "mri")).To(BeADirectory())

			Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version-source": "BP_MRI_VERSION",
						"version":        "2.6.x",
						"launch":         true,
					},
				},
			}))

			Expect(entryResolver.MergeLayerTypesCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
				{
					Name: "mri",
					Metadata: map[string]interface{}{
						"version-source": "BP_MRI_VERSION",
						"version":        "2.6.x",
						"launch":         true,
					},
				},
			}))
			Expect(entryResolver.MergeLayerTypesCall.Receives.String).To(Equal("mri"))

			Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
			Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("ruby"))
			Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("2.6.x"))
			Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{{ID: "mri", Name: "MRI"}}))

			Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{ID: "mri", Name: "MRI"}))
			Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
			Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "mri")))

			Expect(gem.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"env", "path"}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Resolving MRI version"))
			Expect(buffer.String()).To(ContainSubstring("Selected MRI version (using BP_MRI_VERSION): "))
			Expect(buffer.String()).NotTo(ContainSubstring("WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v1.0.0."))
			Expect(buffer.String()).NotTo(ContainSubstring("Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information."))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
			Expect(buffer.String()).To(ContainSubstring("Configuring build environment"))
			Expect(buffer.String()).To(ContainSubstring("Configuring launch environment"))
		})
	})

	context("when the build plan entry includes the build flag", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "mri",

				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "2.5.x",
					"build":          true,
					"launch":         true,
				},
			}

			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the mri layer as cached", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.1.2",
				},
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "mri",
							Metadata: map[string]interface{}{
								"version-source": "buildpack.yml",
								"version":        "2.5.x",
								"build":          true,
								"launch":         true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name: "mri",
						Path: filepath.Join(layersDir, "mri"),
						SharedEnv: packit.Environment{
							"GEM_PATH.override": "/some/mri/gems/path",
						},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						Metadata: map[string]interface{}{
							mri.DepKey: "",
							"built_at": timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "mri",
							Metadata: packit.BOMMetadata{
								Version: "mri-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "mri-dependency-sha",
								},
								URI: "mri-dependency-uri",
							},
						},
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
						{
							Name: "mri",
							Metadata: packit.BOMMetadata{
								Version: "mri-dependency-version",
								Checksum: packit.BOMChecksum{
									Algorithm: packit.SHA256,
									Hash:      "mri-dependency-sha",
								},
								URI: "mri-dependency-uri",
							},
						},
					},
				},
			}))
		})
	})

	context("when there is a dependency cache match", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(layersDir, "mri.toml"), []byte("[metadata]\ndependency-sha = \"some-sha\"\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				ID:     "ruby",
				Name:   "Ruby",
				SHA256: "some-sha",
			}
		})

		it("exits build process early", func() {
			_, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Stack:   "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "0.1.2",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "mri",
							Metadata: map[string]interface{}{
								"version-source": "buildpack.yml",
								"version":        "2.5.x",
								"launch":         true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{{
				ID:     "mri",
				Name:   "MRI",
				SHA256: "some-sha",
			}}))

			Expect(dependencyManager.InstallCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 0.1.2"))
			Expect(buffer.String()).To(ContainSubstring("Resolving MRI version"))
			Expect(buffer.String()).To(ContainSubstring("Selected MRI version (using buildpack.yml): "))
			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v1.0.0."))
			Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information."))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
		})
	})

	context("failure cases", func() {
		context("when a dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when a dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "0.1.2",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})

		context("when the layers directory cannot be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "0.1.2",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the MRI layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "mri", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "mri"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "mri"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "0.1.2",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})
		})

		context("when the layer directory cannot be removed", func() {
			var layerDir string
			it.Before(func() {
				layerDir = filepath.Join(layersDir, mri.MRI)
				Expect(os.MkdirAll(filepath.Join(layerDir, "baller"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(layerDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())
				Expect(os.RemoveAll(layerDir)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "0.1.2",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the executable errors", func() {
			it.Before(func() {
				gem.ExecuteCall.Stub = nil
				gem.ExecuteCall.Returns.Error = errors.New("gem executable failed")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "0.1.2",
					},
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "mri",
								Metadata: map[string]interface{}{
									"version-source": "buildpack.yml",
									"version":        "2.5.x",
								},
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("gem executable failed")))
			})
		})
	})
}

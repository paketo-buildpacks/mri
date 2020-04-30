package mri_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/postal"
	"github.com/paketo-community/mri/mri"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter mri.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = mri.NewLogEmitter(buffer)
	})

	context("SelectedDependency", func() {
		it("prints details about the selected dependency", func() {
			entry := packit.BuildpackPlanEntry{
				Metadata: map[string]interface{}{"version-source": "some-source"},
			}
			dependency := postal.Dependency{
				Name:    "MRI",
				Version: "some-version",
			}

			emitter.SelectedDependency(entry, dependency, time.Now())
			Expect(buffer.String()).To(Equal("    Selected MRI version (using some-source): some-version\n\n"))
		})

		context("when the version source is missing", func() {
			it("prints details about the selected dependency", func() {
				dependency := postal.Dependency{
					Name:    "MRI",
					Version: "some-version",
				}

				emitter.SelectedDependency(packit.BuildpackPlanEntry{}, dependency, time.Now())
				Expect(buffer.String()).To(Equal("    Selected MRI version (using <unknown>): some-version\n\n"))
			})
		})

		context("when it is within 30 days of the deprecation date", func() {
			it("returns a warning that the dependency will be deprecated after the deprecation date", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(-29 * 24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "MRI",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected MRI version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of MRI will be deprecated after 2021-04-01.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of MRI before this time.\n\n"))
			})
		})

		context("when it is on the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "MRI",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected MRI version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of MRI is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of MRI.\n\n"))
			})
		})

		context("when it is after the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "MRI",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected MRI version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of MRI is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of MRI.\n\n"))
			})
		})
	})

	context("Candidates", func() {
		it("prints a formatted map of version source inputs", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
				{
					Name:    "mri",
					Version: "package-json-version",
					Metadata: map[string]interface{}{
						"version-source": "package.json",
					},
				},
				{
					Name:    "mri",
					Version: "other-version",
				},
				{
					Name:    "mri",
					Version: "buildpack-yml-version",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
					},
				},
				{
					Name: "mri",
				},
			})

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"*\""))
		})
	})

	context("Environment", func() {
		it("prints details about the environment", func() {
			emitter.Environment(packit.Environment{
				"GEM_PATH.override": "/some/path",
			})

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    GEM_PATH -> \"/some/path\""))
		})
	})
}

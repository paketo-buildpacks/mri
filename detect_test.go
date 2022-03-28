package mri_test

import (
	"errors"
	"os"
	"testing"

	"github.com/paketo-buildpacks/mri"
	"github.com/paketo-buildpacks/mri/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackYMLParser *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		buildpackYMLParser = &fakes.VersionParser{}

		detect = mri.Detect(buildpackYMLParser)
	})

	it("returns a plan that provides mri", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: mri.MRI},
			},
		}))
	})

	context("when $BP_MRI_VERSION is set and a buildpack.yml is present", func() {
		it.Before(func() {
			os.Setenv("BP_MRI_VERSION", "1.2.3")
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.6"
		})

		it.After(func() {
			os.Unsetenv("BP_MRI_VERSION")
		})

		it("returns a plan that provides and requires the $BP_MRI_VERSION of mri", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: mri.MRI},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: mri.MRI,
						Metadata: mri.BuildPlanMetadata{
							VersionSource: "BP_MRI_VERSION",
							Version:       "1.2.3",
						},
					},
					{
						Name: mri.MRI,
						Metadata: mri.BuildPlanMetadata{
							VersionSource: "buildpack.yml",
							Version:       "4.5.6",
						},
					},
				},
			}))
		})
	})

	context("when the source code contains a buildpack.yml file and $BP_MRI_VERSION is not set", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.6"
		})

		it("returns a plan that provides and requires that version of mri", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: mri.MRI},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: mri.MRI,
						Metadata: mri.BuildPlanMetadata{
							VersionSource: "buildpack.yml",
							Version:       "4.5.6",
						},
					},
				},
			}))

			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal("/working-dir/buildpack.yml"))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml"))
			})
		})
	})
}

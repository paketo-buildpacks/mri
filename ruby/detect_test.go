package ruby_test

import (
	"errors"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/ruby-mri-cnb/ruby"
	"github.com/cloudfoundry/ruby-mri-cnb/ruby/fakes"
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

		detect = ruby.Detect(buildpackYMLParser)
	})

	it("returns a plan that provides ruby", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: ruby.Ruby},
			},
		}))
	})

	context("when the source code contains a buildpack.yml file", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "4.5.6"
		})

		it("returns a plan that provides and requires that version of ruby", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: ruby.Ruby},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:    ruby.Ruby,
						Version: "4.5.6",
						Metadata: ruby.BuildPlanMetadata{
							VersionSource: "buildpack.yml",
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

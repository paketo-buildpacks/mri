package ruby_test

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"path/filepath"
	"ruby-cnb/ruby"
	"testing"

	. "github.com/onsi/gomega"
)

func TestUnitRuby(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Ruby", testRuby, spec.Report(report.Terminal{}))
}

func testRuby(t *testing.T, when spec.G, it spec.S) {
	when("NewContributor", func() {
		var stubRubyFixture = filepath.Join("testdata", "stub-ruby.tar.gz")

		it("will contribute if the dep is in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(ruby.Dependency, buildplan.Dependency{})
			f.AddDependency(ruby.Dependency, stubRubyFixture)

			_, willContribute, err := ruby.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("will not contribute if the dep is not in the build plan", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := ruby.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("contributes for build phase", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(ruby.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true},
			})
			f.AddDependency(ruby.Dependency, stubRubyFixture)

			rubyDep, _, err := ruby.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = rubyDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(ruby.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})

		it("contributes for launch phase", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(ruby.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(ruby.Dependency, stubRubyFixture)

			rubyDep, _, err := ruby.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = rubyDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(ruby.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})
	})
}

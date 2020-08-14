package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testReusingLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack().WithNoColor()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.Build.
				WithNoPull().
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				)

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				"Paketo MRI Buildpack 1.2.3",
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"2.7.x\"",
				"      <unknown>     -> \"*\"",
				"",
				MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.7\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.\d+\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/paketo-community_mri/mri/lib/ruby/gems/2\.7\.\d+"`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				"Paketo MRI Buildpack 1.2.3",
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"2.7.x\"",
				"      <unknown>     -> \"*\"",
				"",
				MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.7\.\d+`),
				"",
				"  Reusing cached layer /layers/paketo-community_mri/mri",
			))

			secondContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(MatchRegexp(`Hello from Ruby 2\.7\.\d+`))

			Expect(secondImage.Buildpacks[0].Layers["mri"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["mri"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.Build.
				WithNoPull().
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				)

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				"Paketo MRI Buildpack 1.2.3",
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"2.7.x\"",
				"      <unknown>     -> \"*\"",
				"",
				MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.7\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.7\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/paketo-community_mri/mri/lib/ruby/gems/2\.7\.\d+"`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			Expect(ioutil.WriteFile(filepath.Join(source, "buildpack.yml"), []byte("---\nmri:\n  version: 2.6.x\n"), 0644)).To(Succeed())

			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/mri"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("mri"))

			Expect(logs).To(ContainLines(
				"Paketo MRI Buildpack 1.2.3",
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"2.6.x\"",
				"      <unknown>     -> \"*\"",
				"",
				MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.6\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.6\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring environment",
				MatchRegexp(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.6\.\d+:/layers/paketo-community_mri/mri/lib/ruby/gems/2\.6\.\d+"`),
			))

			secondContainer, err = docker.Container.Run.WithCommand("ruby run.rb").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(MatchRegexp(`Hello from Ruby 2\.6\.\d+`))

			Expect(secondImage.Buildpacks[0].Layers["mri"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["mri"].Metadata["built_at"]))
		})
	})
}

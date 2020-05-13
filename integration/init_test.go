package integration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/dagger"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	mriBuildpack        string
	offlineMRIBuildpack string
	buildPlanBuildpack  string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())

	mriBuildpack, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	offlineMRIBuildpack, _, err = dagger.PackageCachedBuildpack(root)
	Expect(err).NotTo(HaveOccurred())

	buildPlanBuildpack, err = dagger.GetLatestCommunityBuildpack("ForestEckhardt", "build-plan")
	Expect(err).NotTo(HaveOccurred())

	// HACK: we need to fix dagger and the package.sh scripts so that this isn't required
	mriBuildpack = fmt.Sprintf("%s.tgz", mriBuildpack)
	offlineMRIBuildpack = fmt.Sprintf("%s.tgz", offlineMRIBuildpack)

	defer func() {
		dagger.DeleteBuildpack(mriBuildpack)
		dagger.DeleteBuildpack(offlineMRIBuildpack)
		dagger.DeleteBuildpack(buildPlanBuildpack)
	}()

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Logging", testLogging)
	suite("Offline", testOffline)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)
	suite("SimpleApp", testSimpleApp)
	suite.Run(t)
}

func ContainerLogs(id string) func() string {
	docker := occam.NewDocker()

	return func() string {
		logs, _ := docker.Container.Logs.Execute(id)
		return logs.String()
	}
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})
	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	err = gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--tags", strings.TrimSpace(revListOut.String())},
		Stdout: stdout,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(stdout.String(), "v")), nil
}

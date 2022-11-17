package components

import (
	"io"
	"net/http"

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
)

type RubyBranch struct {
	Name    string `yaml:"name"`
	EolDate string `yaml:"eol_date"`
}

type DeprecationDateRetriever struct{}

func NewDeprecationDateRetriever() DeprecationDateRetriever {
	return DeprecationDateRetriever{}
}

// GetDeprecationDate will look up if a version has an EOL date, and return it
// if there is one, and return "" if there is not one.
func (DeprecationDateRetriever) GetDate(feed, version string) (string, error) {
	resp, err := http.Get(feed) // nolint
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// untested
		return "", err
	}

	var rubyBranches []RubyBranch
	err = yaml.Unmarshal(body, &rubyBranches)
	if err != nil {
		return "", err
	}

	for _, branch := range rubyBranches {
		branchVersion := semver.MustParse(branch.Name)
		branchMajor := branchVersion.Major()
		branchMinor := branchVersion.Minor()

		sVersion := semver.MustParse(version)
		sMajor := sVersion.Major()
		sMinor := sVersion.Minor()

		if sMajor == branchMajor && sMinor == branchMinor {
			return branch.EolDate, nil
		}
	}
	return "", nil
}

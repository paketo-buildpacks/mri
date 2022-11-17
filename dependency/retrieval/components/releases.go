package components

import (
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"
)

type URL struct {
	Gz string `yaml:"gz"`
}

type SHA256 struct {
	Gz string `yaml:"gz"`
}

type RubyRelease struct {
	Version string `yaml:"version"`
	URL     URL    `yaml:"url"`
	SHA256  SHA256 `yaml:"sha256"`
}

type ReleaseFetcher struct {
	releaseIndex string
}

func NewReleaseFetcher(feed string) ReleaseFetcher {
	return ReleaseFetcher{
		releaseIndex: feed,
	}
}

// GetUpstreamVersions will take in a ReleaseFetcher version feed, parse the content, and
// return all available versions in the form of a map, where the key is the
// version (string) and the value is a struct containing obtained version
// metadata.
func (rf ReleaseFetcher) GetUpstreamReleases() (map[string]RubyRelease, error) {
	versions := make(map[string]RubyRelease)

	resp, err := http.Get(rf.releaseIndex) // nolint
	if err != nil {
		return versions, err
	}
	if resp.StatusCode != http.StatusOK {
		return versions, fmt.Errorf("failed to query %s: %d", rf.releaseIndex, resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return versions, err
	}

	var rubyVersionStructs []RubyRelease
	err = yaml.Unmarshal(body, &rubyVersionStructs)
	if err != nil {
		return versions, err
	}

	for _, v := range rubyVersionStructs {
		versions[v.Version] = v
	}
	return versions, nil
}

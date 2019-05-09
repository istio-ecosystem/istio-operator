package manifest

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"istio.io/istio/pkg/log"
)

type Repository interface {
	LoadChannel(ctx context.Context, name string) (*Channel, error)
	LoadManifest(ctx context.Context, packageName string, id string) (string, error)
}

// FSRepository is a Repository backed by a filesystem
type FSRepository struct {
	basedir string
}

var _ Repository = &FSRepository{}

// NewFSRepository is the constructor for an FSRepository
func NewFSRepository(basedir string) *FSRepository {
	return &FSRepository{
		basedir: basedir,
	}
}

var safelistChannelName = "abcdefghijklmnopqrstuvwxyz"

// We validate the channel name - keeping it to a small subset helps with path traversal,
// and also ensures that we can back easily this by other stores (e.g. https)
func allowedChannelName(name string) bool {
	if !matchesSafelist(name, safelistChannelName) {
		return false
	}

	// Double check!
	if strings.HasPrefix(name, ".") {
		return false
	}

	return true
}

var safelistVersion = "abcdefghijklmnopqrstuvwxyz0123456789-."

func allowedManifestId(name string) bool {
	if !matchesSafelist(name, safelistVersion) {
		return false
	}

	// Double check!
	if strings.HasPrefix(name, ".") {
		return false
	}

	return true
}

func matchesSafelist(s string, safelist string) bool {
	for _, c := range s {
		if strings.IndexRune(safelist, c) == -1 {
			return false
		}
	}
	return true
}

func (r *FSRepository) LoadManifest(ctx context.Context, packageName string, id string) (string, error) {
	if !allowedManifestId(packageName) {
		return "", fmt.Errorf("invalid package name: %q", id)
	}

	if !allowedManifestId(id) {
		return "", fmt.Errorf("invalid manifest id: %q", id)
	}

	log.Infof("loading package %s", packageName)

	p := filepath.Join(r.basedir, "packages", packageName, id, "manifest.yaml")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("error reading package %s: %v", p, err)
	}

	return string(b), nil
}
func (r *FSRepository) LoadChannel(ctx context.Context, name string) (*Channel, error) {
	return nil, fmt.Errorf("LoadChannel not implemented")
}

type Channel struct {
	Manifests []Version `json:"manifests,omitempty"`
}

type Version struct {
	Version string
}

func (c *Channel) Latest() (*Version, error) {
	var latest *Version
	for i := range c.Manifests {
		v := &c.Manifests[i]
		if latest == nil {
			latest = v
		} else {
			return nil, fmt.Errorf("version selection not implemented")
		}
	}

	return latest, nil
}

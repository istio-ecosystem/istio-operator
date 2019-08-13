// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"

	"istio.io/operator/pkg/httprequest"
	"istio.io/operator/pkg/util"
	"istio.io/pkg/log"
)

const (
	// InstallationChartsFileName is the name of the installation package to fetch.
	InstallationChartsFileName = "istio-installer-1.3.0.tar.gz"
	// InstallationShaFileName is Sha filename to verify
	InstallationShaFileName = "istio-installer-1.3.0.tar.gz.sha256"
	// ChartsTempFilePrefix is temporary Files prefix
	ChartsTempFilePrefix = "istio-install-package"
)

// FileDownloader is wrapper of HTTP client to download files
type FileDownloader struct {
	// client is a HTTP/HTTPS client.
	client *http.Client
}

// URLFetcher is used to fetch and manipulate charts from remote url
type URLFetcher struct {
	// url is url to download the charts
	url string
	// verifyURL is url to download the verification file
	verifyURL string
	// verify indicates whether the downloaded tar should be verified
	verify bool
	// destDir is path of charts downloaded to, empty as default to temp dir
	destDir string
}

// NewURLFetcher creates an URLFetcher pointing to urls and destination
func NewURLFetcher(repoURL string, destDir string, cFileName string, shaFileName string) (*URLFetcher, error) {
	if destDir == "" {
		destDir = filepath.Join(os.TempDir(), ChartsTempFilePrefix)
	}
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err := os.Mkdir(destDir, os.ModeDir|os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	uf := &URLFetcher{
		url:        filepath.Join(repoURL, cFileName),
		verifyURL:  filepath.Join(repoURL, shaFileName),
		verify:     true,
		destDir:    destDir,
	}
	return uf, nil
}

// GetDestDir returns path of destination dir.
func (f *URLFetcher) GetDestDir() string {
	return f.destDir
}

// FetchBundles fetches the charts, sha and version file
func (f *URLFetcher) FetchBundles() util.Errors {
	errs := util.Errors{}

	shaF, err := f.fetchSha()
	errs = util.AppendErr(errs, err)

	return util.AppendErr(errs, f.fetchChart(shaF))
}

// fetchChart fetches the charts and verifies charts against SHA file if required
func (f *URLFetcher) fetchChart(shaF string) error {
	saved, err := DownloadTo(f.url, f.destDir)
	if err != nil {
		return err
	}
	file, err := os.Open(saved)
	if err != nil {
		return err
	}
	defer file.Close()
	if f.verify {
		// verify with sha file
		_, err := os.Stat(shaF)
		if os.IsNotExist(err) {
			shaF, err = f.fetchSha()
			if err != nil {
				return fmt.Errorf("failed to get sha file: %s", err)
			}
		}
		hashAll, err := ioutil.ReadFile(shaF)
		if err != nil {
			return fmt.Errorf("failed to read sha file: %s", err)
		}
		// SHA file has structure of "sha_value filename"
		hash := strings.Split(string(hashAll), " ")[0]
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			log.Error(err.Error())
		}
		sum := h.Sum(nil)
		actualHash := hex.EncodeToString(sum)
		if !strings.EqualFold(actualHash, hash) {
			return fmt.Errorf("checksum of charts file located at: %s does not match expected SHA file: %s", saved, shaF)
		}
	}
	targz := archiver.TarGz{Tar: &archiver.Tar{OverwriteExisting: true}}
	return targz.Unarchive(saved, f.destDir)
}

// fetchsha downloads the SHA file from url
func (f *URLFetcher) fetchSha() (string, error) {
	if f.verifyURL == "" {
		return "", fmt.Errorf("SHA file url is empty")
	}
	shaF, err := DownloadTo(f.verifyURL, f.destDir)
	if err != nil {
		return "", err
	}
	return shaF, nil
}

// DownloadTo downloads from remote url to dest local file path
func DownloadTo(ref, dest string) (string, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid chart URL: %s", ref)
	}
	data, err := httprequest.Get(u.String())
	if err != nil {
		return "", err
	}

	name := filepath.Base(u.Path)
	destFile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destFile, data, 0666); err != nil {
		return destFile, err
	}

	return destFile, nil
}

package httpget

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Get sends an HTTP GET request and returns the result.
func Get(url string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			Proxy:              http.ProxyFromEnvironment,
		}}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch URL %s : %s", url, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	return buf.Bytes(), err
}

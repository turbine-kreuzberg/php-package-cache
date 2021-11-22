package ppc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
)

type Metadata struct {
	Packages map[string][]struct {
		VersionNormalized string `json:"version_normalized"`
		Dist              struct {
			Reference string `json:"reference"`
			Type_     string `json:"type"`
			URL       string `json:"url"`
		} `json:"dist"`
	}
}

func url(ctx context.Context, package_, version, reference, type_ string) (string, error) {
	metadata, err := loadMetadata(ctx, package_)
	if err != nil {
		return "", fmt.Errorf("load metadata for package %s: %v", package_, err)
	}

	pkg, found := metadata.Packages[package_]
	if !found {
		return "", fmt.Errorf("package %s not found in metadata", package_)
	}

	for _, pkgVersion := range pkg {
		if pkgVersion.VersionNormalized != version {
			continue
		}

		if pkgVersion.Dist.Type_ != type_ {
			continue
		}

		if pkgVersion.Dist.Reference != reference {
			return "", fmt.Errorf("reference %s does not match found reference %s for version %s of package %s", reference, pkgVersion.Dist.Reference, version, package_)
		}

		return pkgVersion.Dist.URL, nil
	}

	return "", fmt.Errorf("version %s with type %s not for for package %s", version, type_, package_)
}

func loadMetadata(ctx context.Context, name string) (*Metadata, error) {
	bytes, err := fetchMetadata(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata: %v", err)
	}

	metadata := &Metadata{}
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %v", err)
	}

	return metadata, nil
}

func fetchMetadata(ctx context.Context, name string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fetch metadata")
	defer span.Finish()

	url := fmt.Sprintf("https://packagist.org/p2/%s.json", name)
	client := httpClient()

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status code %d", resp.StatusCode)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading result: %v", err)
	}

	return result, nil
}

func httpClient() *http.Client {
	var transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
	return client
}

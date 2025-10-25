package qiniu

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"golang.org/x/mod/semver"
)

type Client struct {
	mac    *qbox.Mac
	bucket string
}

func NewClient(accessKey, secretKey, bucket string) *Client {
	mac := qbox.NewMac(accessKey, secretKey)
	return &Client{
		mac:    mac,
		bucket: bucket,
	}
}

func (c *Client) GetAppVersions(ctx context.Context, appName string) ([]string, error) {
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(c.mac, &cfg)

	prefix := fmt.Sprintf("niulink-materials/%s/", appName)
	delimiter := ""
	marker := ""
	limit := 1000

	var versions []string
	versionSet := make(map[string]bool)

	for {
		entries, _, nextMarker, _, err := bucketManager.ListFiles(c.bucket, prefix, delimiter, marker, limit)
		if err != nil {
			return nil, fmt.Errorf("列举文件失败: %v", err)
		}

		for _, entry := range entries {
			fileName := strings.TrimPrefix(entry.Key, prefix)
			
			if strings.HasSuffix(fileName, ".tar.gz") {
				parts := strings.Split(fileName, "_")
				if len(parts) > 0 {
					version := parts[0]
					if version != "" && !versionSet[version] {
						versions = append(versions, version)
						versionSet[version] = true
					}
				}
			}
		}

		if nextMarker == "" {
			break
		}
		marker = nextMarker
	}

	sortVersions(versions)

	return versions, nil
}

func sortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		vi := versions[i]
		vj := versions[j]

		if !strings.HasPrefix(vi, "v") {
			vi = "v" + vi
		}
		if !strings.HasPrefix(vj, "v") {
			vj = "v" + vj
		}

		if semver.IsValid(vi) && semver.IsValid(vj) {
			return semver.Compare(vi, vj) > 0
		}

		return versions[i] > versions[j]
	})
}

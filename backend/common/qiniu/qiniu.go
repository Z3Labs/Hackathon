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
	mac          *qbox.Mac
	bucket       string
	downLoadHost string
}

func NewClient(accessKey, secretKey, bucket string, downloadHost string) *Client {
	mac := qbox.NewMac(accessKey, secretKey)
	return &Client{
		mac:          mac,
		bucket:       bucket,
		downLoadHost: downloadHost,
	}
}

type AppVersion struct {
	Version  string `json:"version"`   // 版本号
	FileName string `json:"file_name"` // 完整文件名
}

func (c *Client) GetAppVersions(ctx context.Context, appName string) ([]AppVersion, error) {
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(c.mac, &cfg)

	prefix := fmt.Sprintf("%s/", appName)
	delimiter := ""
	marker := ""
	limit := 1000

	var versions []AppVersion
	versionSet := make(map[string]bool)

	for {
		entries, _, nextMarker, _, err := bucketManager.ListFiles(c.bucket, prefix, delimiter, marker, limit)
		if err != nil {
			return nil, fmt.Errorf("列举文件失败: %v", err)
		}

		for _, entry := range entries {
			fileName := strings.TrimPrefix(entry.Key, prefix)

			if strings.HasSuffix(fileName, ".tar.gz") {
				fileName = strings.TrimSuffix(fileName, ".tar.gz")
				parts := strings.Split(fileName, "_")
				if len(parts) > 0 {
					version := parts[0]
					if version != "" && !versionSet[version] {
						versions = append(versions, AppVersion{
							Version:  version,
							FileName: fileName,
						})
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

func sortVersions(versions []AppVersion) {
	sort.Slice(versions, func(i, j int) bool {
		vi := versions[i].Version
		vj := versions[j].Version

		if !strings.HasPrefix(vi, "v") {
			vi = "v" + vi
		}
		if !strings.HasPrefix(vj, "v") {
			vj = "v" + vj
		}

		if semver.IsValid(vi) && semver.IsValid(vj) {
			return semver.Compare(vi, vj) > 0
		}

		return versions[i].Version > versions[j].Version
	})
}

func (c *Client) GetFileStat(ctx context.Context, fileName string) (storage.FileInfo, error) {

	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(c.mac, &cfg)

	return bucketManager.Stat(c.bucket, fileName)
}

func (c *Client) GetFileURL(ctx context.Context, fileName string, deadline int64) string {
	return storage.MakePrivateURLv2(c.mac, c.downLoadHost, fileName, deadline)
}

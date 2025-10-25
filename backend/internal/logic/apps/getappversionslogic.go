package apps

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppVersionsLogic {
	return &GetAppVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppVersionsLogic) GetAppVersions(req *types.GetAppVersionsReq) (resp *types.GetAppVersionsResp, err error) {
	if req.AppName == "" {
		return nil, errors.New("应用名称不能为空")
	}

	mac := auth.New(l.svcCtx.Config.Qiniu.AccessKey, l.svcCtx.Config.Qiniu.SecretKey)
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)

	prefix := fmt.Sprintf("niulink-materials/%s/", req.AppName)
	delimiter := ""
	marker := ""
	limit := 1000

	var allFiles []string
	for {
		entries, _, nextMarker, _, err := bucketManager.ListFiles(
			l.svcCtx.Config.Qiniu.Bucket,
			prefix,
			delimiter,
			marker,
			limit,
		)
		if err != nil {
			l.Errorf("[GetAppVersions] ListFiles error:%v", err)
			return nil, errors.New("获取版本列表失败")
		}

		for _, entry := range entries {
			allFiles = append(allFiles, entry.Key)
		}

		if nextMarker == "" {
			break
		}
		marker = nextMarker
	}

	versionPattern := regexp.MustCompile(`^niulink-materials/[^/]+/(v[\d.]+)_.*\.tar\.gz$`)
	versionMap := make(map[string]bool)

	for _, file := range allFiles {
		matches := versionPattern.FindStringSubmatch(file)
		if len(matches) == 2 {
			version := matches[1]
			versionMap[version] = true
		}
	}

	var versions []string
	for version := range versionMap {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	l.Infof("[GetAppVersions] Successfully retrieved versions for app: %s, count: %d", req.AppName, len(versions))

	return &types.GetAppVersionsResp{
		Versions: versions,
	}, nil
}

func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &n1)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &n2)
		}

		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}

	return 0
}

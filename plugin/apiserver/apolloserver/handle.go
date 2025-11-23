package apolloserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/config"
)

var (
	_validFileExt = map[string]struct{}{
		string(conftypes.FileFormatProperties): {},
		string(conftypes.FileFormatYaml):       {},
		string(conftypes.FileFormatJson):       {},
		string(conftypes.FileFormatText):       {},
		string(conftypes.FileFormatXml):        {},
	}
)

// GetConfigFile 获取配置文件
// TODO 当前还不支持 apollo 的公共配置继承，需要低层的配置中心能力能够支持即可
func (a *ApolloServer) GetConfigFile(ctx context.Context, req *GetConfigFileRequest) (*GetConfigFileResponse, error) {
	// apollo 获取配置的顺序，custom point cluster -> dataCenter -> default cluster
	nsArgs := []string{req.Cluster, req.DataCenter, "default"}
	if req.Cluster == "default" || req.Cluster == "" {
		nsArgs = []string{req.DataCenter, "default"}
	}

	filenames := []string{req.Filename}
	fext := config.ResolveFileType(req.Filename)
	if _, ok := _validFileExt[fext]; !ok {
		// 如果文件名没有后缀名，尝试添加默认的后缀名
		filenames = append(filenames, req.Filename+"."+string(conftypes.FileFormatProperties))
	} else if fext == string(conftypes.FileFormatProperties) {
		filenames = append(filenames, strings.TrimSuffix(req.Filename, "."+string(conftypes.FileFormatProperties)))
	}

	for i := range nsArgs {
		for _, filename := range filenames {
			rsp := a.configSvr.GetConfigFileWithCache(ctx, &apiconfig.ConfigFile{
				Namespace: nsArgs[i],
				Group:     req.AppId,
				Name:      filename,
			})

			if api.IsSuccess(rsp) {
				if rsp.GetCode() == uint32(apimodel.Code_DataNoChange) {
					// 数据没变更
					return nil, nil
				}

				cfg, err := convertToConfigurations(rsp.GetFile())
				if err != nil {
					return nil, err
				}

				file := rsp.GetFile()
				releaseKey := fmt.Sprintf("%s-%s-%s-%d", file.Namespace, file.Group, file.FileName, file.Version)

				return &GetConfigFileResponse{
					AppId:          req.AppId,
					Cluster:        req.Cluster,
					NamespaceName:  req.Filename,
					ReleaseKey:     releaseKey,
					Configurations: cfg,
				}, nil
			}
		}
	}

	// 配置没有找到，返回 404
	return nil, Err_NotFoundConfig
}

func convertToConfigurations(f *apiconfig.ConfigFileRelease) (map[string]string, error) {
	items := map[string]string{}
	content := f.GetContent()
	fext := path.Ext(f.GetFileName())
	switch fext {
	case "", "properties":
		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || line[0] == '#' || line[0] == '!' {
				continue
			}
			sepIdx := -1
			for i, c := range line {
				if c == '=' || c == ':' {
					sepIdx = i
					break
				}
			}
			if sepIdx < 0 {
				continue
			}
			key := strings.TrimSpace(line[:sepIdx])
			value := strings.TrimSpace(line[sepIdx+1:])
			items[key] = value
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return items, nil
	default:
		// 其他类型的配置文件，直接返回内容
		items = map[string]string{
			"content": content,
		}
	}
	return items, nil
}

func (a *ApolloServer) WatchConfigFile(ctx context.Context, req *WatchConfigFileRequest) (*WatchConfigFileResponse, error) {
	nsName := func() string {
		if req.Cluster != "" && req.Cluster != "default" {
			return req.Cluster
		}
		if req.DataCenter != "" {
			return req.DataCenter
		}
		return "default"
	}()

	clientSideNotifications := []*apiconfig.ConfigFile{}
	clientWatchFiles := []*apiconfig.ConfigFileRelease{}
	for _, item := range req.Notifications {
		// For AddWatcher
		clientSideNotifications = append(clientSideNotifications, &apiconfig.ConfigFile{
			Namespace: nsName,
			Group:     req.AppId,
			Name:      item.NamespaceName,
			Tags: map[string]string{
				types.ClientLabel_IP: req.ClientIP,
			},
		})

		// For ClientWatchConfigFileRequest
		clientWatchFiles = append(clientWatchFiles, &apiconfig.ConfigFileRelease{
			Namespace: nsName,
			Group:     req.AppId,
			FileName:  item.NamespaceName,
			Version:   uint64(item.NotificationId),
			Labels: map[string]string{
				types.ClientLabel_IP: req.ClientIP,
			},
		})
	}

	specReq := &apiconfig.ClientWatchConfigFileRequest{
		ClientIp: req.ClientIP,
		Files:    clientWatchFiles,
	}

	if val := a.diffChangeFiles(ctx, specReq); len(val) > 0 {
		// 有变更
		return &WatchConfigFileResponse{
			Notifications: val,
		}, nil
	}

	clientId := utils.ParseClientAddress(ctx) + "@" + utils.NewUUID()[0:8]
	// 没有变更，加入到长轮询中
	watchCtx := a.innerSvr.WatchCenter().AddWatcher(clientId, clientSideNotifications, a.BuildTimeoutWatchCtx(ctx, a.watchTimeOut))
	notifyRet := (watchCtx.(*ApolloWatchContext)).GetNotifieResult()
	notifyCode := notifyRet.Code
	if notifyCode != uint32(apimodel.Code_ExecuteSuccess) && notifyCode != uint32(apimodel.Code_DataNoChange) {
		apollolog.Errorf("watch config file failed: %s, code: %d", notifyRet.Info, notifyCode)
		return nil, errors.New("watch config file failed: " + notifyRet.Info)
	}

	var changeKeys []*ApolloConfigNotification
	if notifyCode == uint32(apimodel.Code_DataNoChange) {
		// 如果没有变化，再进行一次 diff 判断是否存在配置变更
		changeKeys = a.diffChangeFiles(ctx, specReq)
	} else {
		// 如果收到一个事件变化，就立即通知这个文件的变化信息
		// 使用原始请求中的文件信息生成通知
		changeKeys = make([]*ApolloConfigNotification, 0, len(req.Notifications))
		for _, item := range req.Notifications {
			// 获取当前活跃的配置文件版本
			active := a.innerSvr.CacheManager().ConfigFile().GetActiveRelease(nsName, req.AppId, item.NamespaceName)
			notificationId := item.NotificationId
			if active != nil {
				notificationId = int64(active.Version)
			}

			changeKeys = append(changeKeys, &ApolloConfigNotification{
				NamespaceName:  item.NamespaceName,
				NotificationId: notificationId,
			})
		}
	}
	return &WatchConfigFileResponse{Notifications: changeKeys}, nil
}

// GetConfigServers 获取配置中心的接入点，二次寻址发现
func (a *ApolloServer) GetConfigServers(ctx context.Context, appId, clientIp string) ([]*ServerNode, error) {
	if a.metaSvrs == nil {
		return []*ServerNode{
			{
				AppName:     a.GetProtocol(),
				InstanceId:  fmt.Sprintf("http://%s:%d/", utils.LocalHost, a.listenPort),
				HomepageUrl: fmt.Sprintf("http://%s:%d/", utils.LocalHost, a.listenPort),
			},
		}, nil
	}

	rsp := a.discoverSvr.ServiceInstancesCache(ctx, &service_manage.DiscoverFilter{
		OnlyHealthyInstance: true,
	}, &service_manage.Service{
		Namespace: a.metaSvrs.Namespace,
		Name:      a.metaSvrs.Name,
	})

	if !api.IsSuccess(rsp) {
		return nil, nil
	}

	nodes := make([]*ServerNode, 0, len(rsp.GetInstances()))
	for _, ins := range rsp.GetInstances() {
		nodes = append(nodes, &ServerNode{
			AppName:     ins.GetService(),
			InstanceId:  ins.GetId(),
			HomepageUrl: fmt.Sprintf("http://%s:%d/", ins.GetHost(), ins.GetPort()),
		})
	}
	return nodes, nil
}

# 对Store的API的mock

## mock文件生成方法

在`./store`目录执行

```
mockgen -source=api.go -aux_files github.com/pole-io/pole-server/pkg/store=config_file_api.go,github.com/pole-io/pole-server/pkg/store=discover_api.go,github.com/pole-io/pole-server/pkg/store=auth_api.go,github.com/pole-io/pole-server/pkg/store=admin_api.go -destination=mock/api_mock.go -package=mock

mockgen -source=mysql/admin.go -destination=mock/admin_mock.go -package=mock
```
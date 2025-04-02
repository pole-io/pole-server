# pole-io

更为云原生、AI Native 的服务治理平台

第一个二开版本预期支持的新特性

- [ ] 支持 MCP 协议，打通大模型与 pole，
- [ ] 兼容 consul、apollo、sofa 协议接入
- [x] 更好的服务注册发现性能以及内存控制
- [ ] 支持服务实例的主动探测
- [ ] 更强的配置中心，支持多环境版本渲染，降低用户的配置维护理解成本
- [ ] 所有治理规则均支持版本控制以及灰度下发
- [ ] 所有版本 Apollo/Nacos/Polaris 客户端都将支持标签灰度下发能力
- [x] 所有资源均支持细粒度鉴权，对标云厂商的 CAM/RAM 等能力


**感谢 Tencent 开源的 polarismesh，本项目 fork 自 polarismesh 并二开定制**
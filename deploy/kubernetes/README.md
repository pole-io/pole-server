# Pole 本地 Kubernetes 部署

本目录把 Pole Control Plane、GreptimeDB 和 OpenTelemetry Collector 编排到同一个 `pole-system` namespace，但保持为三个独立工作负载。MySQL 不进入 Kubernetes，Pod 通过 `pole-mysql` ExternalName Service 访问宿主机 `host.docker.internal:3306`。Console 复用 `tidemind/tidemind-gateway`，通过跨 namespace `HTTPRoute` 暴露域名。

## 部署

前置条件：

- 当前 `kubectl` context 指向本地 Kubernetes，例如 OrbStack。
- 宿主机 MySQL `3306` 可从 Pod 访问，并已创建 `pole_server`、`pole_observability` 数据库。
- 如果 MySQL 由 Docker 容器提供，应配置 `--restart unless-stopped`；既有容器可执行 `docker update --restart unless-stopped pole-mysql`，避免 OrbStack/Docker 重启后 Control Plane 因数据库不可达持续 CrashLoop。
- 已安装 Docker、Go、Node.js 和 npm。
- 集群中已有 `tidemind/tidemind-gateway`，其 `http` listener 允许同 namespace 的 Route。

```bash
./deploy/kubernetes/build-image.sh

MYSQL_USER=root MYSQL_PWD=123456 \
POLE_CONSOLE_JWT_SECRET='replace-with-local-console-secret' \
POLE_AUTH_SALT='1234567890123456' \
  ./deploy/kubernetes/deploy-local.sh
```

密码、Console JWT Secret 和认证 Salt 通过 `pole-runtime-secrets` Secret 注入，不写入 manifests。可使用环境变量覆盖：

```bash
MYSQL_USER=root \
MYSQL_PWD='your-password' \
POLE_CONSOLE_JWT_SECRET='your-console-secret' \
POLE_AUTH_SALT='16-or-24-or-32-byte-salt' \
./deploy/kubernetes/deploy-local.sh
```

## 访问与验证

本地部署完成后优先使用域名访问：

```text
http://pole.localhost/
http://pole.localhost/agent
```

`.localhost` 由浏览器和系统解析到本机，不需要修改 `/etc/hosts`。`HTTPRoute` 位于 `tidemind` namespace；`pole-system` 中的 `ReferenceGrant` 按来源 namespace 和目标 Service 收敛授权，只允许 `tidemind` 的 HTTPRoute 引用 `pole-control-plane:8080`，不会暴露 GreptimeDB、Collector 或 MySQL。

```bash
kubectl -n pole-system get pods,svc,pvc
kubectl -n tidemind get httproute pole-console
kubectl -n pole-system logs deployment/pole-control-plane --tail=100
kubectl -n pole-system logs deployment/pole-otel-collector --tail=100
```

如果 Console 入口拒绝连接且 `pole-control-plane` 为 `CrashLoopBackOff`，先检查前一容器日志和宿主机 MySQL：

```bash
kubectl -n pole-system logs deployment/pole-control-plane --previous --tail=100
docker inspect pole-mysql --format 'status={{.State.Status}} restart={{.HostConfig.RestartPolicy.Name}}'
docker start pole-mysql
docker update --restart unless-stopped pole-mysql
kubectl -n pole-system rollout restart deployment/pole-control-plane
kubectl -n pole-system rollout status deployment/pole-control-plane --timeout=240s
```

如需绕过 Gateway 排障或从宿主机调试 Pole 的 gRPC、Nacos、Apollo、Eureka、xDS 等协议，OrbStack 仍会为 `LoadBalancer` Service 分配本机集群地址。该直连入口是本地开发边界；生产部署应按协议暴露需求拆分或收敛 Service：

```bash
POLE_IP=$(kubectl -n pole-system get service pole-control-plane -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -fsS "http://${POLE_IP}:8080/"
curl -fsS "http://${POLE_IP}:8080/agent"
```

如果本地 Kubernetes 没有 LoadBalancer 实现，可以使用：

```bash
kubectl -n pole-system port-forward service/pole-control-plane 18080:8080 18090:8090
```

## 数据边界

- GreptimeDB 数据写入 `data-pole-greptimedb-0` PVC。
- Pole 本地 OTel 可靠队列和日志使用 `emptyDir`，Pod 重建后不保留；生产部署应替换为 PVC 或外部队列。
- 宿主机 MySQL 是部署外部依赖，删除 namespace 不会删除 MySQL 数据。
- 旧 Docker Compose 栈不会被部署脚本自动删除；完成数据迁移和验证后再停掉，避免误删原有 GreptimeDB volume。

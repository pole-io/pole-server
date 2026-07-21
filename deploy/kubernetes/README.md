# Pole 本地 Kubernetes 部署

本目录把 Pole Control Plane、GreptimeDB 和 OpenTelemetry Collector 编排到同一个 `pole-system` namespace，但保持为三个独立工作负载。MySQL 不进入 Kubernetes，Pod 通过 `pole-mysql` ExternalName Service 访问宿主机 `host.docker.internal:3306`。

## 部署

前置条件：

- 当前 `kubectl` context 指向本地 Kubernetes，例如 OrbStack。
- 宿主机 MySQL `3306` 可从 Pod 访问，并已创建 `pole_server`、`pole_observability` 数据库。
- 已安装 Docker、Go、Node.js 和 npm。

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

## 验证

```bash
kubectl -n pole-system get pods,svc,pvc
kubectl -n pole-system logs deployment/pole-control-plane --tail=100
kubectl -n pole-system logs deployment/pole-otel-collector --tail=100
```

OrbStack 会为 `LoadBalancer` Service 分配本机集群地址：

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

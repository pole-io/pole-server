package apolloserver

// Apollo 配置模型和 Pole 的配置模型的映射

// Apollo 配置模型层级
// 应用（AppId）
// ├── 环境（Environment）
// │   ├── 集群（Cluster）
// │   │   ├── 命名空间（Namespace）
// │   │   │   ├── 配置项（Item1, Item2, ...）
// │   │   │   └── 配置项（Item3, Item4, ...）
// │   │   └── 命名空间（Namespace2）
// │   └── 集群（Cluster2）
// └── 环境（Environment2）

// Pole 配置模型层级
// 命名空间（Namespace）
// ├── 分组（Group1）
// │   ├── 配置文件（File1.env=dev）  # 开发环境配置
// │   ├── 配置文件（File1.env=prod） # 生产环境配置
// │   └── 配置文件（File2）          # 通用配置（不区分环境）
// └── 分组（Group2）
//     └── ...

// Apollo 配置模型到 Pole 配置模型的映射关系
// 1. Apollo 的 Cluster or DataCenter 映射到 Pole 的 Namespace
// 2. Apollo 的 AppId 映射到 Pole 的 Group
// 3. Apollo 的 Namespace 映射到 Pole 的 ConfigFile

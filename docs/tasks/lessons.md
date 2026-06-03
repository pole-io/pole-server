# Lessons

- 当讨论 pole-control-plane 与 pole-console 的启动/部署标准化时，默认启动语义应是合并启动；参数只用于切换单独启动 `server` 或单独启动 `console`，不要把默认设计成只启动 control-plane 后再通过 `--console` 打开控制台。
- “合并启动”在 pole-control-plane / pole-console 场景里应优先理解为代码与运行时合并进 control-plane，而不是由 control-plane 再拉起一个 pole-console 外部子进程；除非用户明确要求双进程 runner，否则不要把子进程方案当成默认实现。
- 合并 console 后，默认产品入口端口优先给 console 使用 `8080`；如果 Apollo 协议端口与 console 冲突，应调整 Apollo 端口，不要挤占 console 默认入口。

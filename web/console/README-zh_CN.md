# Lattice.Hub Console

Pole Control Plane 的管理控制台，基于 React、Vite 和 Fluent UI React v9 构建。

## 本地开发

```bash
npm install --legacy-peer-deps
npm run dev
```

## 验证

```bash
npm run lint
npm run build:test
node scripts/verify-no-tdesign.mjs
```

控制台不得依赖或加载 TDesign React；公共控件统一维护在 `src/components/Fluent/`。

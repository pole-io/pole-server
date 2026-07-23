# Lattice.Hub Console

The management console for Pole Control Plane, built with React, Vite, and Fluent UI React v9.

## Development

```bash
npm install --legacy-peer-deps
npm run dev
```

## Verification

```bash
npm run lint
npm run build:test
node scripts/verify-no-tdesign.mjs
```

The console must not depend on or load TDesign React. Shared controls live under `src/components/Fluent/`.

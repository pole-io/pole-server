import path from 'path';
import { loadEnv } from 'vite';
import { viteMockServe } from 'vite-plugin-mock';
import react from '@vitejs/plugin-react';
import svgr from '@honkhonk/vite-plugin-svgr';

const CWD = process.cwd();

const vendorChunkMap = [
  ['react-vendor', ['react', 'react-dom', 'react-router-dom', 'react-redux', '@reduxjs/toolkit']],
  ['monaco-vendor', ['@monaco-editor/react', 'monaco-editor']],
  ['chart-vendor', ['echarts', 'echarts-for-react']],
  ['i18n-vendor', ['i18next', 'react-i18next']],
  ['lodash-vendor', ['lodash']],
];

const tdesignChunkMap = [
  ['tdesign-data', ['table', 'pagination', 'tree', 'list', 'descriptions', 'collapse']],
  ['tdesign-form', ['form', 'input', 'input-number', 'input-adornment', 'select', 'select-input', 'textarea', 'transfer', 'radio', 'switch', 'checkbox', 'tag-input', 'range-input', 'date-picker', 'time-picker', 'color-picker']],
  ['tdesign-overlay', ['drawer', 'dialog', 'popup', 'tooltip', 'popconfirm', 'loading', 'sticky-tool', 'dropdown']],
  ['tdesign-navigation', ['tabs', 'breadcrumb', 'menu', 'steps']],
  ['tdesign-base', ['button', 'tag', 'space', 'layout', 'row', 'col', 'card', 'avatar', 'badge', 'empty', 'message', 'notification']],
];

const getTDesignChunk = (normalized) => {
  if (normalized.includes('/node_modules/tdesign-icons-react/')) return 'tdesign-icons';
  if (normalized.includes('/node_modules/tvision-color/')) return 'tdesign-shared';
  if (!normalized.includes('/node_modules/tdesign-react/')) return undefined;

  const matched = tdesignChunkMap.find(([, components]) => (
    components.some((component) => normalized.includes(`/tdesign-react/es/${component}/`))
  ));
  return matched?.[0] || 'tdesign-shared';
};

const manualChunks = (id) => {
  if (!id.includes('node_modules')) return undefined;
  const normalized = id.split(path.sep).join('/');
  const tdesignChunk = getTDesignChunk(normalized);
  if (tdesignChunk) return tdesignChunk;
  const matched = vendorChunkMap.find(([, packages]) => packages.some((pkg) => normalized.includes(`/node_modules/${pkg}/`)));
  return matched?.[0] || 'vendor';
};

export default (params) => {
  const { mode } = params;
  const { VITE_BASE_URL } = loadEnv(mode, CWD);

  return {
    base: VITE_BASE_URL || '/',
    resolve: {
      alias: {
        assets: path.resolve(__dirname, './src/assets'),
        components: path.resolve(__dirname, './src/components'),
        configs: path.resolve(__dirname, './src/configs'),
        layouts: path.resolve(__dirname, './src/layouts'),
        modules: path.resolve(__dirname, './src/modules'),
        pages: path.resolve(__dirname, './src/pages'),
        styles: path.resolve(__dirname, './src/styles'),
        utils: path.resolve(__dirname, './src/utils'),
        services: path.resolve(__dirname, './src/services'),
        router: path.resolve(__dirname, './src/router'),
        hooks: path.resolve(__dirname, './src/hooks'),
        types: path.resolve(__dirname, './src/types'),
      },
    },

    css: {
      preprocessorOptions: {
        less: {
          modifyVars: {
            // 如需自定义组件其他 token, 在此处配置
          },
        },
      },
    },

    plugins: [
      svgr(),
      react(),
      mode === 'mock' &&
      viteMockServe({
        mockPath: './mock',
        localEnabled: true,
      }),
    ],

    build: {
      cssCodeSplit: true,
      rollupOptions: {
        output: {
          manualChunks,
        },
      },
    },

    server: {
      host: '0.0.0.0',
      port: 3003,
      proxy: {
        '/auth/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/core/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/config/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/naming/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/admin/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/metrics/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
        '/ai/mcp/v1': {
          target: 'http://127.0.0.1:8090/',
          source: false,
          changeOrigin: true,
        },
      },
    },
  };
};

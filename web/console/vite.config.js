import path from 'path';
import { loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';

const CWD = process.cwd();

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
    ],

    build: {
      cssCodeSplit: true,
      chunkSizeWarningLimit: 1800,
      rollupOptions: {
        onwarn(warning, warn) {
          if (warning.code === 'MODULE_LEVEL_DIRECTIVE' && warning.message.includes('use client')) return;
          warn(warning);
        },
      },
    },

    server: {
      host: '0.0.0.0',
      port: 3003,
      proxy: Object.fromEntries(
        [
          '/auth/v1',
          '/core/v1',
          '/config/v1',
          '/naming/v1',
          '/admin/v1',
          '/metrics/v1',
          '/api/v1',
          '/ai/mcp/v1',
          '/ai/a2a/v1',
          '/ai/agent/v1',
          '/observability/v1',
          '/system-config/v1',
        ].map((prefix) => [
          prefix,
          {
            target: 'http://127.0.0.1:8080/',
            changeOrigin: true,
          },
        ]),
      ),
    },
  };
};

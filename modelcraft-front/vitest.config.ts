import { defineConfig } from 'vitest/config'
import path from 'path'
import { type Plugin, type TransformResult } from 'vite'

// Pre-transform .tsx/.jsx files via esbuild so vite:import-analysis sees plain JS.
// Required because tsconfig.json has `jsx: preserve` (Next.js default), which
// vitest's bundler cannot handle without an explicit JSX plugin.
function tsxTransformPlugin(): Plugin {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const esbuild = require(path.resolve(__dirname, 'node_modules/esbuild')) as typeof import('esbuild')
  return {
    name: 'vitest:tsx-transform',
    enforce: 'pre',
    async transform(code, id): Promise<TransformResult | null> {
      if (!id.endsWith('.tsx') && !id.endsWith('.jsx')) return null
      const result = await esbuild.transform(code, {
        loader: id.endsWith('.tsx') ? 'tsx' : 'jsx',
        jsx: 'automatic',
        sourcemap: true,
        sourcefile: id,
      })
      const map = result.map ? (JSON.parse(result.map) as NonNullable<TransformResult['map']>) : null
      return { code: result.code, map }
    },
  }
}

export default defineConfig({
  plugins: [tsxTransformPlugin()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@web': path.resolve(__dirname, './src/web'),
      '@shared': path.resolve(__dirname, './src/shared'),
      '@api-client': path.resolve(__dirname, './src/api-client'),
    },
  },
  test: {
    environment: 'node',
    setupFiles: ['./src/test-setup.ts'],
  },
})

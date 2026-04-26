/** @type {import('next').NextConfig} */
const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
const buildId =
  process.env.BUILD_ID ||
  process.env.GIT_SHA ||
  process.env.VERCEL_GIT_COMMIT_SHA ||
  'local-dev'

const nextConfig = {
  // Use a stable build id per release to avoid client/runtime chunk mismatch.
  generateBuildId: async () => buildId,
  reactStrictMode: true,

  // 生产环境优化
  productionBrowserSourceMaps: false, // 禁用 sourcemap 减少体积
  swcMinify: true, // 使用 SWC 压缩

  // 图片优化
  images: {
    formats: ['image/avif', 'image/webp'],
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.tencentcs.com', // 云存储域名
      },
    ],
  },

  // 安全头部
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Frame-Options', value: 'SAMEORIGIN' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        ],
      },
    ]
  },
  
  // 转译 ESM 包以解决兼容性问题
  transpilePackages: [
    '@copilotkit/react-core',
    '@copilotkit/react-ui',
    '@copilotkit/runtime',
    'streamdown',
    'shiki',
    'mermaid',
  ],
  
  // 实验性功能：优化 ESM 外部包处理
  experimental: {
    instrumentationHook: true,
    esmExternals: 'loose',
    // 优化打包，减少开发模式编译时间
    optimizePackageImports: [
      'lucide-react',
      '@radix-ui/react-dialog',
      '@radix-ui/react-dropdown-menu',
      '@radix-ui/react-select',
      '@radix-ui/react-tabs',
      '@radix-ui/react-toast',
      '@radix-ui/react-tooltip',
      'antd',
      '@formily/antd-v5',
    ],
  },

  // Webpack 配置优化
  webpack: (config, { dev, isServer }) => {
    // 开发模式下跳过不需要的包的编译
    if (dev && !isServer) {
      config.resolve.alias = {
        ...config.resolve.alias,
        // 在开发模式下，将 mermaid 替换为空模块（如果不需要图表功能）
        // 'mermaid': false,
      }
    }
    return config
  },
  
  // API 代理配置
  async rewrites() {
    return [
      // 认证 API 代理（让 Set-Cookie 从 localhost 下发，浏览器才能保存）
      {
        source: '/auth/:path*',
        destination: `${backendUrl}/auth/:path*`,
      },
      // 认证 API 代理到 Go 后端 (端口 8080) - login-url, logout, check-org
      {
        source: '/api/auth/login-url',
        destination: `${backendUrl}/api/auth/login-url`,
      },
      {
        source: '/api/auth/logout',
        destination: `${backendUrl}/api/auth/logout`,
      },
      {
        source: '/api/auth/check-org',
        destination: `${backendUrl}/api/auth/check-org`,
      },
      // 组织初始化 API 代理到 Go 后端 (端口 8080)
      {
        source: '/api/orgs/initialize',
        destination: `${backendUrl}/api/orgs/initialize`,
      },
      // 组织相关 API (包括 design GraphQL) 代理到 Go 后端
      // 新格式: /org/:orgName/design/:path*
      {
        source: '/org/:orgName/design/:path*',
        destination: `${backendUrl}/org/:orgName/design/:path*`,
      },
      // 向后兼容：保留旧的 /api/design 端点（如果后端还支持）
      {
        source: '/api/design/:path*',
        destination: `${backendUrl}/api/design/:path*`,
      },
      // 运行态 GraphQL API 代理到 Go 后端 (端口 8080)
      // 路径格式: /org/:orgName/project/:projectSlug/db/:database/model/:modelName
      {
        source: '/org/:orgName/project/:projectSlug/db/:database/model/:modelName',
        destination: `${backendUrl}/org/:orgName/project/:projectSlug/db/:database/model/:modelName`,
      },
      // GraphQL API 代理到 Go 后端 (端口 8080)
      // 覆盖组织级、项目级和运行态所有 GraphQL 请求
      {
        source: '/graphql/org/:orgName/:path*',
        destination: `${backendUrl}/graphql/org/:orgName/:path*`,
      },
    ]
  },
}

export default nextConfig

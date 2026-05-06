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
      // BFF GraphQL 前缀代理（前端 Apollo 默认使用 /api/bff/graphql/*）
      // 无尾斜杠组织级端点: /api/bff/graphql/org/{orgName}
      {
        source: '/api/bff/graphql/org/:orgName',
        destination: `${backendUrl}/graphql/org/:orgName`,
      },
      // 含尾斜杠及更深层级（项目级/运行态）端点
      {
        source: '/api/bff/graphql/org/:orgName/:path*',
        destination: `${backendUrl}/graphql/org/:orgName/:path*`,
      },
      // End-User GraphQL（终端用户运行态 GraphQL）
      // Endpoint: /api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}
      {
        source: '/api/bff/graphql/end-user/org/:orgName/project/:projectSlug',
        destination: `${backendUrl}/graphql/end-user/org/:orgName/project/:projectSlug`,
      },
      // End-User Runtime GraphQL（终端用户模型数据 CRUD）
      // 经 Gateway 鉴权，注入 X-User-ID + X-User-Type: end_user
      {
        source: '/api/bff/graphql/end-user/org/:orgName/project/:projectSlug/db/:db/model/:model',
        destination: `${backendUrl}/graphql/end-user/org/:orgName/project/:projectSlug/db/:db/model/:model`,
      },
      // 终端用户公开认证接口（JWT）
      {
        source: '/api/end-user/:path*',
        destination: `${backendUrl}/api/end-user/:path*`,
      },
      // 内部 API（终端用户数据）
      {
        source: '/internal/:path*',
        destination: `${backendUrl}/internal/:path*`,
      },
      // 用户信息 API
      {
        source: '/api/user/:path*',
        destination: `${backendUrl}/api/user/:path*`,
      },
    ]
  },
}

export default nextConfig

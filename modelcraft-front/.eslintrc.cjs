module.exports = {
  root: true,
  extends: ['next/core-web-vitals', 'plugin:tailwindcss/recommended'],
  plugins: ['tailwindcss', 'depend', '@typescript-eslint'],
  rules: {
    'react/no-unescaped-entities': 'off',
    'tailwindcss/classnames-order': 'warn',
    'tailwindcss/enforces-negative-arbitrary-values': 'warn',
    'tailwindcss/enforces-shorthand': 'warn',
    'tailwindcss/migration-from-tailwind-2': 'off',
    'tailwindcss/no-arbitrary-value': 'off',
    'tailwindcss/no-contradicting-classname': 'error',
    'tailwindcss/no-custom-classname': 'off',
    'tailwindcss/no-unnecessary-arbitrary-value': 'warn',

    // --- 字体规范强制 ---
    // 禁止使用超出设计规范的字体权重和非语义化颜色
    // 设计规范参见: src/lib/typography.ts, ai-metadata/style/STYLE.md
    'no-restricted-syntax': [
      'error',
      // 字体权重：禁止 font-bold / font-extrabold / font-black
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\bfont-(bold|extrabold|black)\\b/]',
        message: '禁止使用 font-bold/font-extrabold/font-black。请使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\bfont-(bold|extrabold|black)\\b/]',
        message: '禁止使用 font-bold/font-extrabold/font-black。请使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
      },
      // 文字颜色：禁止 text-gray-{400-900}，应改用语义化变量
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-gray-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-gray-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      // 架构守则：禁止 web 层和 app 页面直连后端 GraphQL 端点
      // 所有 GraphQL 请求必须经过 BFF 代理（/api/bff/graphql/org/...）
      // 参见 ai-metadata/front/development/bff-design.md
      {
        selector: 'Literal[value=/(?<!\\/api\\/bff)\\/graphql\\/org\\//]',
        message: '禁止直连后端 GraphQL 端点。请使用 BFF 代理路径：/api/bff/graphql/org/... 参见 bff-design.md',
      },
      {
        selector: 'TemplateLiteral > TemplateElement[value.raw=/(?<!\\/api\\/bff)\\/graphql\\/org\\//]',
        message: '禁止直连后端 GraphQL 端点。请使用 BFF 代理路径：/api/bff/graphql/org/... 参见 bff-design.md',
      },
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-slate-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-slate-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
    ],
  },
  overrides: [
    // --- TypeScript any 安全规则 ---
    {
      files: ['**/*.{ts,tsx}'],
      excludedFiles: ['src/generated/**', 'src/mocks/**', 'e2e/**', 'playwright.config.ts'],
      parser: '@typescript-eslint/parser',
      parserOptions: {
        project: ['./tsconfig.json'],
        tsconfigRootDir: __dirname,
      },
      rules: {
        // 明确禁止 any 的核心规则：
        '@typescript-eslint/no-explicit-any': 'error',
        '@typescript-eslint/no-unsafe-argument': 'error',
        '@typescript-eslint/no-unsafe-assignment': 'error',
        '@typescript-eslint/no-unsafe-call': 'error',
        '@typescript-eslint/no-unsafe-member-access': 'error',
        '@typescript-eslint/no-unsafe-return': 'error',
      },
    },
    // --- 层级边界强制 ---
    // 注意: eslint-plugin-depend v1.x 不提供 depguard 规则（仅有 ban-dependencies）
    // 使用内置 no-restricted-imports + overrides 实现等效的有向层级隔离
    // Web 层不能直接 import BFF 层内部实现（只允许通过 public facade）
    {
      files: ['src/web/**/*.{ts,tsx,js,jsx}'],
      rules: {
        'no-restricted-imports': ['error', {
          patterns: [
            {
              group: ['@bff/*/!(public)', '../bff/**/!(public)', '../../bff/**/!(public)', '**/bff/**/!(public)'],
              message: 'Web 层不能直接 import BFF 层内部实现，请通过 public facade 访问（如 @bff/auth/public）',
            },
          ],
        }],
      },
    },
    // BFF 层不能依赖 Web 层
    {
      files: ['src/bff/**/*.{ts,tsx,js,jsx}'],
      rules: {
        'no-restricted-imports': ['error', {
          patterns: [
            {
              group: ['@web/*', '../web/**', '../../web/**', '**/web/**'],
              message: 'BFF 层不能依赖 Web 层',
            },
          ],
        }],
      },
    },
    // BFF 代理路由和 go-client 本身需要引用后端路径，豁免直连检测
    {
      files: [
        'src/app/api/bff/**/*.{ts,tsx}',
        'src/bff/**/*.{ts,tsx}',
      ],
      rules: {
        'no-restricted-syntax': 'off',
      },
    },
  ],
  settings: {
    tailwindcss: {
      config: 'tailwind.config.ts',
      cssFiles: ['**/*.css', '!**/node_modules', '!**/.*', '!**/dist', '!**/build'],
    },
  },
}

import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: [
    'contract/graph/org/schema/*.graphql',
    'contract/graph/project/schema/*.graphql',
  ],
  documents: ['src/api-client/**/*.ts', '!src/api-client/rbac/**'],
  generates: {
    // 1. TypeScript 类型生成
    'src/generated/graphql.ts': {
      plugins: [
        'typescript',
        'typescript-operations',
      ],
      config: {
        avoidOptionals: false,
        skipTypename: false,
        enumsAsTypes: true,
        skipDocumentsValidation: {
          skipValidationAgainstSchema: true,
        },
        scalars: {
          DateTime: 'string',
          ID: 'string',
        },
      },
    },
    // 2. MSW mock handlers — org 域（禁止手动编辑，由 codegen 生成）
    'src/mocks/handlers/org/generated.ts': {
      schema: 'contract/graph/org/schema/*.graphql',
      plugins: ['typescript', 'typescript-operations', 'typescript-msw'],
      config: {
        skipDocumentsValidation: {
          skipValidationAgainstSchema: true,
        },
        scalars: {
          DateTime: 'string',
          ID: 'string',
        },
      },
    },
    // 3. MSW mock handlers — project 域（禁止手动编辑，由 codegen 生成）
    'src/mocks/handlers/project/generated.ts': {
      schema: 'contract/graph/project/schema/*.graphql',
      plugins: ['typescript', 'typescript-operations', 'typescript-msw'],
      config: {
        skipDocumentsValidation: {
          skipValidationAgainstSchema: true,
        },
        scalars: {
          DateTime: 'string',
          ID: 'string',
        },
      },
    },
  },
}

export default config

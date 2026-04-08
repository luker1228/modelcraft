import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: [
    'contract/graph/org/schema/*.graphql',
    'contract/graph/project/schema/*.graphql',
  ],
  documents: 'src/web/graphql/**/*.ts',
  generates: {
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
  },
}

export default config

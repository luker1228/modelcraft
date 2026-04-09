import 'dotenv/config'
import { World, setWorldConstructor, IWorldOptions } from '@cucumber/cucumber'
import { GraphQLClient } from './graphql-client'
import { RestClient, RestResult } from './rest-client'

export class ModelCraftWorld extends World {
  // 客户端
  readonly restClient: RestClient
  readonly projectClient: GraphQLClient

  // 认证
  token: string | null = null

  // 固定 fixture（从环境变量读取）
  readonly orgName: string
  readonly projectSlug: string

  // 当前 Scenario 创建的资源（After 钩子清理用）
  createdModelIds: string[] = []
  createdEnumNames: string[] = []

  // 当前操作的 model ID（跨 step 传递，替代模块级变量）
  currentModelId: string | null = null

  // 模型 baseName → 实际 ID 映射（供 lfk.steps.ts 使用）
  modelMap: Record<string, string> = {}

  // 最近一次 Given 创建的实际名称（供 "再次创建" step 使用）
  lastModelName: string | null = null
  lastEnumName: string | null = null

  // 最近操作结果（When → Then 传递）— GraphQL 场景
  lastResponse: Record<string, unknown> | null = null
  lastError: Error | null = null

  // Auth 相关状态（REST 场景）
  lastRestResult: RestResult<unknown> | null = null
  registeredPhone: string | null = null
  registeredUserName: string | null = null
  registeredPassword: string | null = null
  currentRefreshToken: string | null = null
  currentUserId: string | null = null

  // Org 初始化相关状态（REST 场景）
  lastMembershipsCount: number | null = null
  initOrgName: string | null = null
  initDisplayName: string | null = null
  initAlreadyExists: boolean | null = null
  firstInitOrgName: string | null = null
  secondInitOrgName: string | null = null

  constructor(options: IWorldOptions) {
    super(options)

    this.orgName = process.env.TEST_ORG_NAME ?? 'test-org'
    this.projectSlug = process.env.TEST_PROJECT_SLUG ?? 'test-project'

    this.restClient = new RestClient()
    this.projectClient = new GraphQLClient(this.orgName, this.projectSlug)

    // 若提供了预设 token，直接使用（跳过 OAuth 流程）
    const token = process.env.TEST_ACCESS_TOKEN
    if (token) {
      this.token = token
      this.projectClient.setAuth(token)
    }
  }
}

setWorldConstructor(ModelCraftWorld)

import { After, Before, BeforeAll } from '@cucumber/cucumber'
import { ModelCraftWorld } from './world'
import { signJWT } from './jwt'
import { RestClient } from './rest-client'

// ──────── BDD 测试登录用户 ────────
// 优先使用 TEST_ACCESS_TOKEN；若未提供，则通过测试账号登录后签发 JWT
const BDD_LOGIN_PHONE = process.env.TEST_LOGIN_PHONE ?? '19900000001'
const BDD_LOGIN_PASSWORD = process.env.TEST_LOGIN_PASSWORD ?? 'bddtest12345'

BeforeAll(async function () {
  // 如果已提供 TEST_ACCESS_TOKEN，跳过自动配置
  if (process.env.TEST_ACCESS_TOKEN) return

  const client = new RestClient()
  const loginResult = await client.login(BDD_LOGIN_PHONE, BDD_LOGIN_PASSWORD, 'PHONE')
  if (!loginResult.data) {
    throw new Error(`BDD auto-setup: login failed — ${JSON.stringify(loginResult.error)}`)
  }

  // 签发 JWT（有效期 1 小时，含 iss: "modelcraft"）
  process.env.TEST_ACCESS_TOKEN = signJWT(loginResult.data.userId, 3600)
})

// ──────── GraphQL 操作 ────────

const DELETE_MODEL = `
  mutation DeleteModel($id: ID!) {
    deleteModel(id: $id) {
      error { __typename }
    }
  }
`

const DELETE_ENUM = `
  mutation DeleteEnum($name: String!) {
    deleteEnum(name: $name) {
      error { __typename }
    }
  }
`

// 每个 Scenario 前重置追踪列表
Before(function (this: ModelCraftWorld) {
  this.createdModelIds = []
  this.createdEnumNames = []
  this.currentModelId = null
  this.modelMap = {}
  this.lastModelName = null
  this.lastEnumName = null
  this.lastResponse = null
  this.lastError = null
  // Auth 相关状态
  this.lastRestResult = null
  this.registeredPhone = null
  this.registeredUserName = null
  this.registeredPassword = null
  this.currentRefreshToken = null
  this.currentUserId = null
  this.currentOrgName = null
  this.lastMembershipsCount = null
  this.initOrgName = null
  this.initDisplayName = null
  this.initAlreadyExists = null
  this.firstInitOrgName = null
  this.secondInitOrgName = null

  // 如果 BeforeAll 生成了 token，确保 World 也使用它
  if (!this.token && process.env.TEST_ACCESS_TOKEN) {
    this.token = process.env.TEST_ACCESS_TOKEN
    this.projectClient.setAuth(process.env.TEST_ACCESS_TOKEN)
    this.orgClient.setAuth(process.env.TEST_ACCESS_TOKEN)
  }
})

// 每个 Scenario 后通过 API 清理创建的数据（@smoke 除外保留数据方便调试）
After({ tags: 'not @smoke' }, async function (this: ModelCraftWorld) {
  for (const id of [...this.createdModelIds].reverse()) {
    try {
      await this.projectClient.mutate(DELETE_MODEL, { id })
    } catch {
      // 静默处理
    }
  }

  for (const name of this.createdEnumNames) {
    try {
      await this.projectClient.mutate(DELETE_ENUM, { name })
    } catch {
      // 静默处理
    }
  }
})

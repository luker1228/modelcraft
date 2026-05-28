import { Given, When, Then } from '@cucumber/cucumber'
import { expect } from 'expect'
import { ModelCraftWorld } from '../support/world'

// ── GraphQL 文档 ─────────────────────────────────────────────────────────────

const REGISTER_MODEL_DATABASE = `
  mutation RegisterModelDatabase($input: RegisterModelDatabaseInput!) {
    registerModelDatabase(input: $input) {
      ... on ModelDatabase {
        id
        name
        title
        description
        mode
      }
      ... on InvalidInput {
        __typename
        message
      }
      ... on ResourceNotFound {
        __typename
        message
      }
    }
  }
`

const UPDATE_MODEL_DATABASE = `
  mutation UpdateModelDatabase($id: ID!, $input: UpdateModelDatabaseInput!) {
    updateModelDatabase(id: $id, input: $input) {
      id
      name
      title
      description
      mode
    }
  }
`

const UNREGISTER_MODEL_DATABASE = `
  mutation UnregisterModelDatabase($id: ID!) {
    unregisterModelDatabase(id: $id)
  }
`

const LIST_MODEL_DATABASES = `
  query ListModelDatabases {
    modelDatabases {
      id
      name
      title
      description
      mode
    }
  }
`

const LIST_CLUSTER_RAW_DATABASES = `
  query ListClusterRawDatabases {
    clusterRawDatabases {
      name
      isRegistered
    }
  }
`

// ── 类型定义 ──────────────────────────────────────────────────────────────────

interface ModelDatabase {
  id: string
  name: string
  title: string
  description: string
  mode: string
}

interface RegisterResult {
  registerModelDatabase: ModelDatabase | { __typename: string; message: string }
}

interface ListResult {
  modelDatabases: ModelDatabase[]
}

interface RawDatabase {
  name: string
  isRegistered: boolean
}

interface RawListResult {
  clusterRawDatabases: RawDatabase[]
}

// ── World 扩展（database 场景专用属性） ───────────────────────────────────────
// 使用 this 直接挂载，避免修改 world.ts

// ── Given Steps ──────────────────────────────────────────────────────────────

Given('已接管数据库 {string}，模式为 {string}', async function (
  this: ModelCraftWorld,
  dbName: string,
  mode: string,
) {
  // 先取消接管（确保幂等），忽略错误
  const existing = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  const found = existing.modelDatabases.find((db) => db.name === dbName)
  if (found) {
    await this.projectClient.mutate<{ unregisterModelDatabase: boolean }>(
      UNREGISTER_MODEL_DATABASE,
      { id: found.id },
    )
  }

  const res = await this.projectClient.mutate<RegisterResult>(REGISTER_MODEL_DATABASE, {
    input: { name: dbName, title: dbName, mode },
  })

  const result = res.registerModelDatabase
  if ('__typename' in result) {
    throw new Error(`前置条件：接管数据库 ${dbName} 失败：${result.message}`)
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(this as any).currentDatabaseId = result.id
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(this as any).currentDatabaseName = result.name
})

// ── When Steps ───────────────────────────────────────────────────────────────

When(
  '我接管数据库 {string}，模式为 {string}，友好名称为 {string}',
  async function (this: ModelCraftWorld, dbName: string, mode: string, title: string) {
    // 先确保未接管（清理可能残留的记录）
    const existing = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
    const found = existing.modelDatabases.find((db) => db.name === dbName)
    if (found) {
      await this.projectClient.mutate<{ unregisterModelDatabase: boolean }>(
        UNREGISTER_MODEL_DATABASE,
        { id: found.id },
      )
    }

    const res = await this.projectClient.mutate<RegisterResult>(REGISTER_MODEL_DATABASE, {
      input: { name: dbName, title, mode },
    })
    this.lastResponse = { registerModelDatabase: res.registerModelDatabase }

    const result = res.registerModelDatabase
    if (!('__typename' in result)) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(this as any).currentDatabaseId = result.id
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(this as any).currentDatabaseName = result.name
    }
  },
)

When(
  '我再次接管数据库 {string}，模式为 {string}',
  async function (this: ModelCraftWorld, dbName: string, mode: string) {
    const res = await this.projectClient.mutate<RegisterResult>(REGISTER_MODEL_DATABASE, {
      input: { name: dbName, title: dbName, mode },
    })
    this.lastResponse = { registerModelDatabase: res.registerModelDatabase }
  },
)

When('我查询已接管数据库列表', async function (this: ModelCraftWorld) {
  const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  this.lastResponse = { modelDatabases: res.modelDatabases }
})

When(
  '我将该数据库的友好名称更新为 {string}，描述更新为 {string}',
  async function (this: ModelCraftWorld, title: string, description: string) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const id = (this as any).currentDatabaseId as string
    expect(id).toBeTruthy()

    const res = await this.projectClient.mutate<{ updateModelDatabase: ModelDatabase }>(
      UPDATE_MODEL_DATABASE,
      { id, input: { title, description } },
    )
    this.lastResponse = { updateModelDatabase: res.updateModelDatabase }
  },
)

When(
  '我将该数据库的模式更新为 {string}',
  async function (this: ModelCraftWorld, mode: string) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const id = (this as any).currentDatabaseId as string
    expect(id).toBeTruthy()

    const res = await this.projectClient.mutate<{ updateModelDatabase: ModelDatabase }>(
      UPDATE_MODEL_DATABASE,
      { id, input: { mode } },
    )
    this.lastResponse = { updateModelDatabase: res.updateModelDatabase }
  },
)

When('我取消接管该数据库', async function (this: ModelCraftWorld) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const id = (this as any).currentDatabaseId as string
  expect(id).toBeTruthy()

  const res = await this.projectClient.mutate<{ unregisterModelDatabase: boolean }>(
    UNREGISTER_MODEL_DATABASE,
    { id },
  )
  this.lastResponse = { unregisterModelDatabase: res.unregisterModelDatabase }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(this as any).currentDatabaseId = null
})

When('我查询集群原始数据库列表', async function (this: ModelCraftWorld) {
  const res = await this.projectClient.query<RawListResult>(LIST_CLUSTER_RAW_DATABASES, {})
  this.lastResponse = { clusterRawDatabases: res.clusterRawDatabases }
})

// ── Then Steps ───────────────────────────────────────────────────────────────

Then('接管应该成功', function (this: ModelCraftWorld) {
  const result = (this.lastResponse?.registerModelDatabase ?? null) as
    | ModelDatabase
    | { __typename: string; message: string }
    | null
  expect(result).not.toBeNull()
  expect('__typename' in (result ?? {})).toBe(false)
})

Then('应该返回接管错误 {string}', function (this: ModelCraftWorld, expectedType: string) {
  const result = this.lastResponse?.registerModelDatabase as
    | { __typename: string; message: string }
    | undefined
  expect(result?.__typename).toBe(expectedType)
})

Then(
  '已接管数据库列表中应包含名为 {string} 的数据库',
  async function (this: ModelCraftWorld, dbName: string) {
    const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
    const names = res.modelDatabases.map((db) => db.name)
    expect(names).toContain(dbName)
  },
)

Then(
  '已接管数据库列表中不应包含名为 {string} 的数据库',
  async function (this: ModelCraftWorld, dbName: string) {
    const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
    const names = res.modelDatabases.map((db) => db.name)
    expect(names).not.toContain(dbName)
  },
)

Then('该数据库的模式应该是 {string}', async function (this: ModelCraftWorld, expectedMode: string) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dbName = (this as any).currentDatabaseName as string
  expect(dbName).toBeTruthy()
  const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  const db = res.modelDatabases.find((d) => d.name === dbName)
  expect(db).toBeDefined()
  expect(db?.mode).toBe(expectedMode)
})

Then('列表中应包含名为 {string} 的数据库', function (this: ModelCraftWorld, dbName: string) {
  const list = (this.lastResponse?.modelDatabases ?? []) as ModelDatabase[]
  const names = list.map((db) => db.name)
  expect(names).toContain(dbName)
})

Then('更新应该成功', function (this: ModelCraftWorld) {
  const result = this.lastResponse?.updateModelDatabase as ModelDatabase | undefined
  expect(result).toBeDefined()
  expect(result?.id).toBeTruthy()
})

Then('查询该数据库，友好名称应该是 {string}', async function (
  this: ModelCraftWorld,
  expectedTitle: string,
) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dbName = (this as any).currentDatabaseName as string
  const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  const db = res.modelDatabases.find((d) => d.name === dbName)
  expect(db?.title).toBe(expectedTitle)
})

Then('查询该数据库，描述应该是 {string}', async function (
  this: ModelCraftWorld,
  expectedDescription: string,
) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dbName = (this as any).currentDatabaseName as string
  const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  const db = res.modelDatabases.find((d) => d.name === dbName)
  expect(db?.description).toBe(expectedDescription)
})

Then('查询该数据库，模式应该是 {string}', async function (
  this: ModelCraftWorld,
  expectedMode: string,
) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dbName = (this as any).currentDatabaseName as string
  const res = await this.projectClient.query<ListResult>(LIST_MODEL_DATABASES, {})
  const db = res.modelDatabases.find((d) => d.name === dbName)
  expect(db?.mode).toBe(expectedMode)
})

Then('取消接管应该成功', function (this: ModelCraftWorld) {
  const result = this.lastResponse?.unregisterModelDatabase
  expect(result).toBe(true)
})

Then('原始列表中应包含 {string}', function (this: ModelCraftWorld, dbName: string) {
  const list = (this.lastResponse?.clusterRawDatabases ?? []) as RawDatabase[]
  const names = list.map((db) => db.name)
  expect(names).toContain(dbName)
})

Then('原始列表中不应包含系统库 {string}', function (this: ModelCraftWorld, sysDb: string) {
  const list = (this.lastResponse?.clusterRawDatabases ?? []) as RawDatabase[]
  const names = list.map((db) => db.name)
  expect(names).not.toContain(sysDb)
})

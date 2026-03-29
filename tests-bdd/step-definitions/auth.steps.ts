// tests-bdd/step-definitions/auth.steps.ts
import { Given } from '@cucumber/cucumber'
import { ModelCraftWorld } from '../support/world'

/**
 * 以管理员身份登录。
 * 优先使用 .env.test 中的 TEST_ACCESS_TOKEN（已在 World 构造函数中设置）。
 * 若未设置则抛出明确错误，提示配置方法。
 */
Given('我以管理员身份登录', function (this: ModelCraftWorld) {
  if (!this.token) {
    throw new Error(
      '未找到 TEST_ACCESS_TOKEN。请在 tests-bdd/.env.test 中设置:\n' +
      'TEST_ACCESS_TOKEN=<your-token>\n' +
      '获取方式：在 modelcraft-backend/ 目录运行 just test-user-setup'
    )
  }
  // token 已在 World 构造函数中设置到 projectClient，此处只做验证
})

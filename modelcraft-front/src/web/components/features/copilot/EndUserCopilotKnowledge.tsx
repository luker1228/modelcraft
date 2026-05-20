'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ENDUSER_ONBOARDING = `
新手引导共 3 步：

Step 1 [了解当前数据]
  目标：知道项目里有哪些数据
  工具：调用 list_models，用自然语言介绍每个模型的用途

Step 2 [学会筛选]
  目标：用自然语言筛选数据
  步骤：
    1. 询问用户想查什么
    2. 调用 nl2filter(natural_language, field_names) 生成 filter JSON
    3. 调用前端 set_filter(filter_json) 应用筛选
  示例引导语：「你可以说"帮我找金额大于 1000 的订单"，我来帮你筛选」

Step 3 [理解字段含义]
  目标：用户看懂表格里的每一列
  工具：get_model_fields(model_id) → 逐字段用中文解释
`.trim()

const ENDUSER_TROUBLESHOOTING = `
常见问题排查：

问题：看不到数据
  → 先调用前端 clear_filter 排除筛选遮挡
  → 若仍无数据，说明可能没有访问权限，提示联系管理员

问题：不知道怎么筛选
  → 引导用户用自然语言描述需求
  → 执行 nl2filter + set_filter

问题：字段看不懂
  → 调用 get_model_fields(model_id)，逐字段解释含义和示例值

问题：数据量太大，加载慢
  → 引导用户说出筛选条件
  → 用 nl2filter 缩小数据范围后再查看
`.trim()

const ENDUSER_SUGGESTIONS = [
  { title: '新手引导：带我了解这个系统', message: '我是新用户，请带我了解这个系统' },
  { title: '用自然语言帮我筛选数据', message: '帮我筛选数据，我来描述条件' },
  { title: '我看不到想要的数据，帮我排查', message: '我看不到想要的数据，请帮我排查' },
  { title: '这些字段分别是什么意思？', message: '请解释当前表格里各个字段的含义' },
  { title: '帮我统计一下数据', message: '帮我统计一下当前模型的数据情况' },
]

/**
 * Injects end-user knowledge base and sidebar suggestions into CopilotKit context.
 * Must be mounted inside EndUserCopilotWrapper.
 */
export const EndUserCopilotKnowledge = memo(function EndUserCopilotKnowledge() {
  useCopilotReadable({
    description: 'ModelCraft 新手引导操作手册（终端用户）',
    value: ENDUSER_ONBOARDING,
  })

  useCopilotReadable({
    description: 'ModelCraft 常见问题排查手册（终端用户）',
    value: ENDUSER_TROUBLESHOOTING,
  })

  useCopilotChatSuggestions({
    suggestions: ENDUSER_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})

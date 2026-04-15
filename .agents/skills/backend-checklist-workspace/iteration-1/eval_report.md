# backend-checklist Trigger Eval — Iteration 1 结果报告

生成时间：2026-04-15

---

## 汇总

| Eval | 预期 | with_skill 结果 | without_skill 结果 | 结论 |
|------|------|-----------------|-------------------|------|
| 1. trigger-add-mistake | 触发 add | ✅ 触发，识别已有编号 | ✅ 找到文件，但无 add 流程 | **with_skill 更规范** |
| 2. trigger-checklist-review | 触发 review | ✅ 触发，命中 BM-20260415-0001 | ⚠️ 靠经验发现问题，无编号关联 | **with_skill 明显更好** |
| 3. trigger-sql-repo-check | 触发 review | ✅ 触发，命中 2 处，含新问题 | ✅ 靠经验发现 3 个 bug，无规则关联 | **with_skill 有规则溯源** |
| 4. trigger-history-query | 触发 review | ✅ 触发，精确返回错题本内容 | ⚠️ 找到了但内容发散，混入推断 | **with_skill 更聚焦** |
| 5. trigger-record-bug | 触发 add | ✅ 触发，识别已有，不重复写 | ❌ 写了通用说明，不知道错题本位置 | **with_skill 明显更好** |
| 6. no-trigger-code-review | 不触发 | ✅ 不触发（主动读了 md 辅助） | N/A | **正确不触发** |
| 7. no-trigger-bdd-test | 不触发 | ✅ 不触发，去跑 BDD | N/A | **正确不触发** |
| 8. no-trigger-db-migrate | 不触发 | ✅ 不触发，理由清晰 | N/A | **正确不触发** |
| 9. no-trigger-fix-bug | 不触发 | ✅ 不触发，直接修复 | N/A | **正确不触发** |
| 10. no-trigger-memory | 不触发 | ✅ 不触发，走 memory | N/A | **正确不触发** |

---

## 评分

### Trigger 准确率

| 类型 | 正确 | 总计 | 准确率 |
|------|------|------|--------|
| should-trigger（5条）| 5 | 5 | **100%** |
| should-not-trigger（5条）| 5 | 5 | **100%** |
| **总计** | **10** | **10** | **100%** |

### with_skill vs without_skill 质量对比

| Eval | with_skill | without_skill | 差距 |
|------|-----------|---------------|------|
| 1 add | 规范流程，引用编号 | 能找到文件，无标准模板 | 有差距 |
| 2 review-inline | 关联历史编号，格式化输出 | 发现问题但无溯源 | 明显差距 |
| 3 file-review | 结构化输出+规则引用 | 发现更多 bug 但无规则关联 | 中等差距 |
| 4 history | 精确聚焦错题本 | 发散混入其他文档推断 | 有差距 |
| 5 record | 识别重复，不写入 | 不知道错题本位置 | 明显差距 |

---

## 关键发现

### ✅ 表现好的地方

1. **触发词覆盖全面**：「加入错题本」「用错题本 check」「checklist review」「有没有历史 bug」「记录一下」全部正确触发
2. **不触发边界清晰**：代码 review、BDD 测试、DB 迁移、直接修 bug、memory 查询全部正确不触发
3. **add 模式防重复**：发现已有编号时不重复写入，逻辑正确
4. **review 模式关联溯源**：命中规则时能关联到 BM- 编号，有可追溯性

### ⚠️ 待改进的地方

1. **eval 6 边界行为**：no-trigger-code-review 时，agent 虽然没触发 skill 流程，但"主动读了 common-mistakes.md"——这是一个模糊的灰色地带。是否应该在普通代码 review 时也附带 checklist 结果？描述里可以更明确说明：普通 review 时不强制触发，但 backend-reviewer 工作流里会触发。

2. **without_skill 在文件存在时也能工作**：eval 1/4 中，without_skill 也找到了 common-mistakes.md，这是因为文件在项目里可以被搜到。skill 的核心价值在于：① 规范模板 ② 防重复写入 ③ 触发词明确引导。

---

## 结论

**skill 触发效果：合格，100% 准确率。** 无需调整描述。

with_skill 相比 without_skill 的核心价值体现在：
- **标准化**：add 时有规范模板，review 时有结构化输出格式
- **溯源**：命中时关联 BM- 编号，跨会话可追踪
- **边界**：不重复写入，不发散到其他文档

---

## 顺带发现的潜在 Bug（eval 3 中挖出）

`GetModelByName` SQL 缺少 `org_name` 条件，与 BM-20260415-0001 是同类问题，建议另行修复。

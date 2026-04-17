# Adapter 契约测试规范（Go）

> **优先级: 中** - 适用于 `A -> B` 结构体转换、mapper、adapter、converter 代码。

## 核心规则

**Harness 负责执行与守门，不负责定义正确性。**

Adapter 正确性必须先固化为测试资产，再接入 CI：

1. **表驱动测试（必选）**：明确字段映射契约（默认值、空值、枚举、时间、指针/切片）。
2. **Golden 测试（推荐）**：字段多、对象深时，用 JSON 快照做全量回归对比。
3. **Fuzz 测试（推荐）**：覆盖脏数据和边界组合，重点保证不 panic。
4. **不变量断言（必选）**：例如 ID 不丢失、时间格式统一、未知枚举兜底、nil 输入不 panic。

---

## 不变量示例

- ID 转换后不能为空（在合法输入前提下）
- CreatedAt/UpdatedAt 统一输出为 RFC3339（建议 UTC）
- nil 输入不可 panic
- unknown enum 必须落到 `UNKNOWN`（或约定兜底值）
- round-trip（如 DTO->Domain->DTO）关键字段保持一致

---

## Harness 分层执行建议

1. **PR 必跑**：table-driven + 关键 golden。
2. **变更相关执行**：由 Test Intelligence 按代码变更选择相关 adapter 单测。
3. **定时回归（nightly）**：fuzz 或更重的全量回归。

> 原则：先定义契约，再让 Harness 执行契约并拦截回归。

---

## 推荐目录结构

```text
internal/adapters/
  user_adapter.go
  user_adapter_test.go
  order_adapter.go
  order_adapter_test.go

testdata/
  user_adapter/
    basic.golden.json
    nulls.golden.json
    enum_fallback.golden.json
```

---

## 最小测试模板

```go
func TestUserAdapter(t *testing.T) {
    cases := []struct {
        name string
        in   UserDO
        want UserDTO
    }{
        {name: "basic", in: UserDO{ID: 1, Name: "alice"}, want: UserDTO{ID: "1", DisplayName: "alice"}},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := ToUserDTO(tc.in)
            if diff := cmp.Diff(tc.want, got); diff != "" {
                t.Fatalf("mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

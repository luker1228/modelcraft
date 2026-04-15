# model_enums 表结构核查报告

## 是否触发 backend-checklist skill

**未触发**。

**理由**：backend-checklist skill 有两个能力：
1. **add** — 将真实 Bug 案例加入错题本
2. **review** — 读取错题本，对代码逐条检查是否重蹈历史 Bug

触发词为：「加入错题本」、「记录这个错误」、「用错题本 check」、「checklist review」、「有没有历史 bug」等。

本次任务是「确认迁移后数据库表结构与 schema 定义是否一致」，属于**数据库迁移验证**任务，不涉及 Bug 录入或历史 Bug 模式审查，与 backend-checklist skill 的使用场景不匹配，因此不触发。

---

## Schema 定义（`db/schema/mysql/03_model_domain.sql`）

```sql
CREATE TABLE IF NOT EXISTS `model_enums` (
  `id`             VARCHAR(36)   NOT NULL COMMENT '枚举唯一标识符',
  `org_name`       VARCHAR(36)   NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug`   VARCHAR(64)   NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',
  `name`           VARCHAR(64)   NOT NULL COMMENT '枚举名称（唯一标识）',
  `display_name`   VARCHAR(255)  NOT NULL COMMENT '枚举显示名称',
  `description`    TEXT          NULL     COMMENT '枚举描述信息',
  `options`        JSON          NOT NULL COMMENT '枚举选项配置（JSON数组格式）',
  `is_multi_select` TINYINT(1)   NULL DEFAULT 0 COMMENT '是否支持多选',
  `created_at`     DATETIME(3)   NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at`     DATETIME(3)   NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_enums_name` (`org_name`, `project_slug`, `name`),
  KEY `idx_model_enums_project` (`org_name`, `project_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型枚举定义表';
```

---

## 实际数据库表结构（`SHOW CREATE TABLE model_enums`）

```sql
CREATE TABLE `model_enums` (
  `id`             varchar(36)   COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '枚举唯一标识符',
  `org_name`       varchar(36)   COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug`   varchar(64)   COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',
  `name`           varchar(64)   COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '枚举名称（唯一标识）',
  `display_name`   varchar(255)  COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '枚举显示名称',
  `description`    text          COLLATE utf8mb4_unicode_ci COMMENT '枚举描述信息',
  `options`        json          NOT NULL COMMENT '枚举选项配置（JSON数组格式）',
  `is_multi_select` tinyint(1)   DEFAULT '0' COMMENT '是否支持多选',
  `created_at`     datetime(3)   DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at`     datetime(3)   DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_enums_name` (`org_name`,`project_slug`,`name`) COMMENT '组织+项目内枚举名称唯一索引',
  KEY `idx_model_enums_project` (`org_name`,`project_slug`) COMMENT '项目查询索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型枚举定义表'
```

---

## 逐项对比

| 字段/属性 | Schema 定义 | 实际表结构 | 一致 |
|-----------|------------|-----------|------|
| `id` | VARCHAR(36) NOT NULL | varchar(36) NOT NULL | ✅ |
| `org_name` | VARCHAR(36) NOT NULL | varchar(36) NOT NULL | ✅ |
| `project_slug` | VARCHAR(64) NOT NULL | varchar(64) NOT NULL | ✅ |
| `name` | VARCHAR(64) NOT NULL | varchar(64) NOT NULL | ✅ |
| `display_name` | VARCHAR(255) NOT NULL | varchar(255) NOT NULL | ✅ |
| `description` | TEXT NULL | text (NULL允许) | ✅ |
| `options` | JSON NOT NULL | json NOT NULL | ✅ |
| `is_multi_select` | TINYINT(1) NULL DEFAULT 0 | tinyint(1) DEFAULT '0' | ✅ |
| `created_at` | DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) | datetime(3) DEFAULT CURRENT_TIMESTAMP(3) | ✅ |
| `updated_at` | DATETIME(3) ... ON UPDATE CURRENT_TIMESTAMP(3) | datetime(3) ... ON UPDATE CURRENT_TIMESTAMP(3) | ✅ |
| PRIMARY KEY | `id` | `id` | ✅ |
| UNIQUE KEY `idx_model_enums_name` | (org_name, project_slug, name) | (org_name, project_slug, name) | ✅ |
| KEY `idx_model_enums_project` | (org_name, project_slug) | (org_name, project_slug) | ✅ |
| ENGINE | InnoDB | InnoDB | ✅ |
| CHARSET | utf8mb4 COLLATE utf8mb4_unicode_ci | utf8mb4 COLLATE utf8mb4_unicode_ci | ✅ |
| COMMENT | '模型枚举定义表' | '模型枚举定义表' | ✅ |

---

## 结论

**✅ 完全一致。**

`model_enums` 表的实际数据库结构与 `db/schema/mysql/03_model_domain.sql` 中的 Schema 定义**完全一致**，所有字段、类型、约束、索引、引擎和字符集均匹配。数据库迁移执行正确，无偏差。

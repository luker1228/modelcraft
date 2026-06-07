# Constrained GraphQL Runtime Figures Outline

**Date:** 2026-06-07

## Figure 1: Overall system architecture

Caption:
受约束 GraphQL 运行时数据访问控制系统总体架构示意图。

Blocks:
1. 调用方
2. 元数据获取模块
3. 权限策略获取模块
4. 访问边界生成模块
5. 访问边界暴露与查询指导模块
6. GraphQL 查询接口
7. 语义反馈模块
8. 数据访问模块
9. 底层数据库

## Figure 2: Boundary generation flow

Caption:
基于模型元数据或数据库表元数据并结合权限策略生成受约束 GraphQL 访问边界的流程示意图。

Flow:
1. 读取模型元数据或数据库表元数据
2. 提取字段、类型、关系和查询能力属性
3. 获取调用主体权限策略
4. 裁剪 GraphQL 基础访问能力
5. 生成主体专属访问边界

## Figure 3: Query correction closed loop

Caption:
查询超界后的语义反馈与修正闭环流程示意图。

Flow:
1. 调用方构造 GraphQL 查询
2. 判断是否超出访问边界
3. 若超界，返回超界位置与超界原因
4. 调用方依据反馈修正查询并重新提交
5. 返回步骤 2 重新进行访问边界判断
6. 直至查询合法后执行数据访问并返回结果

## Figure 4: Database-introspection embodiment

Caption:
直接读取数据库表元数据中的表、列、主键、唯一约束、索引和外键信息生成 GraphQL 基础访问能力的实施例示意图。

Flow:
1. 读取数据库表元数据
2. 生成表对应字段集合
3. 识别主键、唯一约束、索引和外键
4. 生成 GraphQL 基础访问能力
5. 结合权限策略形成受约束访问边界

-- ============================================
-- Migration: RLS Policy — 唯一约束从 policy_name 改为 role + action
-- Description: 保证同一 model 下同一 role + action 只有一条策略
-- ============================================

ALTER TABLE model_rls_policies
    DROP INDEX uk_policy_name,
    ADD UNIQUE KEY uk_policy_role_action (org_name, project_slug, model_id, role, action);
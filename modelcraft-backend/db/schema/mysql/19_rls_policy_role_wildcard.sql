-- Migration: RLS policy role wildcard '' → '*'
-- Replace empty-string role (legacy wildcard) with explicit '*' wildcard.
-- After this migration, role='' is rejected at the application layer.
UPDATE model_rls_policies SET role = '*' WHERE role = '';

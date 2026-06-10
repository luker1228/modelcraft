package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaContract_SoftDeleteColumnsExist(t *testing.T) {
	required := map[string][]string{
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "01_project.sql"): {
			"`projects`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "02_database_cluster.sql"): {
			"`database_clusters`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "03_model_domain.sql"): {
			"`models`", "`model_groups`", "`logical_foreign_keys`", "`model_enums`", "`field_definitions`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0", "`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "05_organizations.sql"): {
			"`organizations`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "06_users.sql"): {
			"`users`", "`profile`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "07_roles_permissions.sql"): {
			"`roles`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "12_end_user_auth.sql"): {
			"`end_user_users`", "`end_user_roles`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "13_rbac_permissions.sql"): {
			"`end_user_data_permissions`", "`end_user_permission_bundles`",
			"`deleted_at`", "BIGINT UNSIGNED NOT NULL DEFAULT 0", "`delete_token`",
		},
	}

	for file, mustContain := range required {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", file, err)
		}
		text := string(body)
		for _, want := range mustContain {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing %q", file, want)
			}
		}
	}
}

func TestSchemaContract_DeleteTokenUniqueIndexesExist(t *testing.T) {
	cases := map[string][]string{
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "01_project.sql"): {
			"PRIMARY KEY (`org_name`, `slug`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "02_database_cluster.sql"): {
			"`idx_cluster_project_unique` (`org_name`, `project_slug`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "03_model_domain.sql"): {
			"`idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`, `delete_token`)",
			"`idx_model_groups_name` (`org_name`, `project_slug`, `name`, `delete_token`)",
			"`idx_model_enums_name` (`org_name`, `project_slug`, `name`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "06_users.sql"): {
			"`uk_org_user_phone` (`org_name`, `phone`, `delete_token`)",
			"`uk_org_user_name` (`org_name`, `name`, `delete_token`)",
			"`uk_profile_user_id` (`user_id`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "07_roles_permissions.sql"): {
			"`uk_role_name_org` (`name`, `org_name`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "12_end_user_auth.sql"): {
			"`uk_end_user_users_org_username` (`org_name`, `username`, `delete_token`)",
			"`uk_end_user_roles_project_name` (`org_name`, `project_slug`, `name`, `delete_token`)",
		},
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "13_rbac_permissions.sql"): {
			"`uq_permissions_model_name`",
			"`model_id`, `name`, `delete_token`",
			"`uq_bundles_org_project_slug`",
			"`org_name`, `project_slug`, `slug`, `delete_token`",
			"`uq_bundles_org_project_name`",
			"`org_name`, `project_slug`, `name`, `delete_token`",
		},
	}

	for file, wants := range cases {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", file, err)
		}
		text := string(body)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Fatalf("%s missing %q", file, want)
			}
		}
	}
}

package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func schemaPath(filename string) string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "db", "schema", "mysql", filename)
}

func TestSchemaContract_SoftDeleteColumnsExist(t *testing.T) {
	required := map[string][]string{
		schemaPath("01_project.sql"): {
			"`projects`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		schemaPath("02_database_cluster.sql"): {
			"`database_clusters`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		schemaPath("03_model_domain.sql"): {
			"`models`", "`model_groups`", "`logical_foreign_keys`", "`model_enums`", "`field_definitions`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0", "`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		schemaPath("05_organizations.sql"): {
			"`organizations`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		schemaPath("06_users.sql"): {
			"`users`", "`profile`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
		},
		schemaPath("07_roles_permissions.sql"): {
			"`roles`",
			"`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0",
			"`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0",
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
		schemaPath("01_project.sql"): {
			"PRIMARY KEY (`org_name`, `slug`, `delete_token`)",
		},
		schemaPath("02_database_cluster.sql"): {
			"`idx_cluster_project_unique` (`org_name`, `project_slug`, `delete_token`)",
		},
		schemaPath("03_model_domain.sql"): {
			"`idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`, `delete_token`)",
			"`idx_model_groups_name` (`org_name`, `project_slug`, `name`, `delete_token`)",
			"`idx_model_enums_name` (`org_name`, `project_slug`, `name`, `delete_token`)",
		},
		schemaPath("06_users.sql"): {
			"`uk_org_user_phone` (`org_name`, `phone`, `delete_token`)",
			"`uk_org_user_name` (`org_name`, `name`, `delete_token`)",
			"`uk_profile_user_id` (`user_id`, `delete_token`)",
		},
		schemaPath("07_roles_permissions.sql"): {
			"`uk_role_name_org` (`name`, `org_name`, `delete_token`)",
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

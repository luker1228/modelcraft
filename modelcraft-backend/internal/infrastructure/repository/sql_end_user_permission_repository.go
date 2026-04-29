package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"strings"

	"github.com/google/uuid"
)

// SqlEndUserDataPermissionRepository is the sqlc-based implementation of
// rbac.EndUserPermissionRepository.
type SqlEndUserDataPermissionRepository struct {
	q dbgen.Querier
}

// NewSqlEndUserDataPermissionRepository creates a new SqlEndUserDataPermissionRepository
// backed by the given sqlc Querier.
func NewSqlEndUserDataPermissionRepository(q dbgen.Querier) rbac.EndUserPermissionRepository {
	return &SqlEndUserDataPermissionRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// compile-time interface assertion.
var _ rbac.EndUserPermissionRepository = (*SqlEndUserDataPermissionRepository)(nil)

// =========================
// Helpers
// =========================

func toDomainPermission(row dbgen.EndUserDataPermission) *rbac.EndUserPermission {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	var columnPolicy *rbac.ColumnPolicy
	if row.ColumnPolicy != nil {
		var cp rbac.ColumnPolicy
		if err := json.Unmarshal(*row.ColumnPolicy, &cp); err == nil {
			columnPolicy = &cp
		}
	}

	var rowPolicy *rbac.RowPolicy
	if row.RowPolicy != nil {
		var rp rbac.RowPolicy
		if err := json.Unmarshal(*row.RowPolicy, &rp); err == nil {
			rp.Normalize()
			rowPolicy = &rp
		}
	}

	var databaseName *string
	if row.DatabaseName.Valid {
		d := row.DatabaseName.String
		databaseName = &d
	}
	var modelName *string
	if row.ModelName.Valid {
		d := row.ModelName.String
		modelName = &d
	}

	return &rbac.EndUserPermission{
		OrgName:      row.OrgName,
		ProjectSlug:  row.ProjectSlug,
		ID:           row.ID,
		ModelID:      row.ModelID,
		DatabaseName: databaseName,
		ModelName:    modelName,
		Name:         row.Name,
		Description:  description,
		Type:         rbac.PermissionTypeCustom,
		ColumnPolicy: columnPolicy,
		RowPolicy:    rowPolicy,
		Preset:       nil,
	}
}

func toDBColumnPolicy(cp *rbac.ColumnPolicy) *json.RawMessage {
	if cp == nil {
		return nil
	}
	b, err := json.Marshal(cp)
	if err != nil {
		return nil
	}
	var raw json.RawMessage = b
	return &raw
}

func toDBRowPolicy(rp *rbac.RowPolicy) *json.RawMessage {
	if rp == nil {
		return nil
	}
	rp.Normalize()
	b, err := json.Marshal(rp)
	if err != nil {
		return nil
	}
	var raw json.RawMessage = b
	return &raw
}

// =========================
// Permissions
// =========================

func (r *SqlEndUserDataPermissionRepository) CreatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
	params := dbgen.CreateEndUserPermissionParams{
		ID:           p.ID,
		OrgName:      p.OrgName,
		ProjectSlug:  p.ProjectSlug,
		DatabaseName: sqlerr.PtrToNullStr(p.DatabaseName),
		ModelName:    sqlerr.PtrToNullStr(p.ModelName),
		ModelID:      p.ModelID,
		Name:         p.Name,
		Description:  sqlerr.PtrToNullStr(p.Description),
		ColumnPolicy: toDBColumnPolicy(p.ColumnPolicy),
		RowPolicy:    toDBRowPolicy(p.RowPolicy),
	}

	return r.q.CreateEndUserPermission(ctx, params)
}

func (r *SqlEndUserDataPermissionRepository) GetPermissionByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserPermission, error) {
	row, err := r.q.GetEndUserPermissionByID(ctx, dbgen.GetEndUserPermissionByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user permission not found: " + id)
		}
		return nil, err
	}

	return toDomainPermission(row), nil
}

func (r *SqlEndUserDataPermissionRepository) ListPermissionsByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListEndUserPermissionsByProject(ctx, dbgen.ListEndUserPermissionsByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}

func (r *SqlEndUserDataPermissionRepository) ListPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListEndUserPermissionsByModel(ctx, dbgen.ListEndUserPermissionsByModelParams{
		ModelID: modelID,
		OrgName: orgName,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}

func (r *SqlEndUserDataPermissionRepository) ListPresetPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListEndUserPermissionsByModel(ctx, dbgen.ListEndUserPermissionsByModelParams{
		ModelID: modelID,
		OrgName: orgName,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		if strings.HasPrefix(row.Name, "preset:") {
			perms = append(perms, toDomainPermission(row))
		}
	}
	return perms, nil
}

func (r *SqlEndUserDataPermissionRepository) GetPermissionByModelTypeName(
	ctx context.Context,
	orgName, modelID string,
	permissionType rbac.PermissionType,
	name string,
) (*rbac.EndUserPermission, error) {
	row, err := r.q.GetEndUserPermissionByModelAndName(ctx, dbgen.GetEndUserPermissionByModelAndNameParams{
		ModelID: modelID,
		OrgName: orgName,
		Name:    name,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user permission not found")
		}
		return nil, sqlerr.WrapSQLError(err)
	}
	if permissionType == rbac.PermissionTypePreset && !strings.HasPrefix(row.Name, "preset:") {
		return nil, shared.NewNotFoundError("end user permission not found")
	}
	return toDomainPermission(row), nil
}

func (r *SqlEndUserDataPermissionRepository) UpdatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
	params := dbgen.UpdateEndUserPermissionParams{
		Name:         p.Name,
		Description:  sqlerr.PtrToNullStr(p.Description),
		ColumnPolicy: toDBColumnPolicy(p.ColumnPolicy),
		ID:           p.ID,
		OrgName:      p.OrgName,
	}

	result, err := r.q.UpdateEndUserPermission(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user permission not found: "+p.ID)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) DeletePermission(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserPermission(ctx, dbgen.DeleteEndUserPermissionParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user permission not found: "+id)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) UpdatePresetPermission(
	ctx context.Context,
	p *rbac.EndUserPermission,
) error {
	return r.UpdatePermission(ctx, p)
}

func (r *SqlEndUserDataPermissionRepository) IsPermissionReferencedByBundle(
	ctx context.Context,
	permissionID string,
) (bool, error) {
	nullID := sql.NullString{String: permissionID, Valid: permissionID != ""}
	referenced, err := r.q.IsPermissionReferencedByBundleItem(ctx, nullID)
	if err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return referenced, nil
}

func (r *SqlEndUserDataPermissionRepository) DeletePresetPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) error {
	return nil
}

// =========================
// Bundles
// =========================

func toDomainBundle(row dbgen.EndUserPermissionBundle) *rbac.EndUserPermissionBundle {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	return &rbac.EndUserPermissionBundle{
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ID:          row.ID,
		Name:        row.Name,
		Description: description,
	}
}

func (r *SqlEndUserDataPermissionRepository) CreateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
	params := dbgen.CreateEndUserBundleParams{
		ID:          b.ID,
		OrgName:     b.OrgName,
		ProjectSlug: b.ProjectSlug,
		Name:        b.Name,
		Description: sqlerr.PtrToNullStr(b.Description),
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserBundle(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) GetBundleByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserPermissionBundle, error) {
	// projectSlug 未知时先以 org_name+id 定位；
	// 跨 project 防护由 app 层 verifyBundleScope 提供。
	rows, err := r.q.ListEndUserBundlesByProject(ctx, dbgen.ListEndUserBundlesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: "", // list 查询需要 project_slug，退而使用 fallback
	})
	_ = rows
	// 改为用单独查询（不依赖 project_slug）：
	row, err := r.q.GetEndUserBundleByID(ctx, dbgen.GetEndUserBundleByIDParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: "", // projectSlug 在 app 层校验，此处留空兼容旧签名
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user bundle not found: " + id)
		}
		return nil, err
	}

	return toDomainBundle(row), nil
}

func (r *SqlEndUserDataPermissionRepository) ListBundlesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserPermissionBundle, error) {
	rows, err := r.q.ListEndUserBundlesByProject(ctx, dbgen.ListEndUserBundlesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	bundles := make([]*rbac.EndUserPermissionBundle, 0, len(rows))
	for _, row := range rows {
		bundles = append(bundles, toDomainBundle(row))
	}
	return bundles, nil
}

func (r *SqlEndUserDataPermissionRepository) UpdateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
	params := dbgen.UpdateEndUserBundleParams{
		Name:        b.Name,
		Description: sqlerr.PtrToNullStr(b.Description),
		ID:          b.ID,
		OrgName:     b.OrgName,
	}

	result, err := r.q.UpdateEndUserBundle(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user bundle not found: "+b.ID)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) DeleteBundle(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserBundle(ctx, dbgen.DeleteEndUserBundleParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user bundle not found: "+id)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) AddPermissionToBundle(
	ctx context.Context,
	bundleID, permissionID string,
	sortOrder int,
) error {
	_ = ctx
	_ = bundleID
	_ = permissionID
	_ = sortOrder
	// add by permission_id is deprecated; use bundle item APIs
	return shared.NewRepositoryError(shared.ErrTypeUnknown, "use UpsertBundleDataPermissionItem instead")
}

func (r *SqlEndUserDataPermissionRepository) RemovePermissionFromBundle(
	ctx context.Context,
	bundleID, permissionID string,
) error {
	_ = ctx
	_ = bundleID
	_ = permissionID
	// remove by permission_id is deprecated; use RemoveBundleDataPermissionItem
	return shared.NewRepositoryError(shared.ErrTypeUnknown, "use RemoveBundleDataPermissionItem instead")
}

func (r *SqlEndUserDataPermissionRepository) ListPermissionsInBundle(
	ctx context.Context,
	bundleID string,
) ([]*rbac.EndUserPermission, error) {
	_ = ctx
	_ = bundleID
	return []*rbac.EndUserPermission{}, nil
}

func (r *SqlEndUserDataPermissionRepository) ClearBundlePermissions(ctx context.Context, bundleID string) error {
	return sqlerr.WrapSQLError(r.q.ClearBundleDataPermissionItems(ctx, bundleID))
}

func toDomainBundleDataPermissionItem(row dbgen.EndUserBundleDataPermissionItem) *rbac.EndUserBundleDataPermissionItem {
	var preset *rbac.PermissionPreset
	if row.Preset.Valid {
		v := rbac.PermissionPreset(row.Preset.EndUserBundleDataPermissionItemsPreset)
		preset = &v
	}

	var customPermissionID *string
	if row.CustomPermissionID.Valid {
		v := row.CustomPermissionID.String
		customPermissionID = &v
	}

	return &rbac.EndUserBundleDataPermissionItem{
		ID:                 row.ID,
		BundleID:           row.BundleID,
		ModelID:            row.ModelID,
		GrantType:          rbac.PermissionType(row.GrantType),
		Preset:             preset,
		CustomPermissionID: customPermissionID,
		SortOrder:          int(row.SortOrder),
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func (r *SqlEndUserDataPermissionRepository) UpsertBundleDataPermissionItem(
	ctx context.Context,
	item *rbac.EndUserBundleDataPermissionItem,
) error {
	id := item.ID
	if id == "" {
		id = uuid.NewString()
	}

	grantType := dbgen.EndUserBundleDataPermissionItemsGrantType(item.GrantType)
	var preset dbgen.NullEndUserBundleDataPermissionItemsPreset
	if item.Preset != nil {
		preset = dbgen.NullEndUserBundleDataPermissionItemsPreset{
			EndUserBundleDataPermissionItemsPreset: dbgen.EndUserBundleDataPermissionItemsPreset(*item.Preset),
			Valid:                                  true,
		}
	}

	params := dbgen.UpsertBundleDataPermissionItemParams{
		ID:                 id,
		BundleID:           item.BundleID,
		ModelID:            item.ModelID,
		GrantType:          grantType,
		Preset:             preset,
		CustomPermissionID: sqlerr.PtrToNullStr(item.CustomPermissionID),
		SortOrder:          int32(item.SortOrder),
	}
	return sqlerr.WrapSQLError(r.q.UpsertBundleDataPermissionItem(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RemoveBundleDataPermissionItem(
	ctx context.Context,
	bundleID, modelID string,
) error {
	_, err := r.q.RemoveBundleDataPermissionItem(ctx, dbgen.RemoveBundleDataPermissionItemParams{
		BundleID: bundleID,
		ModelID:  modelID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) ListBundleDataPermissionItems(
	ctx context.Context,
	bundleID string,
) ([]*rbac.EndUserBundleDataPermissionItem, error) {
	rows, err := r.q.ListBundleDataPermissionItems(ctx, bundleID)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	items := make([]*rbac.EndUserBundleDataPermissionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainBundleDataPermissionItem(row))
	}
	return items, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleDataPermissionItemByBundleAndModel(
	ctx context.Context,
	bundleID, modelID string,
) (*rbac.EndUserBundleDataPermissionItem, error) {
	params := dbgen.GetBundleDataPermissionItemByBundleAndModelParams{
		BundleID: bundleID,
		ModelID:  modelID,
	}
	row, err := r.q.GetBundleDataPermissionItemByBundleAndModel(ctx, params)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("bundle data permission item not found")
		}
		return nil, sqlerr.WrapSQLError(err)
	}
	return toDomainBundleDataPermissionItem(row), nil
}

// =========================
// Bundle Snapshots
// =========================

// snapshotRawEntry is the JSON structure stored in the snapshot `items` column.
type snapshotRawEntry struct {
	ModelID            string  `json:"modelId"`
	GrantType          string  `json:"grantType"`
	Preset             *string `json:"preset"`
	CustomPermissionID *string `json:"customPermissionId"`
	SortOrder          int     `json:"sortOrder"`
}

// parseSnapshotItems parses the raw JSON column into domain slice pairs.
func parseSnapshotItems(raw []byte) ([]rbac.SnapshotItemEntry, []rbac.SnapshotPermissionEntry) {
	var entries []snapshotRawEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, nil
	}
	items := make([]rbac.SnapshotItemEntry, 0, len(entries))
	perms := make([]rbac.SnapshotPermissionEntry, 0, len(entries))
	for _, e := range entries {
		var preset *rbac.PermissionPreset
		if e.Preset != nil && *e.Preset != "" {
			v := rbac.PermissionPreset(*e.Preset)
			preset = &v
		}
		items = append(items, rbac.SnapshotItemEntry{
			ModelID:            e.ModelID,
			GrantType:          rbac.PermissionType(e.GrantType),
			Preset:             preset,
			CustomPermissionID: e.CustomPermissionID,
			SortOrder:          e.SortOrder,
		})
		permissionID := e.ModelID
		if e.CustomPermissionID != nil && *e.CustomPermissionID != "" {
			permissionID = *e.CustomPermissionID
		}
		perms = append(perms, rbac.SnapshotPermissionEntry{
			PermissionID: permissionID,
			SortOrder:    e.SortOrder,
		})
	}
	return items, perms
}

func toDomainSnapshot(row dbgen.EndUserPermissionBundleSnapshot) rbac.BundleSnapshot {
	var createdBy *string
	if row.CreatedBy.Valid {
		v := row.CreatedBy.String
		createdBy = &v
	}
	var restoredFrom *int
	if row.RestoredFrom.Valid {
		v := int(row.RestoredFrom.Int32)
		restoredFrom = &v
	}

	var perms []rbac.SnapshotPermissionEntry
	var items []rbac.SnapshotItemEntry
	if row.Items != nil {
		items, perms = parseSnapshotItems(row.Items)
	}

	return rbac.BundleSnapshot{
		ID:           row.ID,
		BundleID:     row.BundleID,
		Version:      int(row.Version),
		Permissions:  perms,
		Items:        items,
		CreatedAt:    row.CreatedAt,
		CreatedBy:    createdBy,
		RestoredFrom: restoredFrom,
	}
}

func (r *SqlEndUserDataPermissionRepository) SaveBundleSnapshot(
	ctx context.Context,
	snapshot *rbac.BundleSnapshot,
) error {
	type entry struct {
		ModelID            string  `json:"modelId"`
		GrantType          string  `json:"grantType"`
		Preset             *string `json:"preset,omitempty"`
		CustomPermissionID *string `json:"customPermissionId,omitempty"`
		SortOrder          int     `json:"sortOrder"`
	}

	var entries []entry
	if len(snapshot.Items) > 0 {
		entries = make([]entry, 0, len(snapshot.Items))
		for _, item := range snapshot.Items {
			var preset *string
			if item.Preset != nil {
				v := string(*item.Preset)
				preset = &v
			}
			entries = append(entries, entry{
				ModelID:            item.ModelID,
				GrantType:          string(item.GrantType),
				Preset:             preset,
				CustomPermissionID: item.CustomPermissionID,
				SortOrder:          item.SortOrder,
			})
		}
	} else {
		entries = make([]entry, 0, len(snapshot.Permissions))
		for _, p := range snapshot.Permissions {
			permissionID := p.PermissionID
			entries = append(entries, entry{
				ModelID:            permissionID,
				GrantType:          string(rbac.PermissionTypeCustom),
				CustomPermissionID: &permissionID,
				SortOrder:          p.SortOrder,
			})
		}
	}
	permJSON, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	var restoredFrom sql.NullInt32
	if snapshot.RestoredFrom != nil {
		restoredFrom = sql.NullInt32{Int32: int32(*snapshot.RestoredFrom), Valid: true}
	}

	params := dbgen.InsertBundleSnapshotParams{
		ID:           snapshot.ID,
		BundleID:     snapshot.BundleID,
		Version:      int32(snapshot.Version),
		Items:        permJSON,
		CreatedBy:    sqlerr.PtrToNullStr(snapshot.CreatedBy),
		RestoredFrom: restoredFrom,
	}
	return sqlerr.WrapSQLError(r.q.InsertBundleSnapshot(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) ListBundleSnapshots(
	ctx context.Context,
	bundleID string,
) ([]rbac.BundleSnapshot, error) {
	rows, err := r.q.ListBundleSnapshots(ctx, bundleID)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	snapshots := make([]rbac.BundleSnapshot, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, toDomainSnapshot(row))
	}
	return snapshots, nil
}

func (r *SqlEndUserDataPermissionRepository) DeleteOldBundleSnapshots(
	ctx context.Context,
	bundleID string,
) error {
	return sqlerr.WrapSQLError(r.q.DeleteOldBundleSnapshots(ctx, dbgen.DeleteOldBundleSnapshotsParams{
		BundleID:   bundleID,
		BundleID_2: bundleID,
	}))
}

func (r *SqlEndUserDataPermissionRepository) GetBundleCurrentVersion(
	ctx context.Context,
	bundleID string,
) (int, error) {
	raw, err := r.q.GetBundleCurrentVersion(ctx, bundleID)
	if err != nil {
		return 0, sqlerr.WrapSQLError(err)
	}
	switch v := raw.(type) {
	case int64:
		return int(v), nil
	case []byte:
		// MySQL may return COALESCE result as []byte
		var n int64
		if _, scanErr := fmt.Sscan(string(v), &n); scanErr == nil {
			return int(n), nil
		}
	}
	return 0, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleSnapshotByVersion(
	ctx context.Context,
	bundleID string,
	version int,
) (*rbac.BundleSnapshot, error) {
	row, err := r.q.GetBundleSnapshotByVersion(ctx, dbgen.GetBundleSnapshotByVersionParams{
		BundleID: bundleID,
		Version:  int32(version),
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("bundle snapshot not found")
		}
		return nil, sqlerr.WrapSQLError(err)
	}
	snap := toDomainSnapshot(row)
	return &snap, nil
}

// =========================
// Roles
// =========================

func toDomainRole(row dbgen.EndUserRole) *rbac.EndUserRole {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	return &rbac.EndUserRole{
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ID:          row.ID,
		Name:        row.Name,
		Description: description,
		IsImplicit:  row.IsImplicit,
	}
}

func (r *SqlEndUserDataPermissionRepository) CreateRole(ctx context.Context, role *rbac.EndUserRole) error {
	params := dbgen.CreateEndUserRoleParams{
		ID:          role.ID,
		OrgName:     role.OrgName,
		ProjectSlug: role.ProjectSlug,
		Name:        role.Name,
		Description: sqlerr.PtrToNullStr(role.Description),
		IsImplicit:  role.IsImplicit,
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserRole(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) GetRoleByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserRole, error) {
	row, err := r.q.GetEndUserRoleByID(ctx, dbgen.GetEndUserRoleByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user role not found: " + id)
		}
		return nil, err
	}

	return toDomainRole(row), nil
}

func (r *SqlEndUserDataPermissionRepository) ListRolesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserRole, error) {
	rows, err := r.q.ListEndUserRolesByProject(ctx, dbgen.ListEndUserRolesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	roles := make([]*rbac.EndUserRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, toDomainRole(row))
	}
	return roles, nil
}

func (r *SqlEndUserDataPermissionRepository) UpdateRole(ctx context.Context, role *rbac.EndUserRole) error {
	params := dbgen.UpdateEndUserRoleParams{
		Name:        role.Name,
		Description: sqlerr.PtrToNullStr(role.Description),
		ID:          role.ID,
		OrgName:     role.OrgName,
	}

	result, err := r.q.UpdateEndUserRole(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user role not found: "+role.ID)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) DeleteRole(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserRole(ctx, dbgen.DeleteEndUserRoleParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user role not found: "+id)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) AssignBundleToRole(
	ctx context.Context,
	orgName, projectSlug, roleID, bundleID string,
) error {
	params := dbgen.AssignBundleToRoleParams{
		ID:          uuid.NewString(),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		RoleID:      roleID,
		BundleID:    bundleID,
	}

	return sqlerr.WrapSQLError(r.q.AssignBundleToRole(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RevokeBundleFromRole(
	ctx context.Context,
	roleID, bundleID string,
) error {
	_, err := r.q.RevokeBundleFromRole(ctx, dbgen.RevokeBundleFromRoleParams{
		RoleID:   roleID,
		BundleID: bundleID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) ListBundlesByRole(
	ctx context.Context,
	roleID string,
) ([]*rbac.EndUserPermissionBundle, error) {
	rows, err := r.q.ListBundlesByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	bundles := make([]*rbac.EndUserPermissionBundle, 0, len(rows))
	for _, row := range rows {
		bundles = append(bundles, toDomainBundle(row))
	}
	return bundles, nil
}

// =========================
// User Grants & Authz Chain
// =========================

func (r *SqlEndUserDataPermissionRepository) GrantBundleToUser(
	ctx context.Context,
	userID, orgName, projectSlug, bundleID string,
) error {
	params := dbgen.GrantBundleToUserParams{
		ID:          uuid.NewString(),
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		BundleID:    bundleID,
	}

	return sqlerr.WrapSQLError(r.q.GrantBundleToUser(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RevokeBundleFromUser(
	ctx context.Context,
	userID, orgName, projectSlug, bundleID string,
) error {
	_, err := r.q.RevokeBundleFromUser(ctx, dbgen.RevokeBundleFromUserParams{
		UserID:      userID,
		BundleID:    bundleID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) AssignRoleToUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	params := dbgen.AssignRoleToUserParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		RoleID:  roleID,
		OrgName: orgName,
	}

	return sqlerr.WrapSQLError(r.q.AssignRoleToUser(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RevokeRoleFromUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	_, err := r.q.RevokeRoleFromUser(ctx, dbgen.RevokeRoleFromUserParams{
		UserID:  userID,
		RoleID:  roleID,
		OrgName: orgName,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByUserDirect(
	ctx context.Context,
	userID, orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByUserDirect(ctx, dbgen.GetBundleIDsByUserDirectParams{
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByUserExplicitRoles(
	ctx context.Context,
	userID, orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByUserExplicitRoles(ctx, dbgen.GetBundleIDsByUserExplicitRolesParams{
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByImplicitRoles(
	ctx context.Context,
	orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByImplicitRoles(ctx, dbgen.GetBundleIDsByImplicitRolesParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetPermissionsByBundleIDs(
	ctx context.Context,
	orgName string,
	bundleIDs []string,
) ([]*rbac.EndUserPermission, error) {
	if len(bundleIDs) == 0 {
		return []*rbac.EndUserPermission{}, nil
	}

	items, err := r.q.GetDataPermissionItemsByBundleIDs(ctx, bundleIDs)
	if err != nil {
		return nil, err
	}

	customIDSet := make(map[string]struct{})
	customIDs := make([]string, 0)
	for _, item := range items {
		if !item.CustomPermissionID.Valid || item.CustomPermissionID.String == "" {
			continue
		}
		if _, ok := customIDSet[item.CustomPermissionID.String]; ok {
			continue
		}
		customIDSet[item.CustomPermissionID.String] = struct{}{}
		customIDs = append(customIDs, item.CustomPermissionID.String)
	}
	if len(customIDs) == 0 {
		return []*rbac.EndUserPermission{}, nil
	}

	rows, err := r.q.GetCustomPermissionsByIDs(ctx, dbgen.GetCustomPermissionsByIDsParams{
		Permissionids: customIDs,
		OrgName:       orgName,
	})
	if err != nil {
		return nil, err
	}

	permByID := make(map[string]*rbac.EndUserPermission, len(rows))
	for _, row := range rows {
		permByID[row.ID] = toDomainPermission(row)
	}

	perms := make([]*rbac.EndUserPermission, 0, len(items))
	for _, item := range items {
		if !item.CustomPermissionID.Valid {
			continue
		}
		if p, ok := permByID[item.CustomPermissionID.String]; ok {
			perms = append(perms, p)
		}
	}
	return perms, nil
}

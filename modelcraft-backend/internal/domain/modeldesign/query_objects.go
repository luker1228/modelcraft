package modeldesign

import bizerrors "modelcraft/pkg/bizerrors"

// ModelQuery Domain层的模型查询对象
type ModelQuery struct {
	OrgName      string `json:"orgName"`     // 组织名称（用于多租户数据隔离）
	ProjectSlug  string `json:"projectSlug"` // 项目标识符
	DatabaseName string `json:"databaseName"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	StorageType  string `json:"storageType"`
	Page         int    `json:"page"`
	PageSize     int    `json:"pageSize"`
}

// FieldQuery Domain层的字段查询对象
type FieldQuery struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Nullable *bool  `json:"nullable"`
}

// Validate 验证查询对象
func (q *ModelQuery) Validate() error {
	if q.OrgName == "" {
		return bizerrors.Errorf("OrgName cannot be blank")
	}
	if q.ProjectSlug == "" {
		return bizerrors.Errorf("ProjectSlug cant be blank")
	}
	if q.DatabaseName == "" {
		return bizerrors.Errorf("DatabaseName cant be blank")
	}
	if q.PageSize < 0 {
		return bizerrors.Errorf("PageSize cant be less than 0")
	}
	if q.Page < 0 {
		return bizerrors.Errorf("Page cant be less than 0")
	}
	return nil
}

// SetDefaults 设置默认值
func (q *ModelQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
}

package modeldesign

import (
	"fmt"
	"modelcraft/internal/domain/project"
	"regexp"
	"strings"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// enumNamePattern matches valid enum names: must start with a letter, only letters, digits, and underscores.
var enumNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// EnumOption 枚举选项值对象
type EnumOption struct {
	Code        string `json:"code"`                  // 枚举code,必须非空且唯一
	Label       string `json:"label"`                 // 枚举显示标签,必须非空
	Order       int32  `json:"order"`                 // 显示排序
	Description string `json:"description,omitempty"` // 选项描述
}

// Validate 验证枚举选项
func (eo *EnumOption) Validate() error {
	if strings.TrimSpace(eo.Code) == "" {
		return bizerrors.New("enum option code cannot be empty")
	}
	if strings.TrimSpace(eo.Label) == "" {
		return bizerrors.New("enum option label cannot be empty")
	}
	return nil
}

// String 实现Stringer接口
func (eo *EnumOption) String() string {
	return fmt.Sprintf("EnumOption(%s:%s)", eo.Code, eo.Label)
}

// EnumDefinition 枚举定义聚合根
type EnumDefinition struct {
	ID                   string       `json:"id"`
	project.ProjectScope              // 嵌入项目作用域，包含 OrgName 和 ProjectSlug
	Name                 string       `json:"name"`          // 英文标识,项目内唯一
	DisplayName          string       `json:"displayName"`   // 显示名称
	Description          string       `json:"description"`   // 描述
	Options              []EnumOption `json:"options"`       // 枚举选项列表
	IsMultiSelect        bool         `json:"isMultiSelect"` // 是否多选
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
}

// Validate 验证枚举定义
func (ed *EnumDefinition) Validate() error {
	if err := ed.ProjectScope.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(ed.Name) == "" {
		return bizerrors.New("enum name cannot be empty")
	}
	if !enumNamePattern.MatchString(ed.Name) {
		return bizerrors.Errorf(
			"enum name %q is invalid: must start with a letter and contain only letters, digits, and underscores",
			ed.Name,
		)
	}
	if strings.TrimSpace(ed.DisplayName) == "" {
		return bizerrors.New("enum display name cannot be empty")
	}
	if len(ed.Options) == 0 {
		return bizerrors.New("enum must have at least one option")
	}

	// 验证选项
	codeSet := make(map[string]bool)
	for _, opt := range ed.Options {
		if err := opt.Validate(); err != nil {
			return bizerrors.Wrapf(err, "invalid enum option")
		}
		if codeSet[opt.Code] {
			return bizerrors.Errorf("duplicate enum option code: %s", opt.Code)
		}
		codeSet[opt.Code] = true
	}

	return nil
}

// GetOptionByCode 根据code获取枚举选项
func (ed *EnumDefinition) GetOptionByCode(code string) (*EnumOption, error) {
	for _, opt := range ed.Options {
		if opt.Code == code {
			return &opt, nil
		}
	}
	return nil, bizerrors.Errorf("enum option not found: %s", code)
}

// HasOptionCode 检查是否包含指定code的选项
func (ed *EnumDefinition) HasOptionCode(code string) bool {
	for _, opt := range ed.Options {
		if opt.Code == code {
			return true
		}
	}
	return false
}

// ValidateCodes 验证一组code是否都在枚举选项中
func (ed *EnumDefinition) ValidateCodes(codes []string) error {
	for _, code := range codes {
		if !ed.HasOptionCode(code) {
			return bizerrors.Errorf("invalid enum code: %s", code)
		}
	}
	return nil
}

// Update 更新枚举定义
func (ed *EnumDefinition) Update(displayName, description *string, options []EnumOption) error {
	if displayName != nil {
		if strings.TrimSpace(*displayName) == "" {
			return bizerrors.New("enum display name cannot be empty")
		}
		ed.DisplayName = *displayName
	}
	if description != nil {
		ed.Description = *description
	}
	if options != nil {
		if len(options) == 0 {
			return bizerrors.New("enum must have at least one option")
		}

		// 验证新选项
		codeSet := make(map[string]bool)
		for _, opt := range options {
			if err := opt.Validate(); err != nil {
				return bizerrors.Wrapf(err, "invalid enum option")
			}
			if codeSet[opt.Code] {
				return bizerrors.Errorf("duplicate enum option code: %s", opt.Code)
			}
			codeSet[opt.Code] = true
		}

		ed.Options = options
	}

	ed.UpdatedAt = time.Now()
	return nil
}

// Clone 深拷贝枚举定义
func (ed *EnumDefinition) Clone() *EnumDefinition {
	optionsCopy := make([]EnumOption, len(ed.Options))
	copy(optionsCopy, ed.Options)

	return &EnumDefinition{
		ID:            ed.ID,
		ProjectScope:  ed.ProjectScope,
		Name:          ed.Name,
		DisplayName:   ed.DisplayName,
		Description:   ed.Description,
		Options:       optionsCopy,
		IsMultiSelect: ed.IsMultiSelect,
		CreatedAt:     ed.CreatedAt,
		UpdatedAt:     ed.UpdatedAt,
	}
}

// String 实现Stringer接口
func (ed *EnumDefinition) String() string {
	return fmt.Sprintf("Enum(%s:%s)", ed.Name, ed.DisplayName)
}

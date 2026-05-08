package sqlsoftdelete

import (
	"strings"
)

type Annotations struct {
	IncludeDeleted bool
	OnlyDeleted    bool
	PhysicalDelete bool
}

func ParseAnnotations(src []byte) Annotations {
	if len(src) == 0 {
		return Annotations{}
	}

	text := strings.ToLower(string(src))
	return Annotations{
		IncludeDeleted: strings.Contains(text, "@include_deleted"),
		OnlyDeleted:    strings.Contains(text, "@only_deleted"),
		PhysicalDelete: strings.Contains(text, "@physical_delete"),
	}
}

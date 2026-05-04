package endusergraphql

import (
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"time"
)

func toTuser(dto *appEnduser.MetaUserDTO) *generated.Tuser {
	if dto == nil {
		return nil
	}
	return &generated.Tuser{
		ID:        dto.ID,
		Username:  dto.Username,
		CreatedAt: dto.CreatedAt.Format(time.RFC3339Nano),
	}
}

func buildMetaUserFilter(where *generated.TuserWhereInput) *appEnduser.MetaUserFindManyFilter {
	f := &appEnduser.MetaUserFindManyFilter{}
	if where.ID != nil {
		if where.ID.Eq != nil {
			eq := string(*where.ID.Eq)
			f.IDEq = &eq
		}
		for _, v := range where.ID.In {
			f.IDIn = append(f.IDIn, string(v))
		}
	}
	if where.Username != nil {
		f.UsernameEq = where.Username.Eq
		f.UsernameContains = where.Username.Contains
		f.UsernameStartsWith = where.Username.StartsWith
		f.UsernameIn = where.Username.In
	}
	if where.CreatedAt != nil {
		f.CreatedAtEq = where.CreatedAt.Eq
		f.CreatedAtGte = where.CreatedAt.Gte
		f.CreatedAtLte = where.CreatedAt.Lte
	}
	return f
}

func metaUserInvalidInputResult(requestID string, start time.Time, _ string) *generated.TuserFindOneResult {
	return &generated.TuserFindOneResult{
		Item:     nil,
		TimeCost: int32(time.Since(start).Milliseconds()),
		ReqID:    requestID,
	}
}

func mapMetaUserError(requestID string, start time.Time, _ error) *generated.TuserFindOneResult {
	return &generated.TuserFindOneResult{
		Item:     nil,
		TimeCost: int32(time.Since(start).Milliseconds()),
		ReqID:    requestID,
	}
}

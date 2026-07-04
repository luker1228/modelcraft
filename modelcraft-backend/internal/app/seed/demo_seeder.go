package seed

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"os"

	domainauth "modelcraft/internal/domain/auth"
	domainOrg "modelcraft/internal/domain/organization"
	domainPerm "modelcraft/internal/domain/permission"

	domainUser "modelcraft/internal/domain/user"
)

// DemoOrgName is the fixed org slug for the demo environment.
const DemoOrgName = "demo"

// DemoSeeder idempotently initialises the demo org, owner user, and Guest role binding.
// It is a no-op unless DEMO_ENABLED=true.
type DemoSeeder struct {
	orgRepo      domainOrg.OrganizationRepository
	userRepo     domainUser.UserRepository
	roleRepo     domainPerm.RoleRepository
	userRoleRepo domainPerm.UserRoleRepository
	hasher       domainauth.PasswordHasher
}

func NewDemoSeeder(
	orgRepo domainOrg.OrganizationRepository,
	userRepo domainUser.UserRepository,
	roleRepo domainPerm.RoleRepository,
	userRoleRepo domainPerm.UserRoleRepository,
	hasher domainauth.PasswordHasher,
) *DemoSeeder {
	return &DemoSeeder{
		orgRepo:      orgRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		hasher:       hasher,
	}
}

// Seed creates the demo org and owner user if they do not already exist.
func (s *DemoSeeder) Seed(ctx context.Context) error {
	if os.Getenv("DEMO_ENABLED") != "true" {
		return nil
	}

	logger := logfacade.GetLogger(ctx)

	ownerUserName := os.Getenv("DEMO_OWNER_USERNAME")
	ownerPassword := os.Getenv("DEMO_OWNER_PASSWORD")
	if ownerUserName == "" || ownerPassword == "" {
		logger.Warnf(ctx, "DEMO_ENABLED=true but DEMO_OWNER_USERNAME or DEMO_OWNER_PASSWORD not set, "+
			"skipping demo seed")
		return nil
	}

	// 1. Idempotent: skip if demo org already exists
	existingOrg, err := s.orgRepo.GetByName(ctx, DemoOrgName)
	if err != nil && !shared.IsNotFoundError(err) {
		return bizerrors.Wrapf(err, "check demo org")
	}

	var ownerUserID string

	if existingOrg == nil {
		id, err := s.createDemoOrgAndOwner(ctx, ownerUserName, ownerPassword)
		if err != nil {
			return err
		}
		ownerUserID = id
	} else {
		ownerUser, err := s.userRepo.GetByName(ctx, DemoOrgName, ownerUserName)
		if err != nil {
			return bizerrors.Wrapf(err, "get demo owner user")
		}
		ownerUserID = ownerUser.ID
	}

	// 2. Bind owner user to owner system role (idempotent: AssignRole ignores duplicates)
	ownerRole, err := s.roleRepo.GetRoleByNameAndOrg(ctx, domainPerm.RoleOwner, domainPerm.SystemOrgName)
	if err != nil || ownerRole == nil {
		logfacade.GetLogger(ctx).Warnf(ctx,
			"Owner system role not found, skipping role binding (run SystemRolePermissionsSyncer first)")
		return nil
	}
	_ = s.userRoleRepo.AssignRole(ctx, domainPerm.NewUserRole(ownerUserID, ownerRole.ID, DemoOrgName))

	logfacade.GetLogger(ctx).Infof(ctx, "Demo seed completed")
	return nil
}

func (s *DemoSeeder) createDemoOrgAndOwner(ctx context.Context, ownerUserName, ownerPassword string) (string, error) {
	logger := logfacade.GetLogger(ctx)

	userID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return "", bizerrors.Wrapf(err, "generate demo owner user id")
	}

	org := &domainOrg.Organization{
		Name:        DemoOrgName,
		DisplayName: "Demo",
		OwnerID:     userID,
		Status:      domainOrg.OrgStatusActive,
	}
	if err := s.orgRepo.Create(ctx, org); err != nil {
		return "", bizerrors.Wrapf(err, "create demo org")
	}
	logger.Infof(ctx, "Demo org created: org_name=%s", DemoOrgName)

	hashedPwd, err := s.hasher.Hash(ctx, ownerPassword)
	if err != nil {
		return "", bizerrors.Wrapf(err, "hash demo owner password")
	}
	owner, err := domainUser.NewUser(userID, ownerUserName, domainUser.PhoneNumber{}, hashedPwd, DemoOrgName)
	if err != nil {
		return "", bizerrors.Wrapf(err, "create demo owner user entity")
	}
	if err := s.userRepo.Create(ctx, owner); err != nil {
		return "", bizerrors.Wrapf(err, "save demo owner user")
	}
	logger.Infof(ctx, "Demo owner user created: user_name=%s", ownerUserName)
	return userID, nil
}

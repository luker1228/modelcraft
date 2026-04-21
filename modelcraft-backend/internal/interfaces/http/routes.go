package http

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"modelcraft/internal/app/auth"
	"modelcraft/internal/app/cluster"
	"modelcraft/internal/app/modeldesign"
	"modelcraft/internal/app/modelruntime"
	"modelcraft/internal/app/project"
	"modelcraft/internal/app/rls"
	"modelcraft/internal/infrastructure/database/ddl"
	"modelcraft/internal/infrastructure/dbgen"
	rlsRepo "modelcraft/internal/infrastructure/persistence/rls"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/internal/middleware"
	"modelcraft/pkg/config"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"os"
	"sync"
	"time"

	endusergraphql "modelcraft/internal/interfaces/graphql/enduser"
	orggraphql "modelcraft/internal/interfaces/graphql/org"
	projectgraphql "modelcraft/internal/interfaces/graphql/project"

	runtimeHandler "modelcraft/internal/interfaces/runtime"

	appOrg "modelcraft/internal/app/organization"
	appPermission "modelcraft/internal/app/permission"
	appProfile "modelcraft/internal/app/profile"

	appEnduser "modelcraft/internal/app/enduser"
	appRole "modelcraft/internal/app/role"
	domainEndUser "modelcraft/internal/domain/enduser"
	domainModelDesign "modelcraft/internal/domain/modeldesign"
	domainUser "modelcraft/internal/domain/user"
	infraAuth "modelcraft/internal/infrastructure/auth"

	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	enduserHandlers "modelcraft/internal/interfaces/http/handlers/enduser"
	userHandlers "modelcraft/internal/interfaces/http/handlers/user"

	"github.com/go-chi/chi/v5"
)

// DesignHandlers holds all handlers and services needed for the design-time API.
// HTTP design handlers (Project, Model, Cluster, Enum) have been removed;
// business domain APIs are served exclusively via GraphQL.
// This struct provides the AuthHandler for Chi and app services for GraphQL.
type DesignHandlers struct {
	AuthHandler *authHandlers.Handler
	UserHandler *userHandlers.Handler

	// Services needed for GraphQL setup
	ModelAppService           *modeldesign.ModelDesignAppService
	ClusterAppService         *cluster.DatabaseClusterAppService
	ReverseEngineerAppService *modeldesign.ReverseEngineerAppService
	EnumAppService            *modeldesign.EnumAppService
	ProjectAppService         *project.ProjectAppService
	OrgAppService             *appOrg.OrganizationAppService
	ProfileAppService         *appProfile.AppService
	RoleAppService            *appRole.RoleAppService
	GroupAppService           *modeldesign.ModelGroupAppService
	LogicalFKAppService       *modeldesign.LogicalFKAppService
	RLSPolicyAppService       *rls.ModelRLSPolicyAppService
	AuthSchemaAppService      *rls.AuthSchemaAppService

	// Casbin Permission Services
	PermRoleService       *appPermission.RoleService
	PermPermissionService *appPermission.PermissionService
	PermUserRoleService   *appPermission.UserRoleService

	// Auth Services
	APIKeyService *auth.APIKeyService

	// Repositories
	ModelRepository domainModelDesign.ModelRepository
	UserRepo        domainUser.UserRepository
	ClusterManager  *repository.ClusterConnectionManager

	// End-User Services
	EndUserAuthAppService *appEnduser.EndUserAuthAppService
	EndUserMgmtAppService *appEnduser.EndUserManagementAppService
	EndUserAuthHandler    *enduserHandlers.AuthHandler
	EndUserMgmtHandler    *enduserHandlers.ManagementHandler
	EndUserDataHandler    *enduserHandlers.DataHandler
}

// endUserAuthRepositoryFactory creates end-user repositories from a DB connection.
type endUserAuthRepositoryFactory struct{}

func (f *endUserAuthRepositoryFactory) NewEndUserRepository(db appEnduser.SQLDBTX) domainEndUser.EndUserRepository {
	return repository.NewSqlEndUserRepository(db)
}

func (f *endUserAuthRepositoryFactory) NewEndUserSessionRepository(
	db appEnduser.SQLDBTX,
) domainEndUser.EndUserSessionRepository {
	return repository.NewSqlEndUserSessionRepository(db)
}

// endUserTxManager provides real SQL transaction support on private DBs.
type endUserTxManager struct{}

func (m *endUserTxManager) WithTx(
	ctx context.Context,
	db *sql.DB,
	fn func(ctx context.Context, txDB appEnduser.SQLDBTX) error,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin private transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback private transaction failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit private transaction: %w", err)
	}

	return nil
}

// CreateDesignHandlers creates all handlers and services needed for the design API.
// Route registration is handled by Chi in chi_setup.go (auth/org/webhook only).
// Business domain APIs are served via GraphQL, not REST.
func CreateDesignHandlers( //nolint:funlen // wiring entrypoint intentionally constructs all services in one place
	repoFactory *repository.ConnectionFactory,
	cfg *config.Config,
) (*DesignHandlers, error) {
	// Wrap the raw *sql.DB with a logging layer so all sqlc queries are traced.
	loggingDB := repository.NewSqlcLogger(repoFactory.SqlDB, repository.SqlcLogInfo, 200*time.Millisecond)

	modelRepository := repository.NewSqlModelDesignRepository(dbgen.New(loggingDB))

	clusterManager := repository.DefaultClusterManager
	// Create database cluster related services
	clusterRepository := repository.NewSqlDatabaseClusterRepository(dbgen.New(loggingDB))
	clusterAppService := cluster.NewDatabaseClusterAppService(clusterRepository, clusterManager)

	// Create model related services
	txManager := repository.NewSqlTxManager(repoFactory.SqlDB)
	enumAssocRepository := repository.NewSqlFieldEnumAssociationRepository(dbgen.New(loggingDB))
	fkRepository := repository.NewSqlLogicalForeignKeyRepository(dbgen.New(loggingDB))
	enumRepository := repository.NewSqlEnumRepository(dbgen.New(loggingDB))
	appService := modeldesign.NewModelDesignAppService(modeldesign.ModelDesignAppServiceDeps{
		DeployRepo:    ddl.NewDeploymentService(clusterManager),
		ModelRepo:     modelRepository,
		FKRepo:        fkRepository,
		ClusterRepo:   clusterRepository,
		TxManager:     txManager,
		EnumAssocRepo: enumAssocRepository,
		EnumRepo:      enumRepository,
	})

	// Create logical FK app service
	logicalFKAppService := modeldesign.NewLogicalFKAppService(fkRepository, modelRepository, txManager)

	// Create reverse engineering service
	reverseEngineerApp := modeldesign.NewReverseEngineerAppService(
		appService,
		clusterManager,
		clusterRepository,
		modelRepository,
	)

	// Create project related services
	projectRepository := repository.NewSqlProjectRepository(dbgen.New(loggingDB))
	projectAppService := project.NewProjectAppService(
		projectRepository, clusterRepository, clusterAppService, txManager,
	)

	// Create enum related services
	enumAppService := modeldesign.NewEnumAppService(enumRepository, projectRepository)

	// Create model group related services
	groupRepository := repository.NewSqlModelGroupRepository(dbgen.New(loggingDB))
	groupAppService := modeldesign.NewModelGroupAppService(groupRepository, modelRepository, txManager)

	// Create RLS related services
	modelRLSPolicyRepo := rlsRepo.NewSqlModelRLSPolicyRepository(dbgen.New(loggingDB))
	authSchemaRepo := rlsRepo.NewSqlAuthSchemaRepository(dbgen.New(loggingDB))
	rlsPolicyValidator := rls.NewPolicyValidator()
	rlsPolicyAppService := rls.NewModelRLSPolicyAppService(
		modelRLSPolicyRepo,
		modelRepository,
		authSchemaRepo,
		rlsPolicyValidator,
	)
	authSchemaAppService := rls.NewAuthSchemaAppService(authSchemaRepo, projectRepository)

	// Create user management related services
	userRepo := repository.NewSqlUserRepository(dbgen.New(loggingDB))
	profileRepo := repository.NewSqlProfileRepository(dbgen.New(loggingDB))
	profileAppService := appProfile.NewAppService(userRepo, profileRepo)
	orgRepo := repository.NewSqlOrganizationRepository(dbgen.New(loggingDB))
	membershipRepo := repository.NewSqlMembershipRepository(dbgen.New(loggingDB))

	// ============================================================
	// Create Casbin Permission Services
	// ============================================================
	casbinRoleRepo := repository.NewSqlCasbinRoleRepository(dbgen.New(loggingDB))
	casbinPermRepo := repository.NewSqlCasbinPermissionRepository(dbgen.New(loggingDB))
	casbinUserRoleRepo := repository.NewSqlCasbinUserRoleRepository(dbgen.New(loggingDB))

	permRoleService := appPermission.NewRoleService(casbinRoleRepo, casbinPermRepo, casbinUserRoleRepo)
	permPermissionService := appPermission.NewPermissionService(casbinRoleRepo, casbinPermRepo)
	permUserRoleService := appPermission.NewUserRoleService(
		casbinRoleRepo,
		casbinUserRoleRepo,
		casbinPermRepo,
		nil, // TODO: Add Redis-based version manager in Phase 3.3
	)

	// Create organization and role services
	// Note: OrganizationAppService uses Casbin permission system
	orgAppService := appOrg.NewOrganizationAppService(
		orgRepo,
		userRepo,
		membershipRepo,
		casbinRoleRepo,
		permUserRoleService,
	)

	// RoleAppService using sqlc-based repository
	roleRepo := repository.NewSqlRoleRepository(dbgen.New(loggingDB))
	roleAppService := appRole.NewRoleAppService(roleRepo)

	// ============================================================
	// Create TokenService for Authentication
	// ============================================================
	logger := logfacade.GetLogger(context.Background())

	refreshTokenRepo := repository.NewSqlRefreshTokenRepository(dbgen.New(loggingDB))
	auditLogRepo := repository.NewSqlSecurityAuditLogRepository(dbgen.New(loggingDB))

	// Create organization service for user registration
	createOrgService := appOrg.NewCreateOrganizationService(
		txManager,
		userRepo,
		orgRepo,
		casbinRoleRepo,
		membershipRepo,
	)

	passwordHasher := infraAuth.NewBcryptPasswordHasher()
	tokenService := auth.NewTokenService(
		refreshTokenRepo,
		userRepo,
		profileRepo,
		auditLogRepo,
		passwordHasher,
		7*24*time.Hour, // refresh token TTL
		createOrgService,
		membershipRepo, // for fetching user's org on login
		txManager,
	)

	// Create auth handler with token service
	authHandler := authHandlers.NewHandler(tokenService, logger)

	// Create API key service
	apiKeyRepo := repository.NewSqlAPIKeyRepository(dbgen.New(loggingDB))
	apiKeyService := auth.NewAPIKeyService(apiKeyRepo)

	// Create end-user services and handlers
	privateDBManager := repository.NewPrivateDBManager(clusterManager, &cfg.Database, logger)

	// Wire private DB provisioner into project service so mc_private_{slug} is created on project creation.
	projectAppService.WithPrivateDBProvisioner(privateDBManager)
	// Wire model importer so users/accounts tables are imported as models after private DB is provisioned.
	projectAppService.WithPrivateModelImporter(reverseEngineerApp)
	endUserTxMgr := &endUserTxManager{}
	endUserAuthAppService := appEnduser.NewEndUserAuthAppService(
		privateDBManager,
		&endUserAuthRepositoryFactory{},
		endUserTxMgr,
		logger,
	)
	endUserMgmtAppService := appEnduser.NewEndUserManagementAppService(
		appEnduser.NewPrivateDBManagerAdapter(privateDBManager),
		endUserTxMgr,
	)
	endUserAuthHandler := enduserHandlers.NewAuthHandler(endUserAuthAppService, logger)
	endUserMgmtHandler := enduserHandlers.NewManagementHandler(endUserMgmtAppService, logger)
	endUserDataHandler := enduserHandlers.NewDataHandler(appService, logger)

	return &DesignHandlers{
		AuthHandler:               authHandler,
		UserHandler:               userHandlers.NewHandler(membershipRepo, logger),
		ModelAppService:           appService,
		ClusterAppService:         clusterAppService,
		ReverseEngineerAppService: reverseEngineerApp,
		EnumAppService:            enumAppService,
		ProjectAppService:         projectAppService,
		OrgAppService:             orgAppService,
		ProfileAppService:         profileAppService,
		RoleAppService:            roleAppService,
		PermRoleService:           permRoleService,
		PermPermissionService:     permPermissionService,
		PermUserRoleService:       permUserRoleService,
		ModelRepository:           modelRepository,
		UserRepo:                  userRepo,
		ClusterManager:            clusterManager,
		GroupAppService:           groupAppService,
		LogicalFKAppService:       logicalFKAppService,
		RLSPolicyAppService:       rlsPolicyAppService,
		AuthSchemaAppService:      authSchemaAppService,
		APIKeyService:             apiKeyService,
		EndUserAuthAppService:     endUserAuthAppService,
		EndUserMgmtAppService:     endUserMgmtAppService,
		EndUserAuthHandler:        endUserAuthHandler,
		EndUserMgmtHandler:        endUserMgmtHandler,
		EndUserDataHandler:        endUserDataHandler,
	}, nil
}

// SetupOrgGraphQLRoutesOnChi registers GraphQL endpoints for org domain.
// Route pattern: /graphql/org/{orgName}/
func SetupOrgGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	// Create org resolver with only org-domain services
	orgResolver := &orggraphql.Resolver{
		ProjectAppService:      handlers.ProjectAppService,
		ClusterAppService:      handlers.ClusterAppService,
		AuthSchemaAppService:   handlers.AuthSchemaAppService,
		OrganizationAppService: handlers.OrgAppService,
		ProfileAppService:      handlers.ProfileAppService,
		UserRepo:               handlers.UserRepo,
		RoleAppService:         handlers.RoleAppService,
		RoleService:            handlers.PermRoleService,
		PermissionService:      handlers.PermPermissionService,
		UserRoleService:        handlers.PermUserRoleService,
		APIKeyService:          handlers.APIKeyService,
	}

	jwtConfig := &middleware.JWTAuthConfig{
		ModelCraftSecret: []byte(cfg.JWT.Secret),
		SkipValidation:   cfg.Auth.Design.SkipJWTValidation,
	}
	router.Route("/graphql/org/{orgName}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Post("/", orggraphql.OrgGraphQLHandler(orgResolver))
		r.Get("/", orggraphql.OrgPlaygroundHandler())
	})
}

// SetupProjectGraphQLRoutesOnChi registers GraphQL endpoints for project domain.
// Route pattern: /graphql/org/{orgName}/project/{projectSlug}/
func SetupProjectGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	projectgraphql.SetEndUserManagementAppService(handlers.EndUserMgmtAppService)

	// Create services needed for project domain
	typeMapper := domainModelDesign.NewMySQLTypeMapper()
	schemaComparisonService := domainModelDesign.NewMySQLSchemaComparisonService(typeMapper)
	deployRepo := ddl.NewDeploymentService(handlers.ClusterManager)
	repairUseCase := modeldesign.NewRepairModelUseCase(
		handlers.ModelRepository,
		handlers.ClusterManager,
		deployRepo,
		schemaComparisonService,
	)

	actualSchemaService := ddl.NewActualSchemaService()
	actualSchemaQueryUseCase := modeldesign.NewActualSchemaQueryUseCase(actualSchemaService, handlers.ClusterManager)

	// Create project resolver
	projectResolver := &projectgraphql.Resolver{
		ClusterAppService:        handlers.ClusterAppService,
		ModelDesignService:       handlers.ModelAppService,
		ReverseEngineerService:   handlers.ReverseEngineerAppService,
		RepairModelUseCase:       repairUseCase,
		ActualSchemaQueryUseCase: actualSchemaQueryUseCase,
		GroupAppService:          handlers.GroupAppService,
		LogicalFKAppService:      handlers.LogicalFKAppService,
		EnumAppService:           handlers.EnumAppService,
		UserRoleService:          handlers.PermUserRoleService,
		FieldSelectionChecker:    projectgraphql.NewFieldSelectionChecker(),
		RLSPolicyAppService:      handlers.RLSPolicyAppService,
		AuthSchemaAppService:     handlers.AuthSchemaAppService,
	}

	jwtConfig := &middleware.JWTAuthConfig{
		ModelCraftSecret: []byte(cfg.JWT.Secret),
		SkipValidation:   cfg.Auth.Design.SkipJWTValidation,
	}

	// Register project endpoint: /graphql/org/{orgName}/project/{projectSlug}
	router.Route("/graphql/org/{orgName}/project/{projectSlug}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLProjectMiddleware())
		r.Post("/", projectgraphql.ProjectGraphQLHandler(projectResolver))
		r.Get("/", projectgraphql.ProjectPlaygroundHandler())
	})
}

// LoadRSAPublicKey loads the RSA public key for design API JWT validation.
// It supports two sources (in priority order):
// 1. PEM file path (AUTH_JWT_PUBLIC_KEY_PATH)
// 2. Direct public key string (AUTH_JWT_PUBLIC_KEY)
// This function is exported for use by both Gin and Chi middleware setup.
// rsaPublicKeyOnce ensures LoadRSAPublicKey only parses the key once.
var (
	rsaPublicKeyOnce             sync.Once
	rsaPublicKeyCache            *rsa.PublicKey
	errRSAPublicKeyNotConfigured = errors.New("rsa public key not configured")
)

func LoadRSAPublicKey(cfg *config.Config) *rsa.PublicKey {
	rsaPublicKeyOnce.Do(func() {
		key, err := loadRSAPublicKey(cfg)
		if err != nil && !errors.Is(err, errRSAPublicKeyNotConfigured) {
			logfacade.GetLogger(context.Background()).
				Warnf(context.Background(), "failed to load RSA public key: %v", err)
		}
		rsaPublicKeyCache = key
	})
	return rsaPublicKeyCache
}

func loadRSAPublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	logger := logfacade.GetLogger(context.Background())

	var pemData []byte

	// Priority 1: Public key file path
	if cfg.Auth.Design.JWTPublicKeyPath != "" {
		data, err := os.ReadFile(cfg.Auth.Design.JWTPublicKeyPath)
		if err != nil {
			logger.Errorf(context.Background(),
				"Failed to read JWT public key file %s: %v", cfg.Auth.Design.JWTPublicKeyPath, err)
		} else {
			pemData = data
			logger.Infof(context.Background(), "Loading RSA public key from file: %s", cfg.Auth.Design.JWTPublicKeyPath)
		}
	}

	// Priority 2: Direct public key string
	if pemData == nil && cfg.Auth.Design.JWTPublicKey != "" {
		pemData = []byte(cfg.Auth.Design.JWTPublicKey)
		logger.Infof(context.Background(), "Loading RSA public key from config string")
	}

	if pemData == nil {
		logger.Infof(
			context.Background(), "No RSA public key configured, JWT signature verification will fail "+
				"unless SkipValidation is enabled",
		)
		return nil, errRSAPublicKeyNotConfigured
	}

	// Try parsing as X.509 certificate first
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from certificate/key data")
	}

	// Try X.509 certificate
	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
		}
		rsaKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate public key is not RSA")
		}
		logger.Infof(context.Background(),
			"RSA public key loaded from X.509 certificate, key size: %d bits", rsaKey.N.BitLen())
		return rsaKey, nil
	}

	// Try PKIX public key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}
	rsaKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}
	logger.Infof(context.Background(), "RSA public key loaded, key size: %d bits", rsaKey.N.BitLen())
	return rsaKey, nil
}

// RuntimeHandlers holds the handler needed for the model runtime GraphQL API.
type RuntimeHandlers struct {
	ModelRuntimeHandler *runtimeHandler.ModelRuntimeHandler
}

// CreateRuntimeHandlers initialises repository, application service, and handler
// for the model runtime API.
func CreateRuntimeHandlers(loggingDB dbgen.Querier) *RuntimeHandlers {
	modelRuntimeRepo := repository.NewSqlModelRuntimeRepository(loggingDB)
	lfkRepo := repository.NewSqlLogicalForeignKeyRepository(loggingDB)
	graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo)
	handler := runtimeHandler.NewModelRuntimeHandler(graphqlAppService)
	// TODO: Create and wire RLSResolver with policy repository
	// rlsPolicyRepo := repository.NewSqlModelRLSPolicyRepository(loggingDB)
	// rlsResolver := runtimeHandler.NewRLSResolver(logfacade.GetLogger(), rlsPolicyRepo)
	return &RuntimeHandlers{ModelRuntimeHandler: handler}
}

// SetupRuntimeGraphQLRoutesOnChi registers the runtime GraphQL routes on the Chi router.
// Routes:
//
//	GET  /graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model} → GraphQL Playground
//	POST /graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model} → GraphQL query execution
//
// JWT authentication is enforced when cfg.Auth.Runtime.Enabled is true.
func SetupRuntimeGraphQLRoutesOnChi(router chi.Router, handlers *RuntimeHandlers, cfg *config.Config) {
	jwtConfig := &middleware.JWTAuthConfig{
		ModelCraftSecret: []byte(cfg.JWT.Secret),
		SkipValidation:   !cfg.Auth.Runtime.Enabled,
	}

	runtimeMW := func(next http.Handler) http.Handler {
		orgMW := middleware.ChiGraphQLOrgMiddleware()
		jwtMW := middleware.ChiJWTAuthMiddleware(jwtConfig)
		cacheMW := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				useCache := req.URL.Query().Get("useCache") != "false"
				next.ServeHTTP(w, req.WithContext(ctxutils.SetUseCache(req.Context(), useCache)))
			})
		}
		return requestIDInjectorMiddleware(jwtMW(orgMW(cacheMW(next))))
	}

	runtimePath := "/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
	router.With(runtimeMW).Get(runtimePath, handlers.ModelRuntimeHandler.HandlePlayground)
	router.With(runtimeMW).Post(runtimePath, handlers.ModelRuntimeHandler.HandleQuery)
}

// SetupEndUserRoutesOnChi registers End-User internal HTTP routes.
// All routes keep the same internal-route middleware strategy currently used in this package.
func SetupEndUserRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	internalTokenMW := middleware.ChiInternalTokenMiddleware(cfg.Auth.InternalToken)

	if handlers.EndUserAuthHandler != nil {
		router.Route("/internal/end-user/auth", func(r chi.Router) {
			r.Use(requestIDInjectorMiddleware)
			r.Use(internalTokenMW)
			r.Post("/register", handlers.EndUserAuthHandler.Register)
			r.Post("/login", handlers.EndUserAuthHandler.Login)
			r.Post("/logout", handlers.EndUserAuthHandler.Logout)
			r.Post("/refresh", handlers.EndUserAuthHandler.Refresh)
			r.Get("/me", handlers.EndUserAuthHandler.Me)
		})
	}

	if handlers.EndUserMgmtHandler != nil {
		router.Route("/internal/end-users", func(r chi.Router) {
			r.Use(requestIDInjectorMiddleware)
			r.Use(internalTokenMW)
			r.Post("/", handlers.EndUserMgmtHandler.Create)
			r.Get("/", handlers.EndUserMgmtHandler.List)
			r.Patch("/{userId}/status", handlers.EndUserMgmtHandler.UpdateStatus)
			r.Delete("/{userId}", handlers.EndUserMgmtHandler.Delete)
		})
	}

	if handlers.EndUserDataHandler != nil {
		router.Route("/internal/end-user/data", func(r chi.Router) {
			r.Use(requestIDInjectorMiddleware)
			r.Use(internalTokenMW)
			r.Get("/database-catalog", handlers.EndUserDataHandler.DatabaseCatalog)
			r.Get("/model-catalog", handlers.EndUserDataHandler.ModelCatalog)
		})
	}
}

// SetupEndUserGraphQLRoutesOnChi registers End-User GraphQL endpoint.
// Route pattern: /graphql/end-user/org/{orgName}/project/{projectSlug}
// This endpoint is designed for end-user runtime data access only,
// providing a strict subset of capabilities compared to developer GraphQL.
func SetupEndUserGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	// Create end-user resolver
	endUserResolver := &endusergraphql.Resolver{
		ModelDesignService: handlers.ModelAppService,
	}

	// End-user GraphQL uses internal token auth (BFF-to-backend)
	internalTokenMW := middleware.ChiInternalTokenMiddleware(cfg.Auth.InternalToken)

	// Register end-user GraphQL endpoint
	router.Route("/graphql/end-user/org/{orgName}/project/{projectSlug}", func(r chi.Router) {
		r.Use(requestIDInjectorMiddleware)
		r.Use(internalTokenMW)
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLProjectMiddleware())
		r.Post("/", endusergraphql.EndUserGraphQLHandler(endUserResolver))
		r.Get("/", endusergraphql.EndUserPlaygroundHandler())
	})
}

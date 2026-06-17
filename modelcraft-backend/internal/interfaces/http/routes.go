package http

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"modelcraft/internal/app/auth"
	"modelcraft/internal/app/cluster"
	"modelcraft/internal/app/modeldesign"
	"modelcraft/internal/app/modelruntime"
	"modelcraft/internal/app/project"
	"modelcraft/internal/app/rls"
	"modelcraft/internal/infrastructure/database/ddl"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/internal/middleware"
	"modelcraft/pkg/config"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"os"
	"sync"
	"time"

	appmodeldatabase "modelcraft/internal/app/modeldatabase"

	rlsRepo "modelcraft/internal/infrastructure/persistence/rls"

	orggraphql "modelcraft/internal/interfaces/graphql/org"
	projectgraphql "modelcraft/internal/interfaces/graphql/project"

	runtime "modelcraft/internal/interfaces/runtime"

	appOrg "modelcraft/internal/app/organization"
	appPermission "modelcraft/internal/app/permission"
	appProfile "modelcraft/internal/app/profile"

	appRole "modelcraft/internal/app/role"
	"modelcraft/internal/app/apitoken"
	domainAuth "modelcraft/internal/domain/auth"
	domainEndUser "modelcraft/internal/domain/enduser"
	domainModelDesign "modelcraft/internal/domain/modeldesign"
	domainUser "modelcraft/internal/domain/user"
	infraAuth "modelcraft/internal/infrastructure/auth"

	appAuth "modelcraft/internal/app/auth"
	authHandlers "modelcraft/internal/interfaces/http/handlers/auth"
	userHandlers "modelcraft/internal/interfaces/http/handlers/user"
	httpmiddleware "modelcraft/internal/interfaces/http/middleware"

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

	// Repositories
	ModelRepository domainModelDesign.ModelRepository
	UserRepo        domainUser.UserRepository
	ClusterManager  *repository.ClusterConnectionManager

	// End-User Services
	UserAPITokenService *apitoken.APITokenService

	// Org Creation Service
	CreateOrgService *appOrg.CreateOrganizationService

	// Database management
	ModelDatabaseAppService     *appmodeldatabase.ModelDatabaseAppService
	ModelDatabaseSyncAppService *appmodeldatabase.ModelDatabaseSyncAppService
	SyncModelsAppService        *appmodeldatabase.SyncModelsAppService

	// SystemDB is the system main database connection (stores end_user_users etc.)
	SystemDB *sql.DB

	// IsOrgAdminFn checks whether an end-user has org-admin status (user_orgs.is_admin=true).
	// Used by the runtime PAT middleware to inject the IsAdmin context flag, mirroring
	// the APISIX X-Is-Admin header path for JWT callers.
	IsOrgAdminFn httpmiddleware.IsOrgAdminFn
}

// authEndUserRepoFactory satisfies appAuth.EndUserRepositoryFactory.
// projectSlug is not needed for auth operations.
type authEndUserRepoFactory struct{}

func (a *authEndUserRepoFactory) NewEndUserRepository(
	db appAuth.SQLDBTX,
	orgName string,
) domainEndUser.EndUserRepository {
	return repository.NewSqlEndUserRepository(db, orgName, "")
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

	// Create model database management service
	modelDatabaseRepo := repository.NewSqlModelDatabaseRepository(dbgen.New(loggingDB))
	modelDatabaseAppService := appmodeldatabase.NewModelDatabaseAppService(
		modelDatabaseRepo,
		clusterRepository,
		clusterManager,
	)

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
		modelDatabaseRepo,
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
	modelDatabaseSyncJobRepo := repository.NewSqlModelDatabaseSyncJobRepository(dbgen.New(loggingDB))
	importGroupService := appmodeldatabase.NewImportGroupService(groupRepository, groupAppService)
	modelDatabaseSyncAppService := appmodeldatabase.NewModelDatabaseSyncAppService(
		appmodeldatabase.ModelDatabaseSyncAppServiceDeps{
			ModelDatabaseRepo: modelDatabaseRepo,
			SyncJobRepo:       modelDatabaseSyncJobRepo,
			ReverseEngineer:   reverseEngineerApp,
			ModelRepo:         modelRepository,
			SchemaSync:        appService,
			GroupService:      importGroupService,
		},
	)

	syncModelsSyncJobRepo := repository.NewSqlModelSyncJobRepository(dbgen.New(loggingDB))
	syncModelsAppService := appmodeldatabase.NewSyncModelsAppService(appmodeldatabase.SyncModelsAppServiceDeps{
		SyncJobRepo:     syncModelsSyncJobRepo,
		DBRepo:          modelDatabaseRepo,
		ReverseEngineer: reverseEngineerApp,
		ModelRepo:       modelRepository,
		FieldSyncer:     appService,
		GroupService:    importGroupService,
	})

	// Recover stale model sync jobs on startup (non-blocking)
	go func() {
		ctx := context.Background()
		if err := syncModelsAppService.RecoverStaleJobs(ctx); err != nil {
			log.Printf("failed to recover stale model sync jobs: %v", err)
		}
	}()

	// Create RLS related services (V2: multi-policy matching)
	authSchemaRepo := rlsRepo.NewSqlAuthSchemaRepository(dbgen.New(loggingDB))
	authSchemaAppService := rls.NewAuthSchemaAppService(authSchemaRepo, projectRepository)
	// Create user management related services
	userRepo := repository.NewSqlUserRepository(dbgen.New(loggingDB))
	profileRepo := repository.NewSqlProfileRepository(dbgen.New(loggingDB))
	profileAppService := appProfile.NewAppService(userRepo, profileRepo)
	orgRepo := repository.NewSqlOrganizationRepository(dbgen.New(loggingDB))
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

	createOrgService := appOrg.NewCreateOrganizationService(
		txManager,
		userRepo,
		orgRepo,
		casbinRoleRepo,
	)

	passwordHasher := infraAuth.NewBcryptPasswordHasher()

	// Initialise ES256 JWT signer. Use PEM from config; fall back to ephemeral dev key.
	var jwtSigner *domainAuth.JWTSigner
	if cfg.JWT.PrivateKey != "" {
		var signerErr error
		jwtSigner, signerErr = domainAuth.NewJWTSignerFromPEM(cfg.JWT.PrivateKey, cfg.JWT.Issuer, cfg.JWT.Expiration)
		if signerErr != nil {
			return nil, fmt.Errorf("init jwt signer: %w", signerErr)
		}
	} else {
		var signerErr error
		jwtSigner, signerErr = domainAuth.GenerateDevSigner()
		if signerErr != nil {
			return nil, fmt.Errorf("generate dev jwt signer: %w", signerErr)
		}
	}

	tokenService := auth.NewTokenService(
		refreshTokenRepo,
		userRepo,
		orgRepo,
		profileRepo,
		auditLogRepo,
		passwordHasher,
		7*24*time.Hour, // refresh token TTL
		createOrgService,
		txManager,
		jwtSigner,
	)

	// Create auth handler with token service
	apiTokenRepo := repository.NewSqlAPITokenRepository(repoFactory.SqlDB)
	apiTokenService := apitoken.NewAPITokenService(apiTokenRepo)

	authHandler := authHandlers.NewHandler(tokenService, apiTokenService, cfg.Auth.Cookie, nil)

	// Attach end-user repository support to TokenService (unified auth service).
	tokenService.WithEndUserSupport(&authEndUserRepoFactory{}, repoFactory.SqlDB)

	return &DesignHandlers{
		AuthHandler:                 authHandler,
		UserHandler:                 userHandlers.NewHandler(userRepo, logger),
		ModelAppService:             appService,
		ClusterAppService:           clusterAppService,
		ReverseEngineerAppService:   reverseEngineerApp,
		EnumAppService:              enumAppService,
		ProjectAppService:           projectAppService,
		OrgAppService:               orgAppService,
		ProfileAppService:           profileAppService,
		RoleAppService:              roleAppService,
		PermRoleService:             permRoleService,
		PermPermissionService:       permPermissionService,
		PermUserRoleService:         permUserRoleService,
		ModelRepository:             modelRepository,
		UserRepo:                    userRepo,
		ClusterManager:              clusterManager,
		SystemDB:                    repoFactory.SqlDB,
		GroupAppService:             groupAppService,
		LogicalFKAppService:         logicalFKAppService,
		RLSPolicyAppService:         nil, // TODO: replace with V2 policy CRUD in task 9
		AuthSchemaAppService:        authSchemaAppService,
		UserAPITokenService:      apiTokenService,
		CreateOrgService:            createOrgService,
		ModelDatabaseAppService:     modelDatabaseAppService,
		ModelDatabaseSyncAppService: modelDatabaseSyncAppService,
		SyncModelsAppService:        syncModelsAppService,
		IsOrgAdminFn:                nil,
	}, nil
}

// SetupOrgGraphQLRoutesOnChi registers GraphQL endpoints for org domain.
// Route patterns:
//   - /graphql/org/{orgName}/          → tenant (JWT, internal)
//   - /end-user/graphql/org/{orgName}  → end-user (JWT, gateway-facing)
//
// Both routes use JWT auth — the gateway converts PAT tokens to JWT before
// forwarding, so the backend never sees PAT on design-time GraphQL routes.
func SetupOrgGraphQLRoutesOnChi(
	router chi.Router,
	handlers *DesignHandlers,
	cfg *config.Config,
	jwtConfig *middleware.JWTAuthConfig,
) {
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
		APITokenService:        handlers.UserAPITokenService,
	}

	// ── Tenant route (internal) ──────────────────────────────────────────
	router.Route("/graphql/org/{orgName}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLActionMiddleware())
		r.Post("/", orggraphql.OrgGraphQLHandler(orgResolver))
		r.Get("/", orggraphql.OrgPlaygroundHandler())
	})

	// ── End-user route (gateway-facing, JWT from gateway PAT→JWT conversion)
	router.Route("/end-user/graphql/org/{orgName}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLActionMiddleware())
		r.Post("/", orggraphql.OrgEndUserGraphQLHandler(orgResolver))
		r.Get("/", orggraphql.OrgPlaygroundHandler())
	})
}

// SetupProjectGraphQLRoutesOnChi registers GraphQL endpoints for project domain.
// Route patterns:
//   - /graphql/org/{orgName}/project/{projectSlug}/          → tenant (JWT, internal)
//   - /end-user/graphql/org/{orgName}/project/{projectSlug}  → end-user (JWT, gateway-facing)
//
// Both routes use JWT auth — the gateway converts PAT tokens to JWT before
// forwarding, so the backend never sees PAT on design-time GraphQL routes.
func SetupProjectGraphQLRoutesOnChi(
	router chi.Router,
	handlers *DesignHandlers,
	cfg *config.Config,
	jwtConfig *middleware.JWTAuthConfig,
) {
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

	// Create RLS policy CRUD service
	policyCRUDService := (*rls.PolicyCRUDService)(nil)

	// Create project resolver
	projectResolver := &projectgraphql.Resolver{
		ClusterAppService:           handlers.ClusterAppService,
		ModelDesignService:          handlers.ModelAppService,
		ReverseEngineerService:      handlers.ReverseEngineerAppService,
		RepairModelUseCase:          repairUseCase,
		ActualSchemaQueryUseCase:    actualSchemaQueryUseCase,
		GroupAppService:             handlers.GroupAppService,
		LogicalFKAppService:         handlers.LogicalFKAppService,
		EnumAppService:              handlers.EnumAppService,
		UserRoleService:             handlers.PermUserRoleService,
		FieldSelectionChecker:       projectgraphql.NewFieldSelectionChecker(),
		RLSPolicyAppService:         handlers.RLSPolicyAppService,
		AuthSchemaAppService:        handlers.AuthSchemaAppService,
		ModelDatabaseAppService:     handlers.ModelDatabaseAppService,
		ModelDatabaseSyncAppService: handlers.ModelDatabaseSyncAppService,
		SyncModelsAppService:        handlers.SyncModelsAppService,
		PolicyCRUDService:           policyCRUDService,
	}

	// ── Tenant route (JWT only) ──────────────────────────────────────────
	router.Route("/graphql/org/{orgName}/project/{projectSlug}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLProjectMiddleware())
		r.Use(middleware.ChiGraphQLActionMiddleware())
		r.Post("/", projectgraphql.ProjectGraphQLHandler(projectResolver))
		r.Get("/", projectgraphql.ProjectPlaygroundHandler())
	})

	// ── End-user route (gateway-facing, JWT from gateway PAT→JWT conversion)
	router.Route("/end-user/graphql/org/{orgName}/project/{projectSlug}", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLProjectMiddleware())
		r.Use(middleware.ChiGraphQLActionMiddleware())
		r.Post("/", projectgraphql.ProjectEndUserGraphQLHandler(projectResolver))
		r.Get("/", projectgraphql.ProjectPlaygroundHandler())
	})
}

// RuntimeHandlers holds the handler needed for the model runtime GraphQL API.
type RuntimeHandlers struct {
	ModelRuntimeHandler *runtime.ModelRuntimeHandler
}

// LoadRSAPublicKey loads the RSA public key for design API JWT validation.
// It supports two sources (in priority order):
// 1. PEM file path (AUTH_JWT_PUBLIC_KEY_PATH)
// 2. Direct public key string (AUTH_JWT_PUBLIC_KEY)
// This function is exported for use by both Gin and Chi middleware setup.
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

// CreateRuntimeHandlers initialises repository, application service, and handler
// for the model runtime API.
func CreateRuntimeHandlers(db *sql.DB) *RuntimeHandlers {
	loggedDB := repository.NewSqlcLogger(db, repository.SqlcLogInfo, 200*time.Millisecond)
	loggingDB := dbgen.New(loggedDB)
	modelRuntimeRepo := repository.NewSqlModelRuntimeRepository(loggingDB)
	lfkRepo := repository.NewSqlLogicalForeignKeyRepository(loggingDB)
	graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo, nil, nil)
	handler := runtime.NewModelRuntimeHandler(graphqlAppService)
	return &RuntimeHandlers{
		ModelRuntimeHandler: handler,
	}
}

func SetupRuntimeGraphQLRoutesOnChi(
	router chi.Router,
	handlers *RuntimeHandlers,
	cfg *config.Config,
) {
	jwtConfig := &middleware.JWTAuthConfig{
		SkipValidation: !cfg.Auth.Runtime.Enabled,
	}

	orgMW := middleware.ChiGraphQLOrgMiddleware()
	jwtMW := middleware.ChiJWTAuthMiddleware(jwtConfig)
	cacheMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			useCache := req.URL.Query().Get("useCache") != "false"
			next.ServeHTTP(w, req.WithContext(ctxutils.SetUseCache(req.Context(), useCache)))
		})
	}

	// NOTE: GraphQL responses already carry requestId in extensions, so we skip
	// the JSON-body injector here to avoid duplicating requestId at the top level.
	runtimeMW := func(next http.Handler) http.Handler {
		return jwtMW(orgMW(cacheMW(next)))
	}

	// End-user runtime: JWT auth (PAT→JWT conversion is handled by APISIX gateway).
	// Also accepts X-MC-Auth-* headers for RLS context injection.
	endUserRuntimeMW := func(next http.Handler) http.Handler {
		return httpmiddleware.NewRLSContextMiddleware().Middleware(jwtMW(orgMW(cacheMW(next))))
	}

	runtimePath := "/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
	router.With(runtimeMW).Get(runtimePath, handlers.ModelRuntimeHandler.HandlePlayground)
	router.With(runtimeMW).Post(runtimePath, handlers.ModelRuntimeHandler.HandleQuery)

	endUserRuntimePath := "/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
	router.With(endUserRuntimeMW).Get(endUserRuntimePath, handlers.ModelRuntimeHandler.HandlePlayground)
	router.With(endUserRuntimeMW).Post(endUserRuntimePath, handlers.ModelRuntimeHandler.HandleQuery)
}

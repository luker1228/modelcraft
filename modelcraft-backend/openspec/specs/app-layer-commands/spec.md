# app-layer-commands Specification

## Purpose
TBD - created by archiving change refactor-request-types-to-app-layer. Update Purpose after archive.
## Requirements
### Requirement: App Layer Command Types
Each App Service method that mutates state SHALL accept a single Command struct defined in the same `internal/app/{context}/` package, instead of depending on interface-layer Request types or accepting multiple individual parameters.

#### Scenario: Cluster creation uses Command type
- **WHEN** a GraphQL resolver or HTTP handler needs to create a cluster
- **THEN** it SHALL construct a `cluster.CreateClusterCommand` struct and pass it to `DatabaseClusterAppService.CreateCluster(ctx, cmd)`
- **AND** the Command struct SHALL be defined in `internal/app/cluster/commands.go`
- **AND** the Command struct SHALL NOT contain HTTP binding tags (`json`, `binding`)

#### Scenario: Enum creation uses Command type instead of multiple parameters
- **WHEN** a GraphQL resolver needs to create an enum
- **THEN** it SHALL construct a `CreateEnumCommand` struct and pass it to `EnumAppService.CreateEnum(ctx, cmd)`
- **AND** the App Service method SHALL NOT accept more than 2 positional parameters (ctx + command)

#### Scenario: Project creation uses Command type instead of multiple parameters
- **WHEN** a GraphQL resolver needs to create a project
- **THEN** it SHALL construct a `CreateProjectCommand` struct and pass it to `ProjectAppService.CreateProject(ctx, cmd)`

#### Scenario: Model creation uses Command type
- **WHEN** a GraphQL resolver needs to create a model
- **THEN** it SHALL construct a `CreateModelCommand` struct and pass it to the model design App Service

### Requirement: App Layer Dependency Direction
The App layer (`internal/app/`) SHALL NOT import any package from the Interface layer (`internal/interfaces/`). All data flowing from Interface to App SHALL be through App-layer-defined Command and Query types.

#### Scenario: No interface layer imports in app layer
- **WHEN** inspecting any file under `internal/app/`
- **THEN** there SHALL be zero import statements referencing `internal/interfaces/`

#### Scenario: Interface layer converts to app layer types
- **WHEN** a GraphQL resolver receives a `generated.CreateDatabaseClusterInput`
- **THEN** the resolver or its adapter SHALL convert it to `cluster.CreateClusterCommand` before calling the App Service
- **AND** the `generated.*Input` type SHALL NOT be passed to the App Service

### Requirement: Consistent Mutation Method Signatures
All App Service mutation methods SHALL follow the pattern: `Method(ctx context.Context, cmd XxxCommand) (result, error)`. This ensures consistency across all domains.

#### Scenario: All domains follow the same pattern
- **WHEN** reviewing App Service methods for Create, Update, or Delete operations across Cluster, Model, Enum, and Project domains
- **THEN** each mutation method SHALL accept exactly two parameters: `context.Context` and a Command struct
- **AND** the Command struct name SHALL follow the pattern `{Verb}{Resource}Command` (e.g., `CreateClusterCommand`, `UpdateModelCommand`)


export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Date: { input: any; output: any; }
  Int64: { input: any; output: any; }
  Time: { input: any; output: any; }
};

export type ActualConstraintType =
  | 'NOT_NULL'
  | 'UNIQUE';

export type ActualForeignKey = {
  __typename?: 'ActualForeignKey';
  constraintName: Scalars['String']['output'];
  referencedColumn: Scalars['String']['output'];
  referencedTable: Scalars['String']['output'];
};

export type AddFieldInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  enumConfig?: InputMaybe<EnumConfigInput>;
  enumLabelConfig?: InputMaybe<EnumLabelConfigInput>;
  format: FormatType;
  isArray?: InputMaybe<Scalars['Boolean']['input']>;
  isUnique?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  nonNull?: InputMaybe<Scalars['Boolean']['input']>;
  relateFkId?: InputMaybe<Scalars['String']['input']>;
  required?: InputMaybe<Scalars['Boolean']['input']>;
  storageHint?: InputMaybe<Scalars['String']['input']>;
  title: Scalars['String']['input'];
  validationConfig?: InputMaybe<ValidationConfigInput>;
};

export type AddRolePermissionPayload = {
  __typename?: 'AddRolePermissionPayload';
  error?: Maybe<RolePermissionError>;
  success: Scalars['Boolean']['output'];
};

export type ApiKey = {
  __typename?: 'ApiKey';
  createdAt: Scalars['Time']['output'];
  expiresAt?: Maybe<Scalars['Time']['output']>;
  id: Scalars['ID']['output'];
  keyPrefix: Scalars['String']['output'];
  lastUsedAt?: Maybe<Scalars['Time']['output']>;
  name: Scalars['String']['output'];
  revokedAt?: Maybe<Scalars['Time']['output']>;
};

export type ApiKeyInvalidInput = Error & {
  __typename?: 'ApiKeyInvalidInput';
  message: Scalars['String']['output'];
};

export type ApiKeyLimitExceeded = Error & {
  __typename?: 'ApiKeyLimitExceeded';
  message: Scalars['String']['output'];
};

export type ApiKeyNotFound = Error & {
  __typename?: 'ApiKeyNotFound';
  message: Scalars['String']['output'];
};

export type AssignRoleError = PermissionInvalidInput | PermissionRoleNotFound | PermissionUserNotFound;

export type AssignRolePayload = {
  __typename?: 'AssignRolePayload';
  error?: Maybe<AssignRoleError>;
  userRole?: Maybe<UserRoleAssignment>;
};

export type CannotDeleteDefaultProject = Error & {
  __typename?: 'CannotDeleteDefaultProject';
  message: Scalars['String']['output'];
};

export type CannotDeleteDeployedModel = Error & {
  __typename?: 'CannotDeleteDeployedModel';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type CannotDeleteReferencedEnum = Error & {
  __typename?: 'CannotDeleteReferencedEnum';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type CannotDeleteSystemRole = Error & {
  __typename?: 'CannotDeleteSystemRole';
  message: Scalars['String']['output'];
};

export type ClusterAlreadyExists = Error & {
  __typename?: 'ClusterAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ClusterAlreadyExistsForProject = Error & {
  __typename?: 'ClusterAlreadyExistsForProject';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ClusterConnectionInput = {
  connectionInfo: DatabaseConnectionInput;
  description?: InputMaybe<Scalars['String']['input']>;
  title: Scalars['String']['input'];
};

export type ClusterNotFound = Error & {
  __typename?: 'ClusterNotFound';
  message: Scalars['String']['output'];
};

export type ClusterStatus =
  | 'ACTIVE'
  | 'DISABLED';

export type CreateApiKeyError = ApiKeyInvalidInput | ApiKeyLimitExceeded;

export type CreateApiKeyInput = {
  expiresAt?: InputMaybe<Scalars['Time']['input']>;
  name: Scalars['String']['input'];
};

export type CreateApiKeyPayload = {
  __typename?: 'CreateApiKeyPayload';
  error?: Maybe<CreateApiKeyError>;
  result?: Maybe<CreateApiKeyResult>;
};

export type CreateApiKeyResult = {
  __typename?: 'CreateApiKeyResult';
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  key: Scalars['String']['output'];
  keyPrefix: Scalars['String']['output'];
  name: Scalars['String']['output'];
};

export type CreateCustomRoleError = PermissionInvalidInput | PermissionRoleAlreadyExists;

export type CreateCustomRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  orgName: Scalars['String']['input'];
};

export type CreateCustomRolePayload = {
  __typename?: 'CreateCustomRolePayload';
  error?: Maybe<CreateCustomRoleError>;
  role?: Maybe<PermissionRole>;
};

export type CreateEnumError = EnumAlreadyExists | InvalidEnumInput | ProjectNotFound;

export type CreateEnumInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  name: Scalars['String']['input'];
  options: Array<EnumOptionInput>;
};

export type CreateEnumPayload = {
  __typename?: 'CreateEnumPayload';
  enum?: Maybe<EnumDefinition>;
  error?: Maybe<CreateEnumError>;
};

export type CreateGroupError = GroupAlreadyExists | InvalidGroupName;

export type CreateGroupInput = {
  name: Scalars['String']['input'];
};

export type CreateGroupPayload = {
  __typename?: 'CreateGroupPayload';
  error?: Maybe<CreateGroupError>;
  group?: Maybe<ModelGroup>;
};

export type CreateLogicalForeignKeyInput = {
  modelId: Scalars['ID']['input'];
  refModelId: Scalars['ID']['input'];
  sourceFields: Array<Scalars['String']['input']>;
  targetFields: Array<Scalars['String']['input']>;
};

export type CreateLogicalForeignKeyPayload = {
  __typename?: 'CreateLogicalForeignKeyPayload';
  result: CreateLogicalForeignKeyResult;
};

export type CreateLogicalForeignKeyResult = FkColumnsNotFoundError | FkFieldCountMismatchError | LogicalForeignKey;

export type CreateModelError = InvalidModelInput | ModelAlreadyExists | ModelTableAlreadyExists | ProjectNotFound;

export type CreateModelFromSchemaInput = {
  databaseName: Scalars['String']['input'];
  schema: Scalars['String']['input'];
};

export type CreateModelFromSchemaPayload = {
  __typename?: 'CreateModelFromSchemaPayload';
  model?: Maybe<Model>;
};

export type CreateModelInput = {
  databaseName: Scalars['String']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type CreateModelPayload = {
  __typename?: 'CreateModelPayload';
  error?: Maybe<CreateModelError>;
  model?: Maybe<Model>;
};

export type CreateProjectError = DatabaseConnectionFailed | InvalidProjectInput | ProjectAlreadyExists;

export type CreateProjectInput = {
  clusterInput: ClusterConnectionInput;
  description?: InputMaybe<Scalars['String']['input']>;
  loginUrl?: InputMaybe<Scalars['String']['input']>;
  skipConnectionTest?: InputMaybe<Scalars['Boolean']['input']>;
  slug: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type CreateProjectPayload = {
  __typename?: 'CreateProjectPayload';
  error?: Maybe<CreateProjectError>;
  project?: Maybe<Project>;
};

export type CreateRoleError = InvalidRoleInput | RoleAlreadyExists;

export type CreateRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  permissions: Array<Scalars['String']['input']>;
};

export type CreateRolePayload = {
  __typename?: 'CreateRolePayload';
  error?: Maybe<CreateRoleError>;
  role?: Maybe<Role>;
};

export type CurrentUser = {
  __typename?: 'CurrentUser';
  email: Scalars['String']['output'];
  externalID: Scalars['String']['output'];
  id: Scalars['String']['output'];
  name: Scalars['String']['output'];
  organization?: Maybe<Organization>;
  permissions: Array<Scalars['String']['output']>;
  role?: Maybe<Role>;
};

export type Database = {
  __typename?: 'Database';
  name: Scalars['String']['output'];
};

export type DatabaseCluster = Node & {
  __typename?: 'DatabaseCluster';
  connectionInfo: DatabaseConnectionInfo;
  connectionTimeout: Scalars['Int']['output'];
  createdAt: Scalars['String']['output'];
  deletedAt?: Maybe<Scalars['String']['output']>;
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  projectSlug: Scalars['String']['output'];
  status: ClusterStatus;
  title: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
  version: Scalars['Int']['output'];
};

export type DatabaseConnection = {
  __typename?: 'DatabaseConnection';
  edges: Array<DatabaseEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type DatabaseConnectionFailed = Error & {
  __typename?: 'DatabaseConnectionFailed';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type DatabaseConnectionInfo = {
  __typename?: 'DatabaseConnectionInfo';
  host: Scalars['String']['output'];
  password: Scalars['String']['output'];
  port: Scalars['Int64']['output'];
  username: Scalars['String']['output'];
};

export type DatabaseConnectionInput = {
  connectionTimeout?: InputMaybe<Scalars['Int']['input']>;
  host: Scalars['String']['input'];
  password: Scalars['String']['input'];
  port: Scalars['Int']['input'];
  username: Scalars['String']['input'];
};

export type DatabaseEdge = {
  __typename?: 'DatabaseEdge';
  cursor: Scalars['String']['output'];
  node: Database;
};

export type DbColumnInfo = {
  __typename?: 'DbColumnInfo';
  columnLength?: Maybe<Scalars['Int64']['output']>;
  columnType: Scalars['String']['output'];
  conflicts: Array<FieldConflict>;
  constraints: Array<ActualConstraintType>;
  defaultValue?: Maybe<Scalars['String']['output']>;
  foreignKey?: Maybe<ActualForeignKey>;
  isPrimaryKey: Scalars['Boolean']['output'];
  nonNull: Scalars['Boolean']['output'];
  unique: Scalars['Boolean']['output'];
};

export type DbTableStatus =
  | 'CLUSTER_UNREACHABLE'
  | 'TABLE_EXISTS'
  | 'TABLE_MISSING';

export type DeleteClusterError = ClusterNotFound | ProjectNotFound;

export type DeleteClusterPayload = {
  __typename?: 'DeleteClusterPayload';
  error?: Maybe<DeleteClusterError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEnumError = CannotDeleteReferencedEnum | EnumNotFound | ProjectNotFound;

export type DeleteEnumPayload = {
  __typename?: 'DeleteEnumPayload';
  error?: Maybe<DeleteEnumError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteGroupError = GroupNotFound;

export type DeleteGroupPayload = {
  __typename?: 'DeleteGroupPayload';
  error?: Maybe<DeleteGroupError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteLogicalForeignKeyPayload = {
  __typename?: 'DeleteLogicalForeignKeyPayload';
  result: DeleteLogicalForeignKeyResult;
};

export type DeleteLogicalForeignKeyResult = DeleteLogicalForeignKeySuccess | FkNotFoundError | FkPairHasRelateFieldsError;

export type DeleteLogicalForeignKeySuccess = {
  __typename?: 'DeleteLogicalForeignKeySuccess';
  pairId: Scalars['String']['output'];
};

export type DeleteModelError = CannotDeleteDeployedModel | ModelNotFound | ProjectNotFound;

export type DeleteModelPayload = {
  __typename?: 'DeleteModelPayload';
  error?: Maybe<DeleteModelError>;
  success: Scalars['Boolean']['output'];
};

export type DeletePermissionRoleError = PermissionRoleNotFound | PermissionSystemRoleCannotBeModified;

export type DeletePermissionRolePayload = {
  __typename?: 'DeletePermissionRolePayload';
  error?: Maybe<DeletePermissionRoleError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteProjectError = CannotDeleteDefaultProject | ProjectNotFound;

export type DeleteProjectPayload = {
  __typename?: 'DeleteProjectPayload';
  error?: Maybe<DeleteProjectError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteRoleError = CannotDeleteSystemRole | RoleNotFound;

export type DeleteRolePayload = {
  __typename?: 'DeleteRolePayload';
  error?: Maybe<DeleteRoleError>;
  success: Scalars['Boolean']['output'];
};

export type EnumAlreadyExists = Error & {
  __typename?: 'EnumAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EnumConfigInput = {
  connectEnum: Scalars['Boolean']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  enumName: Scalars['String']['input'];
  options?: InputMaybe<Array<EnumOptionInput>>;
};

export type EnumDefinition = {
  __typename?: 'EnumDefinition';
  createdAt: Scalars['String']['output'];
  description?: Maybe<Scalars['String']['output']>;
  displayName: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  isMultiSelect: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  options: Array<EnumOption>;
  projectSlug: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
};

export type EnumLabelConfigInput = {
  sourceField: Scalars['String']['input'];
};

export type EnumNotFound = Error & {
  __typename?: 'EnumNotFound';
  message: Scalars['String']['output'];
};

export type EnumOption = {
  __typename?: 'EnumOption';
  code: Scalars['String']['output'];
  description?: Maybe<Scalars['String']['output']>;
  label: Scalars['String']['output'];
  order: Scalars['Int']['output'];
};

export type EnumOptionInput = {
  code: Scalars['String']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  label: Scalars['String']['input'];
  order: Scalars['Int']['input'];
};

export type Error = {
  message: Scalars['String']['output'];
};

export type FkColumnsNotFoundError = {
  __typename?: 'FKColumnsNotFoundError';
  message: Scalars['String']['output'];
};

export type FkDirection =
  | 'NORMAL'
  | 'REVERSE';

export type FkFieldCountMismatchError = {
  __typename?: 'FKFieldCountMismatchError';
  message: Scalars['String']['output'];
};

export type FkNotFoundError = {
  __typename?: 'FKNotFoundError';
  message: Scalars['String']['output'];
};

export type FkPairHasRelateFieldsError = {
  __typename?: 'FKPairHasRelateFieldsError';
  message: Scalars['String']['output'];
};

export type Field = {
  __typename?: 'Field';
  belongsToFkId?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['String']['output'];
  dbColumn?: Maybe<DbColumnInfo>;
  description?: Maybe<Scalars['String']['output']>;
  enum?: Maybe<EnumDefinition>;
  format: FormatType;
  isArray: Scalars['Boolean']['output'];
  isDeprecated: Scalars['Boolean']['output'];
  isPrimary: Scalars['Boolean']['output'];
  isUnique: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  nonNull: Scalars['Boolean']['output'];
  relateFkId?: Maybe<Scalars['String']['output']>;
  required: Scalars['Boolean']['output'];
  schemaType: SchemaType;
  storageHint: Scalars['String']['output'];
  title: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
  validationConfig?: Maybe<ValidationConfig>;
};

export type FieldConflict = {
  __typename?: 'FieldConflict';
  actual: Scalars['String']['output'];
  aspect: FieldConflictAspect;
  expected: Scalars['String']['output'];
};

export type FieldConflictAspect =
  | 'NOT_NULL_MISMATCH'
  | 'PRIMARY_MISMATCH'
  | 'UNIQUE_MISMATCH';

export type FormatType =
  | 'BOOLEAN'
  | 'DATE'
  | 'DATETIME'
  | 'DECIMAL'
  | 'ENUM'
  | 'ENUM_LABEL'
  | 'INTEGER'
  | 'NUMBER'
  | 'RELATION'
  | 'STRING'
  | 'TIME'
  | 'UUID';

export type GetClusterError = ClusterNotFound | ProjectNotFound;

export type GetClusterPayload = {
  __typename?: 'GetClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<GetClusterError>;
};

export type GetEnumError = EnumNotFound | ProjectNotFound;

export type GetEnumPayload = {
  __typename?: 'GetEnumPayload';
  enum?: Maybe<EnumDefinition>;
  error?: Maybe<GetEnumError>;
};

export type GetModelError = InvalidModelInput | ModelNotFound | ProjectNotFound;

export type GetModelPayload = {
  __typename?: 'GetModelPayload';
  error?: Maybe<GetModelError>;
  model?: Maybe<Model>;
};

export type GetOrganizationError = OrganizationNotFound;

export type GetOrganizationPayload = {
  __typename?: 'GetOrganizationPayload';
  error?: Maybe<GetOrganizationError>;
  organization?: Maybe<Organization>;
};

export type GetProjectError = ProjectNotFound;

export type GetProjectPayload = {
  __typename?: 'GetProjectPayload';
  error?: Maybe<GetProjectError>;
  project?: Maybe<Project>;
};

export type GroupAlreadyExists = Error & {
  __typename?: 'GroupAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type GroupNotFound = Error & {
  __typename?: 'GroupNotFound';
  message: Scalars['String']['output'];
};

export type HealthStatus =
  | 'BROKEN'
  | 'HEALTHY'
  | 'NEEDS_REPAIR';

export type ImportModelInput = {
  databaseName: Scalars['String']['input'];
  tableName: Scalars['String']['input'];
};

export type ImportModelPayload = {
  __typename?: 'ImportModelPayload';
  fieldsCount: Scalars['Int64']['output'];
  modelId: Scalars['String']['output'];
  modelName: Scalars['String']['output'];
  skippedFields: Array<Scalars['String']['output']>;
};

export type InvalidClusterInput = Error & {
  __typename?: 'InvalidClusterInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidEnumInput = Error & {
  __typename?: 'InvalidEnumInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidGroupName = Error & {
  __typename?: 'InvalidGroupName';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidModelInput = Error & {
  __typename?: 'InvalidModelInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidProjectInput = Error & {
  __typename?: 'InvalidProjectInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidRoleInput = Error & {
  __typename?: 'InvalidRoleInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ListDatabasesInput = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListProjectsInput = {
  status?: InputMaybe<ProjectStatus>;
};

export type ListTablesInput = {
  databaseName: Scalars['String']['input'];
  excludeExisting?: InputMaybe<Scalars['Boolean']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
};

export type LogicalForeignKey = {
  __typename?: 'LogicalForeignKey';
  direction: FkDirection;
  id: Scalars['ID']['output'];
  modelId: Scalars['String']['output'];
  modelName: Scalars['String']['output'];
  pairId: Scalars['String']['output'];
  refModelId: Scalars['String']['output'];
  refModelName: Scalars['String']['output'];
  sourceFields: Array<Scalars['String']['output']>;
  targetFields: Array<Scalars['String']['output']>;
};

export type MembershipStatus =
  | 'ACTIVE'
  | 'INVITED'
  | 'SUSPENDED';

export type Model = Node & {
  __typename?: 'Model';
  createdAt: Scalars['String']['output'];
  databaseName: Scalars['String']['output'];
  dbTable?: Maybe<DbTableStatus>;
  description: Scalars['String']['output'];
  fields: Array<Field>;
  group: ModelGroup;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  projectSlug: Scalars['String']['output'];
  storageType: Scalars['String']['output'];
  title: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
};

export type ModelAlreadyExists = Error & {
  __typename?: 'ModelAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ModelConnection = {
  __typename?: 'ModelConnection';
  edges: Array<ModelEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ModelEdge = {
  __typename?: 'ModelEdge';
  cursor: Scalars['String']['output'];
  node: Model;
};

export type ModelGroup = {
  __typename?: 'ModelGroup';
  displayOrder: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  isVirtual: Scalars['Boolean']['output'];
  models: Array<Model>;
  name: Scalars['String']['output'];
};

export type ModelJsonSchema = {
  __typename?: 'ModelJsonSchema';
  modelId: Scalars['ID']['output'];
  modelName: Scalars['String']['output'];
  schema: Scalars['String']['output'];
};

export type ModelNotFound = Error & {
  __typename?: 'ModelNotFound';
  message: Scalars['String']['output'];
};

export type ModelQueryInput = {
  databaseName: Scalars['String']['input'];
  limit?: InputMaybe<Scalars['Int']['input']>;
  offset?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ModelTableAlreadyExists = Error & {
  __typename?: 'ModelTableAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type MoveModelToGroupError = GroupNotFound | ModelNotFound;

export type MoveModelToGroupInput = {
  groupId?: InputMaybe<Scalars['ID']['input']>;
  modelId: Scalars['ID']['input'];
};

export type MoveModelToGroupPayload = {
  __typename?: 'MoveModelToGroupPayload';
  error?: Maybe<MoveModelToGroupError>;
  success: Scalars['Boolean']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  addFields?: Maybe<Model>;
  addPermissionToRole: AddRolePermissionPayload;
  assignRoleToUser: AssignRolePayload;
  createApiKey: CreateApiKeyPayload;
  createCustomRole: CreateCustomRolePayload;
  createEnum: CreateEnumPayload;
  createGroup: CreateGroupPayload;
  createLogicalForeignKey: CreateLogicalForeignKeyPayload;
  createModel: CreateModelPayload;
  createModelFromSchema: CreateModelFromSchemaPayload;
  createProject: CreateProjectPayload;
  createRole: CreateRolePayload;
  deleteEnum: DeleteEnumPayload;
  deleteGroup: DeleteGroupPayload;
  deleteLogicalForeignKey: DeleteLogicalForeignKeyPayload;
  deleteModel: DeleteModelPayload;
  deletePermissionRole: DeletePermissionRolePayload;
  deleteProject: DeleteProjectPayload;
  deleteRole: DeleteRolePayload;
  /**
   * 将字段标记为废弃。
   * 废弃的字段仍然存在于数据库中，但标识为不推荐使用。
   * 若字段已处于废弃状态，幂等返回成功。
   * 状态转换：ACTIVE → DEPRECATED
   */
  deprecateField?: Maybe<Model>;
  importModel: ImportModelPayload;
  moveModelToGroup: MoveModelToGroupPayload;
  pong: Scalars['String']['output'];
  removeField?: Maybe<Model>;
  removePermissionFromRole: RemoveRolePermissionPayload;
  renameGroup: RenameGroupPayload;
  reorderGroup: ReorderGroupPayload;
  repairModel: RepairModelPayload;
  revokeApiKey: RevokeApiKeyPayload;
  revokeRoleFromUser: RevokeRolePayload;
  syncModelSchema: SyncModelSchemaPayload;
  testDatabaseConnection: TestConnectionPayload;
  /**
   * 解除字段的废弃状态，恢复为正常可用。
   * 若字段未处于废弃状态，幂等返回成功。
   * 状态转换：DEPRECATED → ACTIVE
   */
  undeprecateField?: Maybe<Model>;
  updateApiKey: UpdateApiKeyPayload;
  updateEnum: UpdateEnumPayload;
  updateField?: Maybe<Model>;
  updateModelMeta: UpdateModelMetaPayload;
  updateOrganization: UpdateOrganizationPayload;
  updatePermissionRole: UpdatePermissionRolePayload;
  updateProject: UpdateProjectPayload;
  updateProjectCluster: UpdateClusterPayload;
};


export type MutationAddFieldsArgs = {
  input: Array<AddFieldInput>;
  modelID: Scalars['ID']['input'];
};


export type MutationAddPermissionToRoleArgs = {
  act: Scalars['String']['input'];
  obj: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
};


export type MutationAssignRoleToUserArgs = {
  orgName: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
  userId: Scalars['String']['input'];
};


export type MutationCreateApiKeyArgs = {
  input: CreateApiKeyInput;
};


export type MutationCreateCustomRoleArgs = {
  input: CreateCustomRoleInput;
};


export type MutationCreateEnumArgs = {
  input: CreateEnumInput;
};


export type MutationCreateGroupArgs = {
  input: CreateGroupInput;
};


export type MutationCreateLogicalForeignKeyArgs = {
  input: CreateLogicalForeignKeyInput;
};


export type MutationCreateModelArgs = {
  input: CreateModelInput;
};


export type MutationCreateModelFromSchemaArgs = {
  input: CreateModelFromSchemaInput;
};


export type MutationCreateProjectArgs = {
  input: CreateProjectInput;
};


export type MutationCreateRoleArgs = {
  input: CreateRoleInput;
};


export type MutationDeleteEnumArgs = {
  name: Scalars['String']['input'];
};


export type MutationDeleteGroupArgs = {
  groupId: Scalars['ID']['input'];
};


export type MutationDeleteLogicalForeignKeyArgs = {
  pairId: Scalars['String']['input'];
};


export type MutationDeleteModelArgs = {
  dropTable?: InputMaybe<Scalars['Boolean']['input']>;
  id: Scalars['ID']['input'];
};


export type MutationDeletePermissionRoleArgs = {
  roleId: Scalars['Int']['input'];
};


export type MutationDeleteProjectArgs = {
  slug: Scalars['String']['input'];
};


export type MutationDeleteRoleArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeprecateFieldArgs = {
  fieldName: Scalars['String']['input'];
  modelID: Scalars['ID']['input'];
};


export type MutationImportModelArgs = {
  input: ImportModelInput;
};


export type MutationMoveModelToGroupArgs = {
  input: MoveModelToGroupInput;
};


export type MutationRemoveFieldArgs = {
  fieldName: Scalars['String']['input'];
  modelID: Scalars['ID']['input'];
};


export type MutationRemovePermissionFromRoleArgs = {
  act: Scalars['String']['input'];
  obj: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
};


export type MutationRenameGroupArgs = {
  input: RenameGroupInput;
};


export type MutationReorderGroupArgs = {
  input: ReorderGroupInput;
};


export type MutationRepairModelArgs = {
  input: RepairModelInput;
};


export type MutationRevokeApiKeyArgs = {
  id: Scalars['ID']['input'];
};


export type MutationRevokeRoleFromUserArgs = {
  orgName: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
  userId: Scalars['String']['input'];
};


export type MutationSyncModelSchemaArgs = {
  input: SyncModelSchemaInput;
};


export type MutationTestDatabaseConnectionArgs = {
  input: TestDatabaseConnectionInput;
};


export type MutationUndeprecateFieldArgs = {
  fieldName: Scalars['String']['input'];
  modelID: Scalars['ID']['input'];
};


export type MutationUpdateApiKeyArgs = {
  id: Scalars['ID']['input'];
  input: UpdateApiKeyInput;
};


export type MutationUpdateEnumArgs = {
  input: UpdateEnumInput;
  name: Scalars['String']['input'];
};


export type MutationUpdateFieldArgs = {
  fieldName: Scalars['String']['input'];
  input: UpdateFieldInput;
  modelID: Scalars['ID']['input'];
};


export type MutationUpdateModelMetaArgs = {
  id: Scalars['ID']['input'];
  input: UpdateModelMetaInput;
};


export type MutationUpdateOrganizationArgs = {
  input: UpdateOrganizationInput;
};


export type MutationUpdatePermissionRoleArgs = {
  input: UpdateRoleInput;
  roleId: Scalars['Int']['input'];
};


export type MutationUpdateProjectArgs = {
  input: UpdateProjectInput;
};


export type MutationUpdateProjectClusterArgs = {
  input: UpdateClusterConnectionInput;
  projectSlug: Scalars['String']['input'];
};

export type Node = {
  id: Scalars['ID']['output'];
};

export type Organization = {
  __typename?: 'Organization';
  createdAt: Scalars['String']['output'];
  displayName?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  ownerID: Scalars['String']['output'];
  status: OrganizationStatus;
  updatedAt: Scalars['String']['output'];
};

export type OrganizationMember = {
  __typename?: 'OrganizationMember';
  createdAt: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  joinedAt?: Maybe<Scalars['String']['output']>;
  orgID: Scalars['String']['output'];
  role: Role;
  status: MembershipStatus;
  userID: Scalars['String']['output'];
  userName: Scalars['String']['output'];
};

export type OrganizationNotFound = Error & {
  __typename?: 'OrganizationNotFound';
  message: Scalars['String']['output'];
};

export type OrganizationStatus =
  | 'ACTIVE'
  | 'DELETED'
  | 'SUSPENDED';

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type PermissionDef = {
  __typename?: 'PermissionDef';
  act: Scalars['String']['output'];
  obj: Scalars['String']['output'];
};

export type PermissionInvalidInput = PermissionManagementError & {
  __typename?: 'PermissionInvalidInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type PermissionManagementError = {
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type PermissionRole = {
  __typename?: 'PermissionRole';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['Int']['output'];
  isSystem: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  orgName: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type PermissionRoleAlreadyExists = PermissionManagementError & {
  __typename?: 'PermissionRoleAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type PermissionRoleNotFound = PermissionManagementError & {
  __typename?: 'PermissionRoleNotFound';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type PermissionSystemRoleCannotBeModified = PermissionManagementError & {
  __typename?: 'PermissionSystemRoleCannotBeModified';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type PermissionUserNotFound = PermissionManagementError & {
  __typename?: 'PermissionUserNotFound';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type Project = Node & {
  __typename?: 'Project';
  createdAt: Scalars['String']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  loginUrl?: Maybe<Scalars['String']['output']>;
  orgName: Scalars['String']['output'];
  slug: Scalars['String']['output'];
  status: ProjectStatus;
  title: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
};

export type ProjectAlreadyExists = Error & {
  __typename?: 'ProjectAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ProjectConnection = {
  __typename?: 'ProjectConnection';
  edges: Array<ProjectEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ProjectEdge = {
  __typename?: 'ProjectEdge';
  cursor: Scalars['String']['output'];
  node: Project;
};

export type ProjectNotFound = Error & {
  __typename?: 'ProjectNotFound';
  message: Scalars['String']['output'];
};

export type ProjectStatus =
  | 'ACTIVE'
  | 'ARCHIVED';

export type Query = {
  __typename?: 'Query';
  apiKeys: Array<ApiKey>;
  databaseCluster: GetClusterPayload;
  enum: GetEnumPayload;
  enumReferences: Array<Scalars['String']['output']>;
  enums: Array<EnumDefinition>;
  fields: Array<Field>;
  hello: Scalars['String']['output'];
  listDatabases: DatabaseConnection;
  listTables: TableListConnection;
  logicalForeignKeys: Array<LogicalForeignKey>;
  me: CurrentUser;
  model: GetModelPayload;
  modelByName: GetModelPayload;
  modelGroups: Array<ModelGroup>;
  modelJsonSchema?: Maybe<ModelJsonSchema>;
  models: ModelConnection;
  myOrganizations: Array<Organization>;
  node?: Maybe<Node>;
  organizationMembers: Array<OrganizationMember>;
  permissionRole?: Maybe<PermissionRole>;
  permissionRoles: Array<PermissionRole>;
  ping: Scalars['String']['output'];
  project: GetProjectPayload;
  projects: Array<Project>;
  rolePermissionsList: Array<PermissionDef>;
  roles: Array<Role>;
  userRoleAssignments: Array<UserRoleAssignment>;
};


export type QueryDatabaseClusterArgs = {
  projectSlug: Scalars['String']['input'];
};


export type QueryEnumArgs = {
  name: Scalars['String']['input'];
};


export type QueryEnumReferencesArgs = {
  name: Scalars['String']['input'];
};


export type QueryFieldsArgs = {
  modelID: Scalars['ID']['input'];
};


export type QueryListDatabasesArgs = {
  input: ListDatabasesInput;
};


export type QueryListTablesArgs = {
  input: ListTablesInput;
};


export type QueryLogicalForeignKeysArgs = {
  modelId: Scalars['ID']['input'];
};


export type QueryModelArgs = {
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryModelByNameArgs = {
  databaseName: Scalars['String']['input'];
  name: Scalars['String']['input'];
};


export type QueryModelJsonSchemaArgs = {
  id: Scalars['ID']['input'];
};


export type QueryModelsArgs = {
  input?: InputMaybe<ModelQueryInput>;
};


export type QueryNodeArgs = {
  id: Scalars['ID']['input'];
};


export type QueryPermissionRoleArgs = {
  id: Scalars['Int']['input'];
};


export type QueryPermissionRolesArgs = {
  includeSystem?: InputMaybe<Scalars['Boolean']['input']>;
  orgName: Scalars['String']['input'];
};


export type QueryProjectArgs = {
  slug: Scalars['String']['input'];
};


export type QueryProjectsArgs = {
  input?: InputMaybe<ListProjectsInput>;
};


export type QueryRolePermissionsListArgs = {
  roleId: Scalars['Int']['input'];
};


export type QueryUserRoleAssignmentsArgs = {
  orgName: Scalars['String']['input'];
  userId: Scalars['String']['input'];
};

export type RemoveRolePermissionPayload = {
  __typename?: 'RemoveRolePermissionPayload';
  error?: Maybe<RolePermissionError>;
  success: Scalars['Boolean']['output'];
};

export type RenameGroupError = GroupAlreadyExists | GroupNotFound | InvalidGroupName;

export type RenameGroupInput = {
  groupId: Scalars['ID']['input'];
  newName: Scalars['String']['input'];
};

export type RenameGroupPayload = {
  __typename?: 'RenameGroupPayload';
  error?: Maybe<RenameGroupError>;
  group?: Maybe<ModelGroup>;
};

export type ReorderGroupError = GroupNotFound;

export type ReorderGroupInput = {
  afterGroupId?: InputMaybe<Scalars['ID']['input']>;
  groupId: Scalars['ID']['input'];
};

export type ReorderGroupPayload = {
  __typename?: 'ReorderGroupPayload';
  error?: Maybe<ReorderGroupError>;
  success: Scalars['Boolean']['output'];
};

export type RepairMode =
  | 'ADDITIVE'
  | 'DRY_RUN'
  | 'FULL_SYNC';

export type RepairModelInput = {
  deleteExtraFields?: InputMaybe<Scalars['Boolean']['input']>;
  mode: RepairMode;
  modelId: Scalars['ID']['input'];
};

export type RepairModelPayload = {
  __typename?: 'RepairModelPayload';
  changesApplied: Scalars['Boolean']['output'];
  detectedIssues: Array<SchemaIssue>;
  executedDDL: Array<Scalars['String']['output']>;
  extraFieldsRemoved: Array<Scalars['String']['output']>;
  fieldsAdded: Array<Scalars['String']['output']>;
  healthStatusAfter: HealthStatus;
  healthStatusBefore: HealthStatus;
  model?: Maybe<Model>;
};

export type RevokeApiKeyError = ApiKeyNotFound;

export type RevokeApiKeyPayload = {
  __typename?: 'RevokeApiKeyPayload';
  apiKey?: Maybe<ApiKey>;
  error?: Maybe<RevokeApiKeyError>;
};

export type RevokeRoleError = PermissionRoleNotFound | PermissionUserNotFound;

export type RevokeRolePayload = {
  __typename?: 'RevokeRolePayload';
  error?: Maybe<RevokeRoleError>;
  success: Scalars['Boolean']['output'];
};

export type Role = {
  __typename?: 'Role';
  createdAt: Scalars['String']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  isSystem: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  permissions: Array<Scalars['String']['output']>;
  updatedAt: Scalars['String']['output'];
};

export type RoleAlreadyExists = Error & {
  __typename?: 'RoleAlreadyExists';
  message: Scalars['String']['output'];
};

export type RoleNotFound = Error & {
  __typename?: 'RoleNotFound';
  message: Scalars['String']['output'];
};

export type RolePermissionError = PermissionInvalidInput | PermissionRoleNotFound | PermissionSystemRoleCannotBeModified;

export type SchemaIssue = {
  __typename?: 'SchemaIssue';
  description: Scalars['String']['output'];
  details?: Maybe<Scalars['String']['output']>;
  fieldName?: Maybe<Scalars['String']['output']>;
  tableName: Scalars['String']['output'];
  type: SchemaIssueType;
};

export type SchemaIssueType =
  | 'CLUSTER_NOT_FOUND'
  | 'DATABASE_MISSING'
  | 'FIELD_CONSTRAINT_MISMATCH'
  | 'FIELD_HAS_DEPENDENCIES'
  | 'FIELD_MISSING'
  | 'FIELD_TYPE_MISMATCH'
  | 'TABLE_MISSING';

export type SchemaType =
  | 'ARRAY'
  | 'BOOLEAN'
  | 'NUMBER'
  | 'OBJECT'
  | 'STRING';

export type SyncModelSchemaInput = {
  deleteExtraFields?: InputMaybe<Scalars['Boolean']['input']>;
  id: Scalars['ID']['input'];
  schema: Scalars['String']['input'];
};

export type SyncModelSchemaPayload = {
  __typename?: 'SyncModelSchemaPayload';
  deletedFields: Array<Scalars['String']['output']>;
  fieldsAdded: Scalars['Int64']['output'];
  fieldsDeleted: Scalars['Int64']['output'];
  fieldsSkipped: Array<Scalars['String']['output']>;
  model?: Maybe<Model>;
};

export type TableInfo = {
  __typename?: 'TableInfo';
  name: Scalars['String']['output'];
};

export type TableListConnection = {
  __typename?: 'TableListConnection';
  items: Array<TableInfo>;
  totalCount: Scalars['Int']['output'];
};

export type TestConnectionError = ClusterNotFound | DatabaseConnectionFailed | ProjectNotFound;

export type TestConnectionPayload = {
  __typename?: 'TestConnectionPayload';
  connectionTime?: Maybe<Scalars['Float']['output']>;
  error?: Maybe<TestConnectionError>;
  success: Scalars['Boolean']['output'];
};

export type TestDatabaseConnectionInput = {
  connectionInfo?: InputMaybe<DatabaseConnectionInput>;
  projectSlug?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateApiKeyError = ApiKeyInvalidInput | ApiKeyNotFound;

export type UpdateApiKeyInput = {
  expiresAt?: InputMaybe<Scalars['Time']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateApiKeyPayload = {
  __typename?: 'UpdateApiKeyPayload';
  apiKey?: Maybe<ApiKey>;
  error?: Maybe<UpdateApiKeyError>;
};

export type UpdateClusterConnectionInput = {
  connectionInfo?: InputMaybe<DatabaseConnectionInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  skipConnectionTest?: InputMaybe<Scalars['Boolean']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateClusterError = ClusterNotFound | DatabaseConnectionFailed | InvalidClusterInput | ProjectNotFound;

export type UpdateClusterPayload = {
  __typename?: 'UpdateClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<UpdateClusterError>;
};

export type UpdateEnumError = EnumNotFound | InvalidEnumInput | ProjectNotFound;

export type UpdateEnumInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  options?: InputMaybe<Array<EnumOptionInput>>;
};

export type UpdateEnumPayload = {
  __typename?: 'UpdateEnumPayload';
  enum?: Maybe<EnumDefinition>;
  error?: Maybe<UpdateEnumError>;
};

export type UpdateFieldInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
  validationConfig?: InputMaybe<ValidationConfigInput>;
};

export type UpdateModelError = InvalidModelInput | ModelNotFound | ProjectNotFound;

export type UpdateModelMetaInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateModelMetaPayload = {
  __typename?: 'UpdateModelMetaPayload';
  error?: Maybe<UpdateModelError>;
  model?: Maybe<Model>;
  success: Scalars['Boolean']['output'];
};

export type UpdateOrganizationInput = {
  displayName?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateOrganizationPayload = {
  __typename?: 'UpdateOrganizationPayload';
  error?: Maybe<GetOrganizationError>;
  organization?: Maybe<Organization>;
};

export type UpdatePermissionRoleError = PermissionInvalidInput | PermissionRoleAlreadyExists | PermissionRoleNotFound | PermissionSystemRoleCannotBeModified;

export type UpdatePermissionRolePayload = {
  __typename?: 'UpdatePermissionRolePayload';
  error?: Maybe<UpdatePermissionRoleError>;
  role?: Maybe<PermissionRole>;
};

export type UpdateProjectError = InvalidProjectInput | ProjectNotFound;

export type UpdateProjectInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  loginUrl?: InputMaybe<Scalars['String']['input']>;
  slug: Scalars['String']['input'];
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateProjectPayload = {
  __typename?: 'UpdateProjectPayload';
  error?: Maybe<UpdateProjectError>;
  project?: Maybe<Project>;
};

export type UpdateRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UserRoleAssignment = {
  __typename?: 'UserRoleAssignment';
  createdAt: Scalars['Time']['output'];
  id: Scalars['Int']['output'];
  orgName: Scalars['String']['output'];
  roleId: Scalars['Int']['output'];
  userId: Scalars['String']['output'];
};

export type ValidationConfig = {
  __typename?: 'ValidationConfig';
  maxLength?: Maybe<Scalars['Int64']['output']>;
  maximum?: Maybe<Scalars['Float']['output']>;
  minLength?: Maybe<Scalars['Int64']['output']>;
  minimum?: Maybe<Scalars['Float']['output']>;
  pattern?: Maybe<Scalars['String']['output']>;
};

export type ValidationConfigInput = {
  maxLength?: InputMaybe<Scalars['Int64']['input']>;
  maximum?: InputMaybe<Scalars['Float']['input']>;
  minLength?: InputMaybe<Scalars['Int64']['input']>;
  minimum?: InputMaybe<Scalars['Float']['input']>;
  pattern?: InputMaybe<Scalars['String']['input']>;
};

export type TestClusterConnectionMutationVariables = Exact<{
  input: TestDatabaseConnectionInput;
}>;


export type TestClusterConnectionMutation = { __typename?: 'Mutation', testDatabaseConnection: { __typename?: 'TestConnectionPayload', success: boolean, connectionTime?: number | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type CreateEnumMutationVariables = Exact<{
  input: CreateEnumInput;
}>;


export type CreateEnumMutation = { __typename?: 'Mutation', createEnum: { __typename?: 'CreateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'InvalidEnumInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateEnumMutationVariables = Exact<{
  name: Scalars['String']['input'];
  input: UpdateEnumInput;
}>;


export type UpdateEnumMutation = { __typename?: 'Mutation', updateEnum: { __typename?: 'UpdateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumNotFound', message: string }
      | { __typename: 'InvalidEnumInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type DeleteEnumMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type DeleteEnumMutation = { __typename?: 'Mutation', deleteEnum: { __typename?: 'DeleteEnumPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteReferencedEnum', message: string, suggestion?: string | null }
      | { __typename: 'EnumNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type CreateModelMutationVariables = Exact<{
  input: CreateModelInput;
}>;


export type CreateModelMutation = { __typename?: 'Mutation', createModel: { __typename?: 'CreateModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidModelInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'ModelTableAlreadyExists' }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateModelMetaMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateModelMetaInput;
}>;


export type UpdateModelMetaMutation = { __typename?: 'Mutation', updateModelMeta: { __typename?: 'UpdateModelMetaPayload', success: boolean, model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidModelInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type DeleteModelMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteModelMutation = { __typename?: 'Mutation', deleteModel: { __typename?: 'DeleteModelPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteDeployedModel', message: string, suggestion?: string | null }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type SyncModelSchemaMutationVariables = Exact<{
  input: SyncModelSchemaInput;
}>;


export type SyncModelSchemaMutation = { __typename?: 'Mutation', syncModelSchema: { __typename?: 'SyncModelSchemaPayload', fieldsAdded: any, fieldsSkipped: Array<string>, fieldsDeleted: any, deletedFields: Array<string>, model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null } };

export type RepairModelMutationVariables = Exact<{
  input: RepairModelInput;
}>;


export type RepairModelMutation = { __typename?: 'Mutation', repairModel: { __typename?: 'RepairModelPayload', changesApplied: boolean, executedDDL: Array<string>, healthStatusBefore: HealthStatus, healthStatusAfter: HealthStatus, extraFieldsRemoved: Array<string>, fieldsAdded: Array<string>, model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, detectedIssues: Array<{ __typename?: 'SchemaIssue', type: SchemaIssueType, description: string, tableName: string, fieldName?: string | null, details?: string | null }> } };

export type CreateModelFromSchemaMutationVariables = Exact<{
  input: CreateModelFromSchemaInput;
}>;


export type CreateModelFromSchemaMutation = { __typename?: 'Mutation', createModelFromSchema: { __typename?: 'CreateModelFromSchemaPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null } };

export type ImportModelMutationVariables = Exact<{
  input: ImportModelInput;
}>;


export type ImportModelMutation = { __typename?: 'Mutation', importModel: { __typename?: 'ImportModelPayload', modelId: string, modelName: string, fieldsCount: any, skippedFields: Array<string> } };

export type CreateGroupMutationVariables = Exact<{
  input: CreateGroupInput;
}>;


export type CreateGroupMutation = { __typename?: 'Mutation', createGroup: { __typename?: 'CreateGroupPayload', group?: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string, models: Array<{ __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, createdAt: string, updatedAt: string }> } | null, error?:
      | { __typename: 'GroupAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'InvalidGroupName', message: string, suggestion?: string | null }
     | null } };

export type RenameGroupMutationVariables = Exact<{
  input: RenameGroupInput;
}>;


export type RenameGroupMutation = { __typename?: 'Mutation', renameGroup: { __typename?: 'RenameGroupPayload', group?: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } | null, error?:
      | { __typename: 'GroupAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'GroupNotFound', message: string }
      | { __typename: 'InvalidGroupName', message: string, suggestion?: string | null }
     | null } };

export type DeleteGroupMutationVariables = Exact<{
  groupId: Scalars['ID']['input'];
}>;


export type DeleteGroupMutation = { __typename?: 'Mutation', deleteGroup: { __typename?: 'DeleteGroupPayload', success: boolean, error?: { __typename: 'GroupNotFound', message: string } | null } };

export type ReorderGroupMutationVariables = Exact<{
  input: ReorderGroupInput;
}>;


export type ReorderGroupMutation = { __typename?: 'Mutation', reorderGroup: { __typename?: 'ReorderGroupPayload', success: boolean, error?: { __typename: 'GroupNotFound', message: string } | null } };

export type MoveModelToGroupMutationVariables = Exact<{
  input: MoveModelToGroupInput;
}>;


export type MoveModelToGroupMutation = { __typename?: 'Mutation', moveModelToGroup: { __typename?: 'MoveModelToGroupPayload', success: boolean, error?:
      | { __typename: 'GroupNotFound', message: string }
      | { __typename: 'ModelNotFound', message: string }
     | null } };

export type AddFieldsMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  input: Array<AddFieldInput> | AddFieldInput;
}>;


export type AddFieldsMutation = { __typename?: 'Mutation', addFields?: { __typename?: 'Model', id: string } | null };

export type UpdateFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
  input: UpdateFieldInput;
}>;


export type UpdateFieldMutation = { __typename?: 'Mutation', updateField?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, isDeprecated: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null };

export type RemoveFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
}>;


export type RemoveFieldMutation = { __typename?: 'Mutation', removeField?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null };

export type CreateLogicalForeignKeyMutationVariables = Exact<{
  input: CreateLogicalForeignKeyInput;
}>;


export type CreateLogicalForeignKeyMutation = { __typename?: 'Mutation', createLogicalForeignKey: { __typename?: 'CreateLogicalForeignKeyPayload', result:
      | { __typename: 'FKColumnsNotFoundError', message: string }
      | { __typename: 'FKFieldCountMismatchError', message: string }
      | { __typename: 'LogicalForeignKey', id: string, pairId: string, direction: FkDirection, modelId: string, refModelId: string, sourceFields: Array<string>, targetFields: Array<string> }
     } };

export type DeleteLogicalForeignKeyMutationVariables = Exact<{
  pairId: Scalars['String']['input'];
}>;


export type DeleteLogicalForeignKeyMutation = { __typename?: 'Mutation', deleteLogicalForeignKey: { __typename?: 'DeleteLogicalForeignKeyPayload', result:
      | { __typename: 'DeleteLogicalForeignKeySuccess', pairId: string }
      | { __typename: 'FKNotFoundError', message: string }
      | { __typename: 'FKPairHasRelateFieldsError', message: string }
     } };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'CreateProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, loginUrl?: string | null, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'InvalidProjectInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectAlreadyExists', message: string, suggestion?: string | null }
     | null } };

export type UpdateProjectMutationVariables = Exact<{
  input: UpdateProjectInput;
}>;


export type UpdateProjectMutation = { __typename?: 'Mutation', updateProject: { __typename?: 'UpdateProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, loginUrl?: string | null, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'InvalidProjectInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type DeleteProjectMutationVariables = Exact<{
  slug: Scalars['String']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: { __typename?: 'DeleteProjectPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteDefaultProject', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateProjectClusterMutationVariables = Exact<{
  projectSlug: Scalars['String']['input'];
  input: UpdateClusterConnectionInput;
}>;


export type UpdateProjectClusterMutation = { __typename?: 'Mutation', updateProjectCluster: { __typename?: 'UpdateClusterPayload', cluster?: { __typename?: 'DatabaseCluster', id: string, projectSlug: string, title: string, description: string, status: ClusterStatus, createdAt: string, updatedAt: string, connectionInfo: { __typename?: 'DatabaseConnectionInfo', host: string, port: any, username: string, password: string } } | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'InvalidClusterInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type TestDatabaseConnectionMutationVariables = Exact<{
  input: TestDatabaseConnectionInput;
}>;


export type TestDatabaseConnectionMutation = { __typename?: 'Mutation', testDatabaseConnection: { __typename?: 'TestConnectionPayload', success: boolean, connectionTime?: number | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateOrganizationMutationVariables = Exact<{
  input: UpdateOrganizationInput;
}>;


export type UpdateOrganizationMutation = { __typename?: 'Mutation', updateOrganization: { __typename?: 'UpdateOrganizationPayload', organization?: { __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string } | null, error?: { __typename: 'OrganizationNotFound', message: string } | null } };

export type CreateRoleMutationVariables = Exact<{
  input: CreateRoleInput;
}>;


export type CreateRoleMutation = { __typename?: 'Mutation', createRole: { __typename?: 'CreateRolePayload', role?: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'InvalidRoleInput', message: string, suggestion?: string | null }
      | { __typename: 'RoleAlreadyExists', message: string }
     | null } };

export type DeleteRoleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteRoleMutation = { __typename?: 'Mutation', deleteRole: { __typename?: 'DeleteRolePayload', success: boolean, error?:
      | { __typename: 'CannotDeleteSystemRole', message: string }
      | { __typename: 'RoleNotFound', message: string }
     | null } };

export type GetClusterQueryVariables = Exact<{
  projectSlug: Scalars['String']['input'];
}>;


export type GetClusterQuery = { __typename?: 'Query', databaseCluster: { __typename?: 'GetClusterPayload', cluster?: { __typename?: 'DatabaseCluster', id: string, title: string, description: string, status: ClusterStatus, createdAt: string, updatedAt: string, connectionInfo: { __typename?: 'DatabaseConnectionInfo', host: string, port: any, username: string, password: string } } | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type ListDatabasesQueryVariables = Exact<{
  input: ListDatabasesInput;
}>;


export type ListDatabasesQuery = { __typename?: 'Query', listDatabases: { __typename?: 'DatabaseConnection', totalCount: number, edges: Array<{ __typename?: 'DatabaseEdge', node: { __typename?: 'Database', name: string } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } };

export type GetEnumsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetEnumsQuery = { __typename?: 'Query', enums: Array<{ __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> }> };

export type GetEnumQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetEnumQuery = { __typename?: 'Query', enum: { __typename?: 'GetEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type GetEnumReferencesQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetEnumReferencesQuery = { __typename?: 'Query', enumReferences: Array<string> };

export type GetModelsQueryVariables = Exact<{
  input?: InputMaybe<ModelQueryInput>;
}>;


export type GetModelsQuery = { __typename?: 'Query', models: { __typename?: 'ModelConnection', totalCount: number, edges: Array<{ __typename?: 'ModelEdge', cursor: string, node: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } };

export type GetModelQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetModelQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, isDeprecated: boolean, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, dbColumn?: { __typename?: 'DbColumnInfo', columnType: string, unique: boolean, nonNull: boolean, defaultValue?: string | null, constraints: Array<ActualConstraintType>, foreignKey?: { __typename?: 'ActualForeignKey', referencedTable: string, referencedColumn: string, constraintName: string } | null, conflicts: Array<{ __typename?: 'FieldConflict', aspect: FieldConflictAspect, expected: string, actual: string }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidModelInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type GetModelByNameQueryVariables = Exact<{
  name: Scalars['String']['input'];
  databaseName: Scalars['String']['input'];
}>;


export type GetModelByNameQuery = { __typename?: 'Query', modelByName: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidModelInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type GetModelGroupsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetModelGroupsQuery = { __typename?: 'Query', modelGroups: Array<{ __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string, models: Array<{ __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, createdAt: string, updatedAt: string }> }> };

export type GetModelJsonSchemaQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetModelJsonSchemaQuery = { __typename?: 'Query', modelJsonSchema?: { __typename?: 'ModelJsonSchema', modelId: string, modelName: string, schema: string } | null };

export type GetLogicalForeignKeysQueryVariables = Exact<{
  modelId: Scalars['ID']['input'];
}>;


export type GetLogicalForeignKeysQuery = { __typename?: 'Query', logicalForeignKeys: Array<{ __typename?: 'LogicalForeignKey', id: string, pairId: string, direction: FkDirection, modelId: string, modelName: string, refModelId: string, refModelName: string, sourceFields: Array<string>, targetFields: Array<string> }> };

export type ListTablesQueryVariables = Exact<{
  input: ListTablesInput;
}>;


export type ListTablesQuery = { __typename?: 'Query', listTables: { __typename?: 'TableListConnection', totalCount: number, items: Array<{ __typename?: 'TableInfo', name: string }> } };

export type GetProjectsQueryVariables = Exact<{
  input?: InputMaybe<ListProjectsInput>;
}>;


export type GetProjectsQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, slug: string, title: string, description: string, loginUrl?: string | null, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string }> };

export type GetProjectQueryVariables = Exact<{
  slug: Scalars['String']['input'];
}>;


export type GetProjectQuery = { __typename?: 'Query', project: { __typename?: 'GetProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, loginUrl?: string | null, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?: { __typename: 'ProjectNotFound', message: string } | null } };

export type GetMeQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMeQuery = { __typename?: 'Query', me: { __typename?: 'CurrentUser', id: string, externalID: string, email: string, name: string, permissions: Array<string>, organization?: { __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string } | null, role?: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string } | null } };

export type GetMyOrganizationsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMyOrganizationsQuery = { __typename?: 'Query', myOrganizations: Array<{ __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string }> };

export type GetOrganizationMembersQueryVariables = Exact<{ [key: string]: never; }>;


export type GetOrganizationMembersQuery = { __typename?: 'Query', organizationMembers: Array<{ __typename?: 'OrganizationMember', id: string, userID: string, userName: string, orgID: string, status: MembershipStatus, joinedAt?: string | null, createdAt: string, role: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean } }> };

export type GetRolesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetRolesQuery = { __typename?: 'Query', roles: Array<{ __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string }> };

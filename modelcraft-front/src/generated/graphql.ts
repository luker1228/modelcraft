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

export type ApiTokenLimitReached = Error & {
  __typename?: 'APITokenLimitReached';
  limit: Scalars['Int']['output'];
  message: Scalars['String']['output'];
};

export type ApiTokenNameConflict = Error & {
  __typename?: 'APITokenNameConflict';
  message: Scalars['String']['output'];
};

export type ApiTokenNotFound = Error & {
  __typename?: 'APITokenNotFound';
  message: Scalars['String']['output'];
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
  format: FormatType;
  isArray?: InputMaybe<Scalars['Boolean']['input']>;
  isUnique?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  nonNull?: InputMaybe<Scalars['Boolean']['input']>;
  relateEnumName?: InputMaybe<Scalars['String']['input']>;
  relateFkId?: InputMaybe<Scalars['String']['input']>;
  required?: InputMaybe<Scalars['Boolean']['input']>;
  storageHint?: InputMaybe<Scalars['String']['input']>;
  title: Scalars['String']['input'];
  validationConfig?: InputMaybe<ValidationConfigInput>;
};

export type AddFieldItemResult = {
  __typename?: 'AddFieldItemResult';
  error?: Maybe<AddFieldsError>;
  name: Scalars['String']['output'];
  success: Scalars['Boolean']['output'];
};

export type AddFieldsError = InvalidInput;

export type AddFieldsPayload = {
  __typename?: 'AddFieldsPayload';
  error?: Maybe<AddFieldsError>;
  model?: Maybe<Model>;
  results: Array<AddFieldItemResult>;
};

export type AddRolePermissionPayload = {
  __typename?: 'AddRolePermissionPayload';
  error?: Maybe<RolePermissionError>;
  success: Scalars['Boolean']['output'];
};

export type AssignRoleError = InvalidInput | ResourceNotFound;

export type AssignRolePayload = {
  __typename?: 'AssignRolePayload';
  error?: Maybe<AssignRoleError>;
  userRole?: Maybe<UserRoleAssignment>;
};

export type BatchRegisterError = {
  __typename?: 'BatchRegisterError';
  message: Scalars['String']['output'];
  name: Scalars['String']['output'];
};

export type BatchRegisterModelDatabaseInput = {
  databases: Array<RegisterModelDatabaseInput>;
};

export type BatchRegisterModelDatabaseResult = {
  __typename?: 'BatchRegisterModelDatabaseResult';
  failed: Array<BatchRegisterError>;
  succeeded: Array<ModelDatabase>;
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

export type ClusterStatus =
  | 'ACTIVE'
  | 'DISABLED';

export type CreateApiTokenError = ApiTokenLimitReached | ApiTokenNameConflict | InvalidInput;

export type CreateApiTokenPayload = {
  __typename?: 'CreateAPITokenPayload';
  error?: Maybe<CreateApiTokenError>;
  plaintext?: Maybe<Scalars['String']['output']>;
  token?: Maybe<UserApiToken>;
};

export type CreateCustomRoleError = InvalidInput | PermissionRoleAlreadyExists;

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

export type CreateEnumError = EnumAlreadyExists | InvalidInput | ResourceNotFound;

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
  createMode?: InputMaybe<FkCreateMode>;
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

export type CreateModelError = InvalidInput | ModelAlreadyExists | ModelTableAlreadyExists | ResourceNotFound;

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
  displayField?: InputMaybe<Scalars['String']['input']>;
  insertionOrderField?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type CreateModelPayload = {
  __typename?: 'CreateModelPayload';
  error?: Maybe<CreateModelError>;
  model?: Maybe<Model>;
};

export type CreateProjectError = DatabaseConnectionFailed | InvalidInput | ProjectAlreadyExists;

export type CreateProjectInput = {
  clusterInput: ClusterConnectionInput;
  description?: InputMaybe<Scalars['String']['input']>;
  skipConnectionTest?: InputMaybe<Scalars['Boolean']['input']>;
  slug: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type CreateProjectPayload = {
  __typename?: 'CreateProjectPayload';
  error?: Maybe<CreateProjectError>;
  project?: Maybe<Project>;
};

export type CreateRoleError = InvalidInput | RoleAlreadyExists;

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

export type DangerousPolicyNotConfirmed = Error & {
  __typename?: 'DangerousPolicyNotConfirmed';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
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

export type DatabaseLite = {
  __typename?: 'DatabaseLite';
  name: Scalars['String']['output'];
};

export type DatabaseMode =
  | 'MANAGED'
  | 'SELF_HOSTED';

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

export type DeleteClusterError = ResourceNotFound;

export type DeleteClusterPayload = {
  __typename?: 'DeleteClusterPayload';
  error?: Maybe<DeleteClusterError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEnumError = CannotDeleteReferencedEnum | ResourceNotFound;

export type DeleteEnumPayload = {
  __typename?: 'DeleteEnumPayload';
  error?: Maybe<DeleteEnumError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteGroupError = ResourceNotFound;

export type DeleteGroupPayload = {
  __typename?: 'DeleteGroupPayload';
  error?: Maybe<DeleteGroupError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteLogicalForeignKeyPayload = {
  __typename?: 'DeleteLogicalForeignKeyPayload';
  result: DeleteLogicalForeignKeyResult;
};

export type DeleteLogicalForeignKeyResult = DeleteLogicalForeignKeySuccess | FkNotDeletableError | FkNotFoundError | FkPairHasRelateFieldsError;

export type DeleteLogicalForeignKeySuccess = {
  __typename?: 'DeleteLogicalForeignKeySuccess';
  pairId: Scalars['String']['output'];
};

export type DeleteModelError = CannotDeleteDeployedModel | ResourceNotFound;

export type DeleteModelPayload = {
  __typename?: 'DeleteModelPayload';
  error?: Maybe<DeleteModelError>;
  success: Scalars['Boolean']['output'];
};

export type DeletePermissionRoleError = PermissionSystemRoleCannotBeModified | ResourceNotFound;

export type DeletePermissionRolePayload = {
  __typename?: 'DeletePermissionRolePayload';
  error?: Maybe<DeletePermissionRoleError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteProjectError = CannotDeleteDefaultProject | ResourceNotFound;

export type DeleteProjectPayload = {
  __typename?: 'DeleteProjectPayload';
  error?: Maybe<DeleteProjectError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteRlsPolicyError = ResourceNotFound;

export type DeleteRlsPolicyPayload = {
  __typename?: 'DeleteRlsPolicyPayload';
  error?: Maybe<DeleteRlsPolicyError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteRoleError = CannotDeleteSystemRole | ResourceNotFound;

export type DeleteRolePayload = {
  __typename?: 'DeleteRolePayload';
  error?: Maybe<DeleteRoleError>;
  success: Scalars['Boolean']['output'];
};

export type EndUserRefAlreadyExists = Error & {
  __typename?: 'EndUserRefAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EnumAlreadyExists = Error & {
  __typename?: 'EnumAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
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
  orgName: Scalars['String']['output'];
  projectSlug: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
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

export type FkCreateMode =
  | 'BIDIRECTIONAL'
  | 'UNIDIRECTIONAL';

export type FkDirection =
  | 'NORMAL'
  | 'REVERSE';

export type FkFieldCountMismatchError = {
  __typename?: 'FKFieldCountMismatchError';
  message: Scalars['String']['output'];
};

export type FkNotDeletableError = {
  __typename?: 'FKNotDeletableError';
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
  enumName?: Maybe<Scalars['String']['output']>;
  format: FormatType;
  isArray: Scalars['Boolean']['output'];
  isDeprecated: Scalars['Boolean']['output'];
  isPrimary: Scalars['Boolean']['output'];
  isUnique: Scalars['Boolean']['output'];
  metadata?: Maybe<Scalars['String']['output']>;
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

export type FieldFormatImmutable = Error & {
  __typename?: 'FieldFormatImmutable';
  code: Scalars['String']['output'];
  message: Scalars['String']['output'];
};

export type FieldReferenceInUse = Error & {
  __typename?: 'FieldReferenceInUse';
  code: Scalars['String']['output'];
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type FormatType =
  | 'BOOLEAN'
  | 'DATE'
  | 'DATETIME'
  | 'DECIMAL'
  | 'END_USER_REF'
  | 'ENUM'
  | 'INTEGER'
  | 'NUMBER'
  | 'RELATION'
  | 'STRING'
  | 'TIME'
  | 'UUID';

export type GetClusterError = ResourceNotFound;

export type GetClusterPayload = {
  __typename?: 'GetClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<GetClusterError>;
};

export type GetEnumError = ResourceNotFound;

export type GetEnumPayload = {
  __typename?: 'GetEnumPayload';
  enum?: Maybe<EnumDefinition>;
  error?: Maybe<GetEnumError>;
};

export type GetModelError = InvalidInput | ResourceNotFound;

export type GetModelPayload = {
  __typename?: 'GetModelPayload';
  error?: Maybe<GetModelError>;
  model?: Maybe<Model>;
};

export type GetMyUserProfileError = ResourceNotFound;

export type GetMyUserProfilePayload = {
  __typename?: 'GetMyUserProfilePayload';
  error?: Maybe<GetMyUserProfileError>;
  user?: Maybe<User>;
};

export type GetOrganizationError = ResourceNotFound;

export type GetOrganizationPayload = {
  __typename?: 'GetOrganizationPayload';
  error?: Maybe<GetOrganizationError>;
  organization?: Maybe<Organization>;
};

export type GetProjectError = ResourceNotFound;

export type GetProjectPayload = {
  __typename?: 'GetProjectPayload';
  error?: Maybe<GetProjectError>;
  project?: Maybe<Project>;
};

export type GetRegisteredDatabasesPayload = {
  __typename?: 'GetRegisteredDatabasesPayload';
  data?: Maybe<RegisteredDatabasesPayload>;
  error?: Maybe<RegisteredDatabasesError>;
};

export type GroupAlreadyExists = Error & {
  __typename?: 'GroupAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
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

export type InvalidAuthVariable = Error & {
  __typename?: 'InvalidAuthVariable';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
  variable?: Maybe<Scalars['String']['output']>;
};

export type InvalidGroupName = Error & {
  __typename?: 'InvalidGroupName';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidInput = Error & {
  __typename?: 'InvalidInput';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type InvalidRlsExpression = Error & {
  __typename?: 'InvalidRLSExpression';
  message: Scalars['String']['output'];
  path?: Maybe<Scalars['String']['output']>;
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
  isDeletable: Scalars['Boolean']['output'];
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
  createdVia: Scalars['String']['output'];
  databaseName: Scalars['String']['output'];
  dbTable?: Maybe<DbTableStatus>;
  description: Scalars['String']['output'];
  displayField?: Maybe<Scalars['String']['output']>;
  fields: Array<Field>;
  group: ModelGroup;
  id: Scalars['ID']['output'];
  insertionOrderField?: Maybe<Scalars['String']['output']>;
  isReadOnly: Scalars['Boolean']['output'];
  jsonSchema?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  projectSlug: Scalars['String']['output'];
  /** RLS 策略配置（无 owner 字段时返回 null） */
  rlsPolicy?: Maybe<ModelRlsPolicy>;
  storageType: Scalars['String']['output'];
  title: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
};

export type ModelAlreadyExists = Error & {
  __typename?: 'ModelAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ModelDatabase = {
  __typename?: 'ModelDatabase';
  createdAt: Scalars['Time']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  latestSyncJobId?: Maybe<Scalars['ID']['output']>;
  mode: DatabaseMode;
  name: Scalars['String']['output'];
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type ModelDatabaseSyncFailedTable = {
  __typename?: 'ModelDatabaseSyncFailedTable';
  message: Scalars['String']['output'];
  tableName: Scalars['String']['output'];
};

export type ModelDatabaseSyncJob = {
  __typename?: 'ModelDatabaseSyncJob';
  createdAt: Scalars['Time']['output'];
  createdModels: Scalars['Int']['output'];
  databaseId: Scalars['ID']['output'];
  failedCount: Scalars['Int']['output'];
  failedTables: Array<ModelDatabaseSyncFailedTable>;
  finishedAt?: Maybe<Scalars['Time']['output']>;
  id: Scalars['ID']['output'];
  processedTables: Scalars['Int']['output'];
  startedAt?: Maybe<Scalars['Time']['output']>;
  status: ModelDatabaseSyncJobStatus;
  syncedModels: Scalars['Int']['output'];
  totalTables: Scalars['Int']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type ModelDatabaseSyncJobStatus =
  | 'FAILED'
  | 'PARTIAL_SUCCESS'
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCEEDED';

export type ModelGroup = {
  __typename?: 'ModelGroup';
  displayOrder: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  isVirtual: Scalars['Boolean']['output'];
  models: Array<Model>;
  name: Scalars['String']['output'];
};

export type ModelHasNoOwnerField = Error & {
  __typename?: 'ModelHasNoOwnerField';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type ModelJsonSchema = {
  __typename?: 'ModelJsonSchema';
  modelId: Scalars['ID']['output'];
  modelName: Scalars['String']['output'];
  schema: Scalars['String']['output'];
};

export type ModelListResult = {
  __typename?: 'ModelListResult';
  hasNextPage: Scalars['Boolean']['output'];
  items: Array<Model>;
};

export type ModelQueryInput = {
  databaseName: Scalars['String']['input'];
  pageIndex?: InputMaybe<Scalars['Int']['input']>;
  pageSize?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ModelRlsPolicy = {
  __typename?: 'ModelRLSPolicy';
  /** 创建时间 */
  createdAt: Scalars['String']['output'];
  /** DELETE USING 谓词（JSON 字符串） */
  deletePredicate: Scalars['String']['output'];
  /** INSERT WITH CHECK 谓词（JSON 字符串） */
  insertCheck: Scalars['String']['output'];
  /** 所属模型 ID */
  modelId: Scalars['ID']['output'];
  /** 当前策略匹配的 Preset，自定义组合时返回 null */
  preset?: Maybe<RlsPreset>;
  /** SELECT USING 谓词（JSON 字符串） */
  selectPredicate: Scalars['String']['output'];
  /** UPDATE WITH CHECK 谓词（JSON 字符串） */
  updateCheck: Scalars['String']['output'];
  /** UPDATE USING 谓词（JSON 字符串） */
  updatePredicate: Scalars['String']['output'];
  /** 更新时间 */
  updatedAt: Scalars['String']['output'];
};

export type ModelSyncFailedTable = {
  __typename?: 'ModelSyncFailedTable';
  message: Scalars['String']['output'];
  tableName: Scalars['String']['output'];
};

export type ModelSyncJob = {
  __typename?: 'ModelSyncJob';
  batchId: Scalars['ID']['output'];
  createdAt: Scalars['Time']['output'];
  createdModels: Scalars['Int']['output'];
  databaseId: Scalars['ID']['output'];
  databaseName: Scalars['String']['output'];
  failedCount: Scalars['Int']['output'];
  failedTables: Array<ModelSyncFailedTable>;
  finishedAt?: Maybe<Scalars['Time']['output']>;
  id: Scalars['ID']['output'];
  processedTables: Scalars['Int']['output'];
  startedAt?: Maybe<Scalars['Time']['output']>;
  status: ModelSyncJobStatus;
  syncedModels: Scalars['Int']['output'];
  tableNames: Array<Scalars['String']['output']>;
  totalTables: Scalars['Int']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type ModelSyncJobRef = {
  __typename?: 'ModelSyncJobRef';
  databaseId: Scalars['ID']['output'];
  jobId: Scalars['ID']['output'];
};

export type ModelSyncJobStatus =
  | 'FAILED'
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCEEDED';

export type ModelSyncTargetInput = {
  databaseId: Scalars['ID']['input'];
  tableNames?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type ModelTableAlreadyExists = Error & {
  __typename?: 'ModelTableAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type MoveModelToGroupError = ResourceNotFound;

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
  addFields: AddFieldsPayload;
  addPermissionToRole: AddRolePermissionPayload;
  assignRoleToUser: AssignRolePayload;
  batchRegisterModelDatabases: BatchRegisterModelDatabaseResult;
  createCustomRole: CreateCustomRolePayload;
  createEnum: CreateEnumPayload;
  createGroup: CreateGroupPayload;
  createLogicalForeignKey: CreateLogicalForeignKeyPayload;
  createModel: CreateModelPayload;
  createModelFromSchema: CreateModelFromSchemaPayload;
  createProject: CreateProjectPayload;
  createRole: CreateRolePayload;
  createUserAPIToken: CreateApiTokenPayload;
  deleteEnum: DeleteEnumPayload;
  deleteGroup: DeleteGroupPayload;
  deleteLogicalForeignKey: DeleteLogicalForeignKeyPayload;
  deleteModel: DeleteModelPayload;
  deletePermissionRole: DeletePermissionRolePayload;
  deleteProject: DeleteProjectPayload;
  /** Delete all RLS policies for a model */
  deleteRlsPoliciesByModel: DeleteRlsPolicyPayload;
  /** Delete a single RLS policy by ID */
  deleteRlsPolicy: DeleteRlsPolicyPayload;
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
  registerModelDatabase: RegisterModelDatabaseResult;
  removeField: RemoveFieldPayload;
  removePermissionFromRole: RemoveRolePermissionPayload;
  renameGroup: RenameGroupPayload;
  reorderGroup: ReorderGroupPayload;
  repairModel: RepairModelPayload;
  revokeRoleFromUser: RevokeRolePayload;
  revokeUserAPIToken: RevokeApiTokenPayload;
  /**
   * 设置 Model RLS 策略
   * 支持完整的五件套 JSON 表达式，不限于 Preset
   */
  setModelRLSPolicy: SetModelRlsPolicyPayload;
  /** @deprecated Use startModelSync */
  startModelDatabaseSync: StartModelDatabaseSyncPayload;
  startModelSync: StartModelSyncPayload;
  syncModelSchema: SyncModelSchemaPayload;
  /** @deprecated Use startModelSync */
  syncModelsFromDB: SyncModelsFromDbPayload;
  testDatabaseConnection: TestConnectionPayload;
  /**
   * 解除字段的废弃状态，恢复为正常可用。
   * 若字段未处于废弃状态，幂等返回成功。
   * 状态转换：DEPRECATED → ACTIVE
   */
  undeprecateField?: Maybe<Model>;
  unregisterModelDatabase: Scalars['Boolean']['output'];
  updateEnum: UpdateEnumPayload;
  updateField: UpdateFieldPayload;
  updateModelDatabase: ModelDatabase;
  updateModelMeta: UpdateModelMetaPayload;
  updateMyProfile: UpdateMyProfilePayload;
  updateOrganization: UpdateOrganizationPayload;
  updatePermissionRole: UpdatePermissionRolePayload;
  updateProject: UpdateProjectPayload;
  updateProjectCluster: UpdateClusterPayload;
  /** Create or update an RLS policy */
  upsertRlsPolicy: UpsertRlsPolicyPayload;
  /**
   * 校验 RLS 表达式合法性
   * 用于 Policy 配置页面的实时校验
   */
  validateRLSExpr: ValidateRlsExprPayload;
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


export type MutationBatchRegisterModelDatabasesArgs = {
  input: BatchRegisterModelDatabaseInput;
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


export type MutationCreateUserApiTokenArgs = {
  expiresAt?: InputMaybe<Scalars['Time']['input']>;
  name: Scalars['String']['input'];
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


export type MutationDeleteRlsPoliciesByModelArgs = {
  modelId: Scalars['ID']['input'];
};


export type MutationDeleteRlsPolicyArgs = {
  id: Scalars['ID']['input'];
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


export type MutationRegisterModelDatabaseArgs = {
  input: RegisterModelDatabaseInput;
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


export type MutationRevokeRoleFromUserArgs = {
  orgName: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
  userId: Scalars['String']['input'];
};


export type MutationRevokeUserApiTokenArgs = {
  id: Scalars['ID']['input'];
};


export type MutationSetModelRlsPolicyArgs = {
  input: SetModelRlsPolicyInput;
};


export type MutationStartModelDatabaseSyncArgs = {
  databaseId: Scalars['ID']['input'];
};


export type MutationStartModelSyncArgs = {
  targets: Array<ModelSyncTargetInput>;
};


export type MutationSyncModelSchemaArgs = {
  input: SyncModelSchemaInput;
};


export type MutationSyncModelsFromDbArgs = {
  input: SyncModelsFromDbInput;
};


export type MutationTestDatabaseConnectionArgs = {
  input: TestDatabaseConnectionInput;
};


export type MutationUndeprecateFieldArgs = {
  fieldName: Scalars['String']['input'];
  modelID: Scalars['ID']['input'];
};


export type MutationUnregisterModelDatabaseArgs = {
  id: Scalars['ID']['input'];
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


export type MutationUpdateModelDatabaseArgs = {
  id: Scalars['ID']['input'];
  input: UpdateModelDatabaseInput;
};


export type MutationUpdateModelMetaArgs = {
  id: Scalars['ID']['input'];
  input: UpdateModelMetaInput;
};


export type MutationUpdateMyProfileArgs = {
  input: UpdateMyProfileInput;
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


export type MutationUpsertRlsPolicyArgs = {
  input: RlsPolicyInput;
  modelId: Scalars['ID']['input'];
};


export type MutationValidateRlsExprArgs = {
  input: ValidateRlsExprInput;
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

export type PermissionSystemRoleCannotBeModified = PermissionManagementError & {
  __typename?: 'PermissionSystemRoleCannotBeModified';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type Profile = Node & {
  __typename?: 'Profile';
  avatarUrl?: Maybe<Scalars['String']['output']>;
  bio?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  nickname: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
  userId: Scalars['ID']['output'];
};

export type Project = Node & {
  __typename?: 'Project';
  createdAt: Scalars['String']['output'];
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
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

export type ProjectStatus =
  | 'ACTIVE'
  | 'ARCHIVED';

export type Query = {
  __typename?: 'Query';
  clusterRawDatabases: Array<RawDatabase>;
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
  /** @deprecated Use modelSyncJobs */
  modelDatabaseSyncJob?: Maybe<ModelDatabaseSyncJob>;
  modelDatabases: Array<ModelDatabase>;
  modelGroups: Array<ModelGroup>;
  modelJsonSchema?: Maybe<ModelJsonSchema>;
  /** 获取 Model RLS 策略配置 */
  modelRLSPolicy?: Maybe<ModelRlsPolicy>;
  /** @deprecated Use modelSyncJobs */
  modelSyncJob?: Maybe<ModelSyncJob>;
  modelSyncJobs: Array<ModelSyncJob>;
  models: ModelListResult;
  myOrganizations: Array<Organization>;
  myUserProfile: GetMyUserProfilePayload;
  node?: Maybe<Node>;
  organizationMembers: Array<OrganizationMember>;
  permissionRole?: Maybe<PermissionRole>;
  permissionRoles: Array<PermissionRole>;
  ping: Scalars['String']['output'];
  project: GetProjectPayload;
  projects: Array<Project>;
  registeredDatabases: GetRegisteredDatabasesPayload;
  /** List all RLS policies for a model */
  rlsPolicies: Array<RlsPolicy>;
  rolePermissionsList: Array<PermissionDef>;
  roles: Array<Role>;
  userAPITokens: Array<UserApiToken>;
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


export type QueryModelDatabaseSyncJobArgs = {
  jobId: Scalars['ID']['input'];
};


export type QueryModelJsonSchemaArgs = {
  id: Scalars['ID']['input'];
};


export type QueryModelRlsPolicyArgs = {
  modelId: Scalars['ID']['input'];
};


export type QueryModelSyncJobArgs = {
  jobId: Scalars['ID']['input'];
};


export type QueryModelSyncJobsArgs = {
  batchId?: InputMaybe<Scalars['ID']['input']>;
  jobIds?: InputMaybe<Array<Scalars['ID']['input']>>;
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


export type QueryRegisteredDatabasesArgs = {
  input?: InputMaybe<RegisteredDatabasesInput>;
};


export type QueryRlsPoliciesArgs = {
  modelId: Scalars['ID']['input'];
  orderBy?: InputMaybe<RlsPoliciesOrderBy>;
};


export type QueryRolePermissionsListArgs = {
  roleId: Scalars['Int']['input'];
};


export type QueryUserRoleAssignmentsArgs = {
  orgName: Scalars['String']['input'];
  userId: Scalars['String']['input'];
};

export type RlsCheckViolation = Error & {
  __typename?: 'RLSCheckViolation';
  message: Scalars['String']['output'];
  operation?: Maybe<Scalars['String']['output']>;
};

export type RlsExprDryRun = {
  __typename?: 'RLSExprDryRun';
  params?: Maybe<Array<Scalars['String']['output']>>;
  result?: Maybe<Scalars['Boolean']['output']>;
  sql?: Maybe<Scalars['String']['output']>;
};

export type RlsExprType =
  | 'DELETE_PREDICATE'
  | 'INSERT_CHECK'
  | 'SELECT_PREDICATE'
  | 'UPDATE_CHECK'
  | 'UPDATE_PREDICATE';

export type RlsPreset =
  /**
   * 无访问权限
   * 五件套均为 false
   */
  | 'NO_ACCESS'
  /**
   * 只读全部
   * selectPredicate=true，其余为 false
   */
  | 'READ_ALL'
  /**
   * 读取全部，写自己
   * selectPredicate=true，其余为 OWNER_EQUALS_USER
   */
  | 'READ_ALL_WRITE_OWNER'
  /**
   * 读写全部（⚠️ 高危策略）
   * 五件套均为 true
   */
  | 'READ_WRITE_ALL'
  /**
   * 默认策略：读写自己
   * 五件套均为 {"owner":{"_eq":{"_auth":"uid"}}}
   */
  | 'READ_WRITE_OWNER';

export type RawDatabase = {
  __typename?: 'RawDatabase';
  isRegistered: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
};

export type RegisterModelDatabaseInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  mode: DatabaseMode;
  name: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type RegisterModelDatabaseResult = InvalidInput | ModelDatabase | ResourceNotFound;

export type RegisteredDatabasesError = InvalidInput | ResourceNotFound;

export type RegisteredDatabasesInput = {
  page?: InputMaybe<Scalars['Int']['input']>;
  pageSize?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type RegisteredDatabasesPayload = {
  __typename?: 'RegisteredDatabasesPayload';
  databases: Array<DatabaseLite>;
  page: Scalars['Int']['output'];
  pageSize: Scalars['Int']['output'];
  totalCount: Scalars['Int']['output'];
};

export type RemoveFieldError = FieldReferenceInUse | InvalidInput;

export type RemoveFieldPayload = {
  __typename?: 'RemoveFieldPayload';
  error?: Maybe<RemoveFieldError>;
  model?: Maybe<Model>;
};

export type RemoveRolePermissionPayload = {
  __typename?: 'RemoveRolePermissionPayload';
  error?: Maybe<RolePermissionError>;
  success: Scalars['Boolean']['output'];
};

export type RenameGroupError = GroupAlreadyExists | InvalidGroupName | ResourceNotFound;

export type RenameGroupInput = {
  groupId: Scalars['ID']['input'];
  newName: Scalars['String']['input'];
};

export type RenameGroupPayload = {
  __typename?: 'RenameGroupPayload';
  error?: Maybe<RenameGroupError>;
  group?: Maybe<ModelGroup>;
};

export type ReorderGroupError = ResourceNotFound;

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

export type ResourceNotFound = Error & {
  __typename?: 'ResourceNotFound';
  message: Scalars['String']['output'];
  resourceType: ResourceType;
};

export type ResourceType =
  | 'CLUSTER'
  | 'END_USER'
  | 'END_USER_IN_PROJECT'
  | 'END_USER_PERMISSION'
  | 'END_USER_PERMISSION_BUNDLE'
  | 'END_USER_PERMISSION_BUNDLE_SNAPSHOT'
  | 'END_USER_ROLE'
  | 'ENUM'
  | 'GROUP'
  | 'MODEL'
  | 'ORGANIZATION'
  | 'PERMISSION_ROLE'
  | 'PERMISSION_USER'
  | 'PROFILE'
  | 'PROJECT'
  | 'ROLE'
  | 'UNKNOWN'
  | 'USER';

export type RevokeApiTokenError = ApiTokenNotFound | InvalidInput;

export type RevokeApiTokenPayload = {
  __typename?: 'RevokeAPITokenPayload';
  error?: Maybe<RevokeApiTokenError>;
  success?: Maybe<Scalars['Boolean']['output']>;
};

export type RevokeRoleError = ResourceNotFound;

export type RevokeRolePayload = {
  __typename?: 'RevokeRolePayload';
  error?: Maybe<RevokeRoleError>;
  success: Scalars['Boolean']['output'];
};

export type RlsAction =
  | 'create'
  | 'delete'
  | 'read'
  | 'update';

export type RlsPoliciesOrderBy =
  | 'ACTION_ASC'
  | 'ACTION_DESC'
  | 'ROLE_ASC'
  | 'ROLE_DESC';

export type RlsPolicy = {
  __typename?: 'RlsPolicy';
  action: RlsAction;
  createdAt: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  policyName: Scalars['String']['output'];
  role: Scalars['String']['output'];
  updatedAt: Scalars['String']['output'];
  usingExpr?: Maybe<Scalars['String']['output']>;
  withCheckExpr?: Maybe<Scalars['String']['output']>;
};

export type RlsPolicyInput = {
  action: RlsAction;
  policyName: Scalars['String']['input'];
  role: Scalars['String']['input'];
  usingExpr?: InputMaybe<Scalars['String']['input']>;
  withCheckExpr?: InputMaybe<Scalars['String']['input']>;
};

export type RlsPolicyNotFound = Error & {
  __typename?: 'RlsPolicyNotFound';
  message: Scalars['String']['output'];
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

export type RolePermissionError = InvalidInput | PermissionSystemRoleCannotBeModified | ResourceNotFound;

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

export type SetModelRlsPolicyError = InvalidAuthVariable | InvalidRlsExpression | ModelHasNoOwnerField | ResourceNotFound;

export type SetModelRlsPolicyInput = {
  /** DELETE USING 谓词（JSON 字符串） */
  deletePredicate: Scalars['String']['input'];
  /** INSERT WITH CHECK 谓词（JSON 字符串） */
  insertCheck: Scalars['String']['input'];
  /** 模型 ID */
  modelId: Scalars['ID']['input'];
  /** SELECT USING 谓词（JSON 字符串） */
  selectPredicate: Scalars['String']['input'];
  /** UPDATE WITH CHECK 谓词（JSON 字符串） */
  updateCheck: Scalars['String']['input'];
  /** UPDATE USING 谓词（JSON 字符串） */
  updatePredicate: Scalars['String']['input'];
};

export type SetModelRlsPolicyPayload = {
  __typename?: 'SetModelRLSPolicyPayload';
  error?: Maybe<SetModelRlsPolicyError>;
  policy?: Maybe<ModelRlsPolicy>;
};

export type StartModelDatabaseSyncPayload = {
  __typename?: 'StartModelDatabaseSyncPayload';
  job: ModelDatabaseSyncJob;
};

export type StartModelSyncPayload = {
  __typename?: 'StartModelSyncPayload';
  batchId: Scalars['ID']['output'];
  jobs: Array<ModelSyncJobRef>;
};

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

export type SyncModelsFromDbInput = {
  databaseName: Scalars['String']['input'];
  syncAll?: InputMaybe<Scalars['Boolean']['input']>;
  tableNames?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type SyncModelsFromDbPayload = {
  __typename?: 'SyncModelsFromDBPayload';
  jobId: Scalars['ID']['output'];
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

export type TestConnectionError = DatabaseConnectionFailed | ResourceNotFound;

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

export type UpdateClusterConnectionInput = {
  connectionInfo?: InputMaybe<DatabaseConnectionInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  skipConnectionTest?: InputMaybe<Scalars['Boolean']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateClusterError = DatabaseConnectionFailed | InvalidInput | ResourceNotFound;

export type UpdateClusterPayload = {
  __typename?: 'UpdateClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<UpdateClusterError>;
};

export type UpdateEnumError = InvalidInput | ResourceNotFound;

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

export type UpdateFieldError = FieldFormatImmutable | InvalidInput;

export type UpdateFieldInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
  validationConfig?: InputMaybe<ValidationConfigInput>;
};

export type UpdateFieldPayload = {
  __typename?: 'UpdateFieldPayload';
  error?: Maybe<UpdateFieldError>;
  model?: Maybe<Model>;
};

export type UpdateModelDatabaseInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  mode?: InputMaybe<DatabaseMode>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateModelError = InvalidInput | ResourceNotFound;

export type UpdateModelMetaInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  displayField?: InputMaybe<Scalars['String']['input']>;
  insertionOrderField?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateModelMetaPayload = {
  __typename?: 'UpdateModelMetaPayload';
  error?: Maybe<UpdateModelError>;
  model?: Maybe<Model>;
  success: Scalars['Boolean']['output'];
};

export type UpdateMyProfileError = InvalidInput | ResourceNotFound;

export type UpdateMyProfileInput = {
  avatarUrl?: InputMaybe<Scalars['String']['input']>;
  bio?: InputMaybe<Scalars['String']['input']>;
  nickname?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateMyProfilePayload = {
  __typename?: 'UpdateMyProfilePayload';
  error?: Maybe<UpdateMyProfileError>;
  profile?: Maybe<Profile>;
};

export type UpdateOrganizationInput = {
  displayName?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateOrganizationPayload = {
  __typename?: 'UpdateOrganizationPayload';
  error?: Maybe<GetOrganizationError>;
  organization?: Maybe<Organization>;
};

export type UpdatePermissionRoleError = InvalidInput | PermissionRoleAlreadyExists | PermissionSystemRoleCannotBeModified | ResourceNotFound;

export type UpdatePermissionRolePayload = {
  __typename?: 'UpdatePermissionRolePayload';
  error?: Maybe<UpdatePermissionRoleError>;
  role?: Maybe<PermissionRole>;
};

export type UpdateProjectError = InvalidInput | ResourceNotFound;

export type UpdateProjectInput = {
  description?: InputMaybe<Scalars['String']['input']>;
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

export type UpsertRlsPolicyError = InvalidInput | ResourceNotFound;

export type UpsertRlsPolicyPayload = {
  __typename?: 'UpsertRlsPolicyPayload';
  error?: Maybe<UpsertRlsPolicyError>;
  policy?: Maybe<RlsPolicy>;
};

export type User = Node & {
  __typename?: 'User';
  createdAt: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  phone: Scalars['String']['output'];
  profile: Profile;
  status: UserStatus;
  updatedAt: Scalars['String']['output'];
  userName: Scalars['String']['output'];
};

export type UserApiToken = {
  __typename?: 'UserAPIToken';
  createdAt: Scalars['Time']['output'];
  expiresAt?: Maybe<Scalars['Time']['output']>;
  id: Scalars['ID']['output'];
  lastUsedAt?: Maybe<Scalars['Time']['output']>;
  name: Scalars['String']['output'];
};

export type UserRoleAssignment = {
  __typename?: 'UserRoleAssignment';
  createdAt: Scalars['Time']['output'];
  id: Scalars['Int']['output'];
  orgName: Scalars['String']['output'];
  roleId: Scalars['Int']['output'];
  userId: Scalars['String']['output'];
};

export type UserStatus =
  | 'ACTIVE'
  | 'REGISTERED'
  | 'SUSPENDED';

export type ValidateRlsExprError = InvalidAuthVariable | InvalidRlsExpression | ResourceNotFound;

export type ValidateRlsExprInput = {
  /** 要校验的谓词类型 */
  exprType: RlsExprType;
  /** 表达式 JSON 字符串 */
  expression: Scalars['String']['input'];
  /** 所属模型 ID（用于字段名白名单校验） */
  modelId: Scalars['ID']['input'];
  /** check 表达式 dry run 时使用的示例输入 JSON */
  sampleInput?: InputMaybe<Scalars['String']['input']>;
};

export type ValidateRlsExprPayload = {
  __typename?: 'ValidateRLSExprPayload';
  dryRun?: Maybe<RlsExprDryRun>;
  error?: Maybe<ValidateRlsExprError>;
  result: ValidationResult;
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

export type ValidationError = {
  __typename?: 'ValidationError';
  /** 错误码 */
  code: Scalars['String']['output'];
  /** 错误描述 */
  message: Scalars['String']['output'];
  /** 错误位置（如 "selectPredicate._and[0].owner"） */
  path: Scalars['String']['output'];
};

export type ValidationResult = {
  __typename?: 'ValidationResult';
  /** 校验错误信息列表（valid=false 时返回） */
  errors?: Maybe<Array<ValidationError>>;
  /** 是否校验通过 */
  valid: Scalars['Boolean']['output'];
};

export type GetClusterQueryVariables = Exact<{
  projectSlug: Scalars['String']['input'];
}>;


export type GetClusterQuery = { __typename?: 'Query', databaseCluster: { __typename?: 'GetClusterPayload', cluster?: { __typename?: 'DatabaseCluster', id: string, title: string, description: string, status: ClusterStatus, createdAt: string, updatedAt: string, connectionInfo: { __typename?: 'DatabaseConnectionInfo', host: string, port: any, username: string, password: string } } | null, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type ListDatabasesQueryVariables = Exact<{
  input: ListDatabasesInput;
}>;


export type ListDatabasesQuery = { __typename?: 'Query', listDatabases: { __typename?: 'DatabaseConnection', totalCount: number, edges: Array<{ __typename?: 'DatabaseEdge', node: { __typename?: 'Database', name: string } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } };

export type RegisteredDatabasesQueryVariables = Exact<{
  input?: InputMaybe<RegisteredDatabasesInput>;
}>;


export type RegisteredDatabasesQuery = { __typename?: 'Query', registeredDatabases: { __typename?: 'GetRegisteredDatabasesPayload', data?: { __typename?: 'RegisteredDatabasesPayload', totalCount: number, page: number, pageSize: number, databases: Array<{ __typename?: 'DatabaseLite', name: string }> } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type TestClusterConnectionMutationVariables = Exact<{
  input: TestDatabaseConnectionInput;
}>;


export type TestClusterConnectionMutation = { __typename?: 'Mutation', testDatabaseConnection: { __typename?: 'TestConnectionPayload', success: boolean, connectionTime?: number | null, error?:
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type GetEnumsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetEnumsQuery = { __typename?: 'Query', enums: Array<{ __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> }> };

export type GetEnumQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetEnumQuery = { __typename?: 'Query', enum: { __typename?: 'GetEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type GetEnumReferencesQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetEnumReferencesQuery = { __typename?: 'Query', enumReferences: Array<string> };

export type CreateEnumMutationVariables = Exact<{
  input: CreateEnumInput;
}>;


export type CreateEnumMutation = { __typename?: 'Mutation', createEnum: { __typename?: 'CreateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type UpdateEnumMutationVariables = Exact<{
  name: Scalars['String']['input'];
  input: UpdateEnumInput;
}>;


export type UpdateEnumMutation = { __typename?: 'Mutation', updateEnum: { __typename?: 'UpdateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type DeleteEnumMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type DeleteEnumMutation = { __typename?: 'Mutation', deleteEnum: { __typename?: 'DeleteEnumPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteReferencedEnum', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type GetModelEnumSourceFieldsQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetModelEnumSourceFieldsQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, enum?: { __typename?: 'EnumDefinition', name: string } | null }> } | null } };

export type GetModelsQueryVariables = Exact<{
  input?: InputMaybe<ModelQueryInput>;
}>;


export type GetModelsQuery = { __typename?: 'Query', models: { __typename?: 'ModelListResult', hasNextPage: boolean, items: Array<{ __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, createdVia: string, isReadOnly: boolean, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } }> } };

export type GetModelsByDatabaseQueryVariables = Exact<{
  input?: InputMaybe<ModelQueryInput>;
}>;


export type GetModelsByDatabaseQuery = { __typename?: 'Query', models: { __typename?: 'ModelListResult', items: Array<{ __typename?: 'Model', id: string, name: string, title: string, databaseName: string }> } };

export type GetModelQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetModelQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, displayField?: string | null, insertionOrderField?: string | null, databaseName: string, storageType: string, createdVia: string, isReadOnly: boolean, jsonSchema?: string | null, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, isDeprecated: boolean, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, dbColumn?: { __typename?: 'DbColumnInfo', columnType: string, unique: boolean, nonNull: boolean, defaultValue?: string | null, constraints: Array<ActualConstraintType>, foreignKey?: { __typename?: 'ActualForeignKey', referencedTable: string, referencedColumn: string, constraintName: string } | null, conflicts: Array<{ __typename?: 'FieldConflict', aspect: FieldConflictAspect, expected: string, actual: string }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type GetModelRecordWorkspaceQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetModelRecordWorkspaceQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, name: string, title: string, description: string, databaseName: string, createdVia: string, isReadOnly: boolean, jsonSchema?: string | null, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, isPrimary: boolean, isDeprecated: boolean }> } | null, error?:
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type GetModelByNameQueryVariables = Exact<{
  name: Scalars['String']['input'];
  databaseName: Scalars['String']['input'];
}>;


export type GetModelByNameQuery = { __typename?: 'Query', modelByName: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
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

export type CreateModelMutationVariables = Exact<{
  input: CreateModelInput;
}>;


export type CreateModelMutation = { __typename?: 'Mutation', createModel: { __typename?: 'CreateModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'ModelTableAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type UpdateModelMetaMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateModelMetaInput;
}>;


export type UpdateModelMetaMutation = { __typename?: 'Mutation', updateModelMeta: { __typename?: 'UpdateModelMetaPayload', success: boolean, model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, displayField?: string | null, insertionOrderField?: string | null, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type DeleteModelMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteModelMutation = { __typename?: 'Mutation', deleteModel: { __typename?: 'DeleteModelPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteDeployedModel', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
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
      | { __typename: 'InvalidGroupName', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type DeleteGroupMutationVariables = Exact<{
  groupId: Scalars['ID']['input'];
}>;


export type DeleteGroupMutation = { __typename?: 'Mutation', deleteGroup: { __typename?: 'DeleteGroupPayload', success: boolean, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type ReorderGroupMutationVariables = Exact<{
  input: ReorderGroupInput;
}>;


export type ReorderGroupMutation = { __typename?: 'Mutation', reorderGroup: { __typename?: 'ReorderGroupPayload', success: boolean, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type MoveModelToGroupMutationVariables = Exact<{
  input: MoveModelToGroupInput;
}>;


export type MoveModelToGroupMutation = { __typename?: 'Mutation', moveModelToGroup: { __typename?: 'MoveModelToGroupPayload', success: boolean, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type AddFieldsMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  input: Array<AddFieldInput> | AddFieldInput;
}>;


export type AddFieldsMutation = { __typename?: 'Mutation', addFields: { __typename?: 'AddFieldsPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, isDeprecated: boolean, isArray: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?: { __typename: 'InvalidInput', message: string, suggestion?: string | null } | null } };

export type UpdateFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
  input: UpdateFieldInput;
}>;


export type UpdateFieldMutation = { __typename?: 'Mutation', updateField: { __typename?: 'UpdateFieldPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, isDeprecated: boolean, isArray: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'FieldFormatImmutable', message: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
     | null } };

export type RemoveFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
}>;


export type RemoveFieldMutation = { __typename?: 'Mutation', removeField: { __typename?: 'RemoveFieldPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'FieldReferenceInUse', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
     | null } };

export type DeprecateFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
}>;


export type DeprecateFieldMutation = { __typename?: 'Mutation', deprecateField?: { __typename?: 'Model', id: string } | null };

export type UndeprecateFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
}>;


export type UndeprecateFieldMutation = { __typename?: 'Mutation', undeprecateField?: { __typename?: 'Model', id: string } | null };

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
      | { __typename: 'FKNotDeletableError' }
      | { __typename: 'FKNotFoundError', message: string }
      | { __typename: 'FKPairHasRelateFieldsError', message: string }
     } };

export type SyncModelsFromDbMutationVariables = Exact<{
  input: SyncModelsFromDbInput;
}>;


export type SyncModelsFromDbMutation = { __typename?: 'Mutation', syncModelsFromDB: { __typename?: 'SyncModelsFromDBPayload', jobId: string } };

export type ModelSyncJobQueryVariables = Exact<{
  jobId: Scalars['ID']['input'];
}>;


export type ModelSyncJobQuery = { __typename?: 'Query', modelSyncJob?: { __typename?: 'ModelSyncJob', id: string, databaseName: string, tableNames: Array<string>, status: ModelSyncJobStatus, totalTables: number, processedTables: number, createdModels: number, syncedModels: number, failedCount: number, startedAt?: any | null, finishedAt?: any | null, createdAt: any, updatedAt: any, failedTables: Array<{ __typename?: 'ModelSyncFailedTable', tableName: string, message: string }> } | null };

export type ModelSyncJobsQueryVariables = Exact<{
  jobIds?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
  batchId?: InputMaybe<Scalars['ID']['input']>;
}>;


export type ModelSyncJobsQuery = { __typename?: 'Query', modelSyncJobs: Array<{ __typename?: 'ModelSyncJob', id: string, batchId: string, databaseId: string, databaseName: string, tableNames: Array<string>, status: ModelSyncJobStatus, totalTables: number, processedTables: number, createdModels: number, syncedModels: number, failedCount: number, startedAt?: any | null, finishedAt?: any | null, createdAt: any, updatedAt: any, failedTables: Array<{ __typename?: 'ModelSyncFailedTable', tableName: string, message: string }> }> };

export type StartModelSyncMutationVariables = Exact<{
  targets: Array<ModelSyncTargetInput> | ModelSyncTargetInput;
}>;


export type StartModelSyncMutation = { __typename?: 'Mutation', startModelSync: { __typename?: 'StartModelSyncPayload', batchId: string, jobs: Array<{ __typename?: 'ModelSyncJobRef', databaseId: string, jobId: string }> } };

export type NoopQueryQueryVariables = Exact<{ [key: string]: never; }>;


export type NoopQueryQuery = { __typename: 'Query' };

export type NoopMutationMutationVariables = Exact<{ [key: string]: never; }>;


export type NoopMutationMutation = { __typename: 'Mutation' };

export type ListTablesQueryVariables = Exact<{
  input: ListTablesInput;
}>;


export type ListTablesQuery = { __typename?: 'Query', listTables: { __typename?: 'TableListConnection', totalCount: number, items: Array<{ __typename?: 'TableInfo', name: string }> } };

export type GetProjectsQueryVariables = Exact<{
  input?: InputMaybe<ListProjectsInput>;
}>;


export type GetProjectsQuery = { __typename?: 'Query', projects: Array<{ __typename?: 'Project', id: string, slug: string, title: string, description: string, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string }> };

export type GetProjectQueryVariables = Exact<{
  slug: Scalars['String']['input'];
}>;


export type GetProjectQuery = { __typename?: 'Query', project: { __typename?: 'GetProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'CreateProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectAlreadyExists', message: string, suggestion?: string | null }
     | null } };

export type UpdateProjectMutationVariables = Exact<{
  input: UpdateProjectInput;
}>;


export type UpdateProjectMutation = { __typename?: 'Mutation', updateProject: { __typename?: 'UpdateProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type DeleteProjectMutationVariables = Exact<{
  slug: Scalars['String']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: { __typename?: 'DeleteProjectPayload', success: boolean, error?:
      | { __typename: 'CannotDeleteDefaultProject', message: string }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type UpdateProjectClusterMutationVariables = Exact<{
  projectSlug: Scalars['String']['input'];
  input: UpdateClusterConnectionInput;
}>;


export type UpdateProjectClusterMutation = { __typename?: 'Mutation', updateProjectCluster: { __typename?: 'UpdateClusterPayload', cluster?: { __typename?: 'DatabaseCluster', id: string, projectSlug: string, title: string, description: string, status: ClusterStatus, createdAt: string, updatedAt: string, connectionInfo: { __typename?: 'DatabaseConnectionInfo', host: string, port: any, username: string, password: string } } | null, error?:
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type TestDatabaseConnectionMutationVariables = Exact<{
  input: TestDatabaseConnectionInput;
}>;


export type TestDatabaseConnectionMutation = { __typename?: 'Mutation', testDatabaseConnection: { __typename?: 'TestConnectionPayload', success: boolean, connectionTime?: number | null, error?:
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type ModelDatabaseFieldsFragment = { __typename?: 'ModelDatabase', id: string, name: string, title: string, description: string, mode: DatabaseMode, latestSyncJobId?: string | null, createdAt: any, updatedAt: any };

export type ModelDatabaseSyncJobFieldsFragment = { __typename?: 'ModelDatabaseSyncJob', id: string, databaseId: string, status: ModelDatabaseSyncJobStatus, totalTables: number, processedTables: number, createdModels: number, syncedModels: number, failedCount: number, startedAt?: any | null, finishedAt?: any | null, createdAt: any, updatedAt: any, failedTables: Array<{ __typename?: 'ModelDatabaseSyncFailedTable', tableName: string, message: string }> };

export type ListModelDatabasesQueryVariables = Exact<{ [key: string]: never; }>;


export type ListModelDatabasesQuery = { __typename?: 'Query', modelDatabases: Array<{ __typename?: 'ModelDatabase', id: string, name: string, title: string, description: string, mode: DatabaseMode, latestSyncJobId?: string | null, createdAt: any, updatedAt: any }> };

export type ListClusterRawDatabasesQueryVariables = Exact<{ [key: string]: never; }>;


export type ListClusterRawDatabasesQuery = { __typename?: 'Query', clusterRawDatabases: Array<{ __typename?: 'RawDatabase', name: string, isRegistered: boolean }> };

export type GetModelDatabaseSyncJobQueryVariables = Exact<{
  jobId: Scalars['ID']['input'];
}>;


export type GetModelDatabaseSyncJobQuery = { __typename?: 'Query', modelDatabaseSyncJob?: { __typename?: 'ModelDatabaseSyncJob', id: string, databaseId: string, status: ModelDatabaseSyncJobStatus, totalTables: number, processedTables: number, createdModels: number, syncedModels: number, failedCount: number, startedAt?: any | null, finishedAt?: any | null, createdAt: any, updatedAt: any, failedTables: Array<{ __typename?: 'ModelDatabaseSyncFailedTable', tableName: string, message: string }> } | null };

export type RegisterModelDatabaseMutationVariables = Exact<{
  input: RegisterModelDatabaseInput;
}>;


export type RegisterModelDatabaseMutation = { __typename?: 'Mutation', registerModelDatabase:
    | { __typename?: 'InvalidInput', message: string }
    | { __typename?: 'ModelDatabase', id: string, name: string, title: string, description: string, mode: DatabaseMode, latestSyncJobId?: string | null, createdAt: any, updatedAt: any }
    | { __typename?: 'ResourceNotFound', message: string, resourceType: ResourceType }
   };

export type BatchRegisterModelDatabasesMutationVariables = Exact<{
  input: BatchRegisterModelDatabaseInput;
}>;


export type BatchRegisterModelDatabasesMutation = { __typename?: 'Mutation', batchRegisterModelDatabases: { __typename?: 'BatchRegisterModelDatabaseResult', succeeded: Array<{ __typename?: 'ModelDatabase', id: string, name: string, title: string, description: string, mode: DatabaseMode, latestSyncJobId?: string | null, createdAt: any, updatedAt: any }>, failed: Array<{ __typename?: 'BatchRegisterError', name: string, message: string }> } };

export type UpdateModelDatabaseMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateModelDatabaseInput;
}>;


export type UpdateModelDatabaseMutation = { __typename?: 'Mutation', updateModelDatabase: { __typename?: 'ModelDatabase', id: string, name: string, title: string, description: string, mode: DatabaseMode, latestSyncJobId?: string | null, createdAt: any, updatedAt: any } };

export type UnregisterModelDatabaseMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type UnregisterModelDatabaseMutation = { __typename?: 'Mutation', unregisterModelDatabase: boolean };

export type StartModelDatabaseSyncMutationVariables = Exact<{
  databaseId: Scalars['ID']['input'];
}>;


export type StartModelDatabaseSyncMutation = { __typename?: 'Mutation', startModelDatabaseSync: { __typename?: 'StartModelDatabaseSyncPayload', job: { __typename?: 'ModelDatabaseSyncJob', id: string, databaseId: string, status: ModelDatabaseSyncJobStatus, totalTables: number, processedTables: number, createdModels: number, syncedModels: number, failedCount: number, startedAt?: any | null, finishedAt?: any | null, createdAt: any, updatedAt: any, failedTables: Array<{ __typename?: 'ModelDatabaseSyncFailedTable', tableName: string, message: string }> } } };

export type GetRlsPoliciesQueryVariables = Exact<{
  modelId: Scalars['ID']['input'];
  orderBy?: InputMaybe<RlsPoliciesOrderBy>;
}>;


export type GetRlsPoliciesQuery = { __typename?: 'Query', rlsPolicies: Array<{ __typename?: 'RlsPolicy', id: string, policyName: string, action: RlsAction, role: string, usingExpr?: string | null, withCheckExpr?: string | null, createdAt: string, updatedAt: string }> };

export type UpsertRlsPolicyMutationVariables = Exact<{
  modelId: Scalars['ID']['input'];
  input: RlsPolicyInput;
}>;


export type UpsertRlsPolicyMutation = { __typename?: 'Mutation', upsertRlsPolicy: { __typename?: 'UpsertRlsPolicyPayload', policy?: { __typename?: 'RlsPolicy', id: string, policyName: string, action: RlsAction, role: string, usingExpr?: string | null, withCheckExpr?: string | null, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ResourceNotFound', message: string }
     | null } };

export type DeleteRlsPolicyMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteRlsPolicyMutation = { __typename?: 'Mutation', deleteRlsPolicy: { __typename?: 'DeleteRlsPolicyPayload', success: boolean, error?: { __typename: 'ResourceNotFound', message: string } | null } };

export type ValidateRlsExprMutationVariables = Exact<{
  input: ValidateRlsExprInput;
}>;


export type ValidateRlsExprMutation = { __typename?: 'Mutation', validateRLSExpr: { __typename?: 'ValidateRLSExprPayload', result: { __typename?: 'ValidationResult', valid: boolean, errors?: Array<{ __typename?: 'ValidationError', path: string, message: string, code: string }> | null }, error?:
      | { __typename: 'InvalidAuthVariable', message: string, suggestion?: string | null, variable?: string | null }
      | { __typename: 'InvalidRLSExpression', message: string, suggestion?: string | null, path?: string | null }
      | { __typename: 'ResourceNotFound', message: string }
     | null, dryRun?: { __typename?: 'RLSExprDryRun', sql?: string | null, params?: Array<string> | null, result?: boolean | null } | null } };

export type GetMeQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMeQuery = { __typename?: 'Query', me: { __typename?: 'CurrentUser', id: string, externalID: string, email: string, name: string, permissions: Array<string>, organization?: { __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string } | null, role?: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string } | null } };

export type GetMyOrganizationsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetMyOrganizationsQuery = { __typename?: 'Query', myOrganizations: Array<{ __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string }> };

export type GetOrganizationMembersQueryVariables = Exact<{ [key: string]: never; }>;


export type GetOrganizationMembersQuery = { __typename?: 'Query', organizationMembers: Array<{ __typename?: 'OrganizationMember', id: string, userID: string, userName: string, orgID: string, status: MembershipStatus, joinedAt?: string | null, createdAt: string, role: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean } }> };

export type GetRolesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetRolesQuery = { __typename?: 'Query', roles: Array<{ __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string }> };

export type GetPermissionRolesQueryVariables = Exact<{
  orgName: Scalars['String']['input'];
  includeSystem?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetPermissionRolesQuery = { __typename?: 'Query', permissionRoles: Array<{ __typename?: 'PermissionRole', id: number, name: string, description?: string | null, isSystem: boolean, orgName: string, createdAt: any, updatedAt: any }> };

export type GetRolePermissionsListQueryVariables = Exact<{
  roleId: Scalars['Int']['input'];
}>;


export type GetRolePermissionsListQuery = { __typename?: 'Query', rolePermissionsList: Array<{ __typename?: 'PermissionDef', obj: string, act: string }> };

export type GetApiTokensQueryVariables = Exact<{ [key: string]: never; }>;


export type GetApiTokensQuery = { __typename?: 'Query', userAPITokens: Array<{ __typename?: 'UserAPIToken', id: string, name: string, createdAt: any, expiresAt?: any | null, lastUsedAt?: any | null }> };

export type UpdateOrganizationMutationVariables = Exact<{
  input: UpdateOrganizationInput;
}>;


export type UpdateOrganizationMutation = { __typename?: 'Mutation', updateOrganization: { __typename?: 'UpdateOrganizationPayload', organization?: { __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string } | null, error?: { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType } | null } };

export type CreateRoleMutationVariables = Exact<{
  input: CreateRoleInput;
}>;


export type CreateRoleMutation = { __typename?: 'Mutation', createRole: { __typename?: 'CreateRolePayload', role?: { __typename?: 'Role', id: string, name: string, description?: string | null, permissions: Array<string>, isSystem: boolean, createdAt: string, updatedAt: string } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'RoleAlreadyExists', message: string }
     | null } };

export type DeleteRoleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteRoleMutation = { __typename?: 'Mutation', deleteRole: { __typename?: 'DeleteRolePayload', success: boolean, error?:
      | { __typename: 'CannotDeleteSystemRole', message: string }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type AddPermissionToRoleMutationVariables = Exact<{
  roleId: Scalars['Int']['input'];
  obj: Scalars['String']['input'];
  act: Scalars['String']['input'];
}>;


export type AddPermissionToRoleMutation = { __typename?: 'Mutation', addPermissionToRole: { __typename?: 'AddRolePermissionPayload', success: boolean, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'PermissionSystemRoleCannotBeModified', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type RemovePermissionFromRoleMutationVariables = Exact<{
  roleId: Scalars['Int']['input'];
  obj: Scalars['String']['input'];
  act: Scalars['String']['input'];
}>;


export type RemovePermissionFromRoleMutation = { __typename?: 'Mutation', removePermissionFromRole: { __typename?: 'RemoveRolePermissionPayload', success: boolean, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'PermissionSystemRoleCannotBeModified', message: string, suggestion?: string | null }
      | { __typename: 'ResourceNotFound', message: string, resourceType: ResourceType }
     | null } };

export type CreateApiTokenMutationVariables = Exact<{
  name: Scalars['String']['input'];
  expiresAt?: InputMaybe<Scalars['Time']['input']>;
}>;


export type CreateApiTokenMutation = { __typename?: 'Mutation', createUserAPIToken: { __typename?: 'CreateAPITokenPayload', plaintext?: string | null, token?: { __typename?: 'UserAPIToken', id: string, name: string, createdAt: any, expiresAt?: any | null, lastUsedAt?: any | null } | null, error?:
      | { __typename: 'APITokenLimitReached', message: string, limit: number }
      | { __typename: 'APITokenNameConflict', message: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
     | null } };

export type RevokeApiTokenMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RevokeApiTokenMutation = { __typename?: 'Mutation', revokeUserAPIToken: { __typename?: 'RevokeAPITokenPayload', success?: boolean | null, error?:
      | { __typename: 'APITokenNotFound', message: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
     | null } };

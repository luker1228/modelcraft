import { graphql, type GraphQLResponseResolver, type RequestHandlerOptions } from 'msw'
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

export enum ActualConstraintType {
  NotNull = 'NOT_NULL',
  Unique = 'UNIQUE'
}

export type ActualForeignKey = {
  __typename?: 'ActualForeignKey';
  constraintName: Scalars['String']['output'];
  referencedColumn: Scalars['String']['output'];
  referencedTable: Scalars['String']['output'];
};

export type AddEndUserPermissionToBundleError = EndUserPermissionBundleNotFound | EndUserPermissionNotFound | InvalidInput | ProjectNotFound;

export type AddEndUserPermissionToBundleInput = {
  bundleId: Scalars['ID']['input'];
  permissionId: Scalars['ID']['input'];
  sortOrder?: InputMaybe<Scalars['Int']['input']>;
};

export type AddEndUserPermissionToBundlePayload = {
  __typename?: 'AddEndUserPermissionToBundlePayload';
  bundle?: Maybe<EndUserPermissionBundle>;
  error?: Maybe<AddEndUserPermissionToBundleError>;
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

export type ApplyEndUserPresetPolicyError = ModelNotFound | PresetRequiresOwnerField | ProjectNotFound;

export type ApplyEndUserPresetPolicyInput = {
  modelId: Scalars['ID']['input'];
  preset: EndUserPermissionPreset;
};

export type ApplyEndUserPresetPolicyPayload = {
  __typename?: 'ApplyEndUserPresetPolicyPayload';
  error?: Maybe<ApplyEndUserPresetPolicyError>;
  /** 应用预设后，该模型当前所有的权限点（含原有自定义权限点） */
  permissions: Array<EndUserPermission>;
};

export type AssignBundleToEndUserError = EndUserNotFoundInProject | EndUserPermissionBundleNotFound | ProjectNotFound | UserBundleAlreadyAssigned;

export type AssignBundleToEndUserInput = {
  bundleId: Scalars['ID']['input'];
  endUserId: Scalars['ID']['input'];
};

export type AssignBundleToEndUserPayload = {
  __typename?: 'AssignBundleToEndUserPayload';
  bundle?: Maybe<EndUserPermissionBundle>;
  endUserId: Scalars['ID']['output'];
  error?: Maybe<AssignBundleToEndUserError>;
};

export type AssignBundleToEndUserRoleError = EndUserPermissionBundleNotFound | EndUserRoleNotFound | ProjectNotFound;

export type AssignBundleToEndUserRoleInput = {
  bundleId: Scalars['ID']['input'];
  roleId: Scalars['ID']['input'];
};

export type AssignBundleToEndUserRolePayload = {
  __typename?: 'AssignBundleToEndUserRolePayload';
  error?: Maybe<AssignBundleToEndUserRoleError>;
  role?: Maybe<EndUserRole>;
};

export type AssignEndUserRoleError = EndUserCannotAssignImplicitRole | EndUserNotFoundInProject | EndUserRoleNotFound | ProjectNotFound | UserRoleAlreadyAssigned;

export type AssignEndUserRoleInput = {
  endUserId: Scalars['ID']['input'];
  roleId: Scalars['ID']['input'];
};

export type AssignEndUserRolePayload = {
  __typename?: 'AssignEndUserRolePayload';
  endUserId: Scalars['ID']['output'];
  error?: Maybe<AssignEndUserRoleError>;
  role?: Maybe<EndUserRole>;
};

export type AssignRoleError = InvalidInput | PermissionRoleNotFound | PermissionUserNotFound;

export type AssignRolePayload = {
  __typename?: 'AssignRolePayload';
  error?: Maybe<AssignRoleError>;
  userRole?: Maybe<UserRoleAssignment>;
};

export type AuthVariable = {
  __typename?: 'AuthVariable';
  /** 变量名（如 "tenant_id"） */
  name: Scalars['String']['output'];
  /** JWT 来源路径（如 "jwt.tenant_id"） */
  source: Scalars['String']['output'];
  /** 变量类型 */
  type: AuthVariableType;
};

export type AuthVariableInput = {
  /** 变量名 */
  name: Scalars['String']['input'];
  /** JWT 来源路径 */
  source: Scalars['String']['input'];
  /** 变量类型 */
  type: AuthVariableType;
};

export enum AuthVariableType {
  Integer = 'INTEGER',
  String = 'STRING',
  Uuid = 'UUID'
}

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

export enum ClusterStatus {
  Active = 'ACTIVE',
  Disabled = 'DISABLED'
}

/** 列访问模式 */
export enum ColumnAccessMode {
  /** 完全隐藏 */
  Hidden = 'HIDDEN',
  /** 脱敏显示 */
  Masked = 'MASKED',
  /** 可见但只读 */
  Readonly = 'READONLY',
  /** 可见且可编辑 */
  Visible = 'VISIBLE'
}

export type ColumnPolicy = {
  __typename?: 'ColumnPolicy';
  defaultMode: ColumnAccessMode;
  rules: Array<ColumnRule>;
};

export type ColumnPolicyInput = {
  defaultMode: ColumnAccessMode;
  rules: Array<ColumnRuleInput>;
};

export type ColumnRule = {
  __typename?: 'ColumnRule';
  fieldName: Scalars['String']['output'];
  maskPattern?: Maybe<Scalars['String']['output']>;
  mode: ColumnAccessMode;
};

export type ColumnRuleInput = {
  fieldName: Scalars['String']['input'];
  maskPattern?: InputMaybe<Scalars['String']['input']>;
  mode: ColumnAccessMode;
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

export type CreateEndUserError = ClusterNotFound | EndUserAlreadyExists | EndUserPasswordTooWeak | InvalidInput | ProjectNotFound;

export type CreateEndUserInput = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type CreateEndUserPayload = {
  __typename?: 'CreateEndUserPayload';
  endUser?: Maybe<EndUser>;
  error?: Maybe<CreateEndUserError>;
};

export type CreateEndUserPermissionBundleError = EndUserPermissionBundleAlreadyExists | InvalidInput | ProjectNotFound;

export type CreateEndUserPermissionBundleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
};

export type CreateEndUserPermissionBundlePayload = {
  __typename?: 'CreateEndUserPermissionBundlePayload';
  bundle?: Maybe<EndUserPermissionBundle>;
  error?: Maybe<CreateEndUserPermissionBundleError>;
};

export type CreateEndUserPermissionError = InvalidInput | ModelNotFound | ProjectNotFound | RowScopeFieldMissing;

export type CreateEndUserPermissionInput = {
  action: RbacAction;
  columnPolicy: ColumnPolicyInput;
  description?: InputMaybe<Scalars['String']['input']>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  modelId: Scalars['ID']['input'];
  rowScope: RowScopeType;
};

export type CreateEndUserPermissionPayload = {
  __typename?: 'CreateEndUserPermissionPayload';
  error?: Maybe<CreateEndUserPermissionError>;
  permission?: Maybe<EndUserPermission>;
};

export type CreateEndUserRoleError = EndUserRoleAlreadyExists | InvalidInput | ProjectNotFound;

export type CreateEndUserRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
};

export type CreateEndUserRolePayload = {
  __typename?: 'CreateEndUserRolePayload';
  error?: Maybe<CreateEndUserRoleError>;
  role?: Maybe<EndUserRole>;
};

export type CreateEnumError = EnumAlreadyExists | InvalidInput | ProjectNotFound;

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

export type CreateModelError = InvalidInput | ModelAlreadyExists | ModelTableAlreadyExists | ProjectNotFound;

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

export enum DbTableStatus {
  ClusterUnreachable = 'CLUSTER_UNREACHABLE',
  TableExists = 'TABLE_EXISTS',
  TableMissing = 'TABLE_MISSING'
}

export type DeleteClusterError = ClusterNotFound | ProjectNotFound;

export type DeleteClusterPayload = {
  __typename?: 'DeleteClusterPayload';
  error?: Maybe<DeleteClusterError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEndUserError = ClusterNotFound | EndUserNotFound | ProjectNotFound;

export type DeleteEndUserInput = {
  userId: Scalars['ID']['input'];
};

export type DeleteEndUserPayload = {
  __typename?: 'DeleteEndUserPayload';
  error?: Maybe<DeleteEndUserError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEndUserPermissionBundleError = EndUserPermissionBundleInUse | EndUserPermissionBundleNotFound | ProjectNotFound;

export type DeleteEndUserPermissionBundlePayload = {
  __typename?: 'DeleteEndUserPermissionBundlePayload';
  error?: Maybe<DeleteEndUserPermissionBundleError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEndUserPermissionError = EndUserPermissionInUse | EndUserPermissionNotFound | ProjectNotFound;

export type DeleteEndUserPermissionPayload = {
  __typename?: 'DeleteEndUserPermissionPayload';
  error?: Maybe<DeleteEndUserPermissionError>;
  success: Scalars['Boolean']['output'];
};

export type DeleteEndUserRoleError = EndUserImplicitRoleCannotBeModified | EndUserRoleNotFound | ProjectNotFound;

export type DeleteEndUserRolePayload = {
  __typename?: 'DeleteEndUserRolePayload';
  error?: Maybe<DeleteEndUserRoleError>;
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

export type DeleteLogicalForeignKeyResult = DeleteLogicalForeignKeySuccess | FkNotDeletableError | FkNotFoundError | FkPairHasRelateFieldsError;

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

export type EffectiveGrant = {
  __typename?: 'EffectiveGrant';
  action: RbacAction;
  columnPolicy: ColumnPolicy;
  rowScope: RowScopeType;
};

export type EffectivePermissionSources = {
  __typename?: 'EffectivePermissionSources';
  directBundles: Array<EndUserPermissionBundle>;
  explicitRoleBundles: Array<EndUserRoleBundleSource>;
  implicitRoleBundles: Array<EndUserRoleBundleSource>;
};

export type EffectivePermissions = {
  __typename?: 'EffectivePermissions';
  endUserId: Scalars['ID']['output'];
  grants: Array<EffectiveGrant>;
  modelId: Scalars['ID']['output'];
  sources: EffectivePermissionSources;
};

export type EndUser = Node & {
  __typename?: 'EndUser';
  createdAt: Scalars['Time']['output'];
  createdBy: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  isForbidden: Scalars['Boolean']['output'];
  updatedAt: Scalars['Time']['output'];
  username: Scalars['String']['output'];
};

export type EndUserAlreadyExists = Error & {
  __typename?: 'EndUserAlreadyExists';
  message: Scalars['String']['output'];
};

export type EndUserBundleAssignment = {
  __typename?: 'EndUserBundleAssignment';
  assignedAt: Scalars['Time']['output'];
  bundle: EndUserPermissionBundle;
  endUserId: Scalars['ID']['output'];
};

export type EndUserBundlePermissionEntry = {
  __typename?: 'EndUserBundlePermissionEntry';
  permission: EndUserPermission;
  sortOrder: Scalars['Int']['output'];
};

export type EndUserCannotAssignImplicitRole = Error & {
  __typename?: 'EndUserCannotAssignImplicitRole';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserConnection = {
  __typename?: 'EndUserConnection';
  nodes: Array<EndUser>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EndUserImplicitRoleCannotBeModified = Error & {
  __typename?: 'EndUserImplicitRoleCannotBeModified';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserNotFound = Error & {
  __typename?: 'EndUserNotFound';
  message: Scalars['String']['output'];
};

export type EndUserNotFoundInProject = Error & {
  __typename?: 'EndUserNotFoundInProject';
  message: Scalars['String']['output'];
};

export type EndUserPasswordTooWeak = Error & {
  __typename?: 'EndUserPasswordTooWeak';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserPermission = Node & {
  __typename?: 'EndUserPermission';
  action: RbacAction;
  columnPolicy: ColumnPolicy;
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  displayName?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  modelId: Scalars['ID']['output'];
  /** 来源预设，null 表示手动创建的自定义权限点 */
  preset?: Maybe<EndUserPermissionPreset>;
  rowScope: RowScopeType;
  updatedAt: Scalars['Time']['output'];
};

export type EndUserPermissionBundle = Node & {
  __typename?: 'EndUserPermissionBundle';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  permissions: Array<EndUserBundlePermissionEntry>;
  updatedAt: Scalars['Time']['output'];
};

export type EndUserPermissionBundleAlreadyExists = Error & {
  __typename?: 'EndUserPermissionBundleAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserPermissionBundleConnection = {
  __typename?: 'EndUserPermissionBundleConnection';
  edges: Array<EndUserPermissionBundleEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EndUserPermissionBundleEdge = {
  __typename?: 'EndUserPermissionBundleEdge';
  cursor: Scalars['String']['output'];
  node: EndUserPermissionBundle;
};

export type EndUserPermissionBundleInUse = Error & {
  __typename?: 'EndUserPermissionBundleInUse';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserPermissionBundleNotFound = Error & {
  __typename?: 'EndUserPermissionBundleNotFound';
  message: Scalars['String']['output'];
};

export type EndUserPermissionConnection = {
  __typename?: 'EndUserPermissionConnection';
  edges: Array<EndUserPermissionEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EndUserPermissionEdge = {
  __typename?: 'EndUserPermissionEdge';
  cursor: Scalars['String']['output'];
  node: EndUserPermission;
};

export type EndUserPermissionInUse = Error & {
  __typename?: 'EndUserPermissionInUse';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserPermissionNotFound = Error & {
  __typename?: 'EndUserPermissionNotFound';
  message: Scalars['String']['output'];
};

/**
 * 权限点预设策略类型。
 *
 * 应用预设时会替换该模型下所有 preset != null 的权限点，
 * preset = null（手动创建）的权限点不受影响。
 *
 * READ_WRITE_ALL      — 读写全部（不依赖 END_USER_REF 字段）
 *                       SELECT ALL + INSERT ALL + UPDATE ALL + DELETE ALL
 * READ_ALL            — 只读全部（不依赖 END_USER_REF 字段）
 *                       SELECT ALL
 * READ_WRITE_OWNER    — 读写自己（依赖 END_USER_REF 字段）
 *                       SELECT SELF + INSERT SELF + UPDATE SELF + DELETE SELF
 * READ_ALL_WRITE_OWNER — 读所有写自己（依赖 END_USER_REF 字段）
 *                       SELECT ALL + INSERT SELF + UPDATE SELF + DELETE SELF
 */
export enum EndUserPermissionPreset {
  ReadAll = 'READ_ALL',
  ReadAllWriteOwner = 'READ_ALL_WRITE_OWNER',
  ReadWriteAll = 'READ_WRITE_ALL',
  ReadWriteOwner = 'READ_WRITE_OWNER'
}

export type EndUserProjectAccess = Node & {
  __typename?: 'EndUserProjectAccess';
  endUser: EndUser;
  grantedAt: Scalars['Time']['output'];
  grantedBy: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  permissionBundleId: Scalars['ID']['output'];
  permissionBundleName: Scalars['String']['output'];
};

export type EndUserProjectAccessAlreadyExists = Error & {
  __typename?: 'EndUserProjectAccessAlreadyExists';
  message: Scalars['String']['output'];
};

export type EndUserProjectAccessConnection = {
  __typename?: 'EndUserProjectAccessConnection';
  nodes: Array<EndUserProjectAccess>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EndUserProjectAccessNotFound = Error & {
  __typename?: 'EndUserProjectAccessNotFound';
  message: Scalars['String']['output'];
};

export type EndUserRefAlreadyExists = Error & {
  __typename?: 'EndUserRefAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserRole = Node & {
  __typename?: 'EndUserRole';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  isImplicit: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  permissionBundles: Array<EndUserRoleBundleEntry>;
  updatedAt: Scalars['Time']['output'];
};

export type EndUserRoleAlreadyExists = Error & {
  __typename?: 'EndUserRoleAlreadyExists';
  message: Scalars['String']['output'];
  suggestion?: Maybe<Scalars['String']['output']>;
};

export type EndUserRoleAssignment = {
  __typename?: 'EndUserRoleAssignment';
  assignedAt: Scalars['Time']['output'];
  endUserId: Scalars['ID']['output'];
  role: EndUserRole;
};

export type EndUserRoleBundleEntry = {
  __typename?: 'EndUserRoleBundleEntry';
  assignedAt: Scalars['Time']['output'];
  bundle: EndUserPermissionBundle;
};

export type EndUserRoleBundleSource = {
  __typename?: 'EndUserRoleBundleSource';
  bundles: Array<EndUserPermissionBundle>;
  role: EndUserRole;
};

export type EndUserRoleConnection = {
  __typename?: 'EndUserRoleConnection';
  edges: Array<EndUserRoleEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type EndUserRoleEdge = {
  __typename?: 'EndUserRoleEdge';
  cursor: Scalars['String']['output'];
  node: EndUserRole;
};

export type EndUserRoleNotFound = Error & {
  __typename?: 'EndUserRoleNotFound';
  message: Scalars['String']['output'];
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

export enum FkCreateMode {
  Bidirectional = 'BIDIRECTIONAL',
  Unidirectional = 'UNIDIRECTIONAL'
}

export enum FkDirection {
  Normal = 'NORMAL',
  Reverse = 'REVERSE'
}

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

export enum FieldConflictAspect {
  NotNullMismatch = 'NOT_NULL_MISMATCH',
  PrimaryMismatch = 'PRIMARY_MISMATCH',
  UniqueMismatch = 'UNIQUE_MISMATCH'
}

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

export enum FormatType {
  Boolean = 'BOOLEAN',
  Date = 'DATE',
  Datetime = 'DATETIME',
  Decimal = 'DECIMAL',
  EndUserRef = 'END_USER_REF',
  Enum = 'ENUM',
  Integer = 'INTEGER',
  Number = 'NUMBER',
  Relation = 'RELATION',
  String = 'STRING',
  Time = 'TIME',
  Uuid = 'UUID'
}

export type GetClusterError = ClusterNotFound | ProjectNotFound;

export type GetClusterPayload = {
  __typename?: 'GetClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<GetClusterError>;
};

export type GetEffectivePermissionsError = EndUserNotFoundInProject | ModelNotFound | ProjectNotFound;

export type GetEffectivePermissionsInput = {
  endUserId: Scalars['ID']['input'];
  modelId: Scalars['ID']['input'];
};

export type GetEffectivePermissionsPayload = {
  __typename?: 'GetEffectivePermissionsPayload';
  effectivePermissions?: Maybe<EffectivePermissions>;
  error?: Maybe<GetEffectivePermissionsError>;
};

export type GetEnumError = EnumNotFound | ProjectNotFound;

export type GetEnumPayload = {
  __typename?: 'GetEnumPayload';
  enum?: Maybe<EnumDefinition>;
  error?: Maybe<GetEnumError>;
};

export type GetModelDatabaseCatalogPayload = {
  __typename?: 'GetModelDatabaseCatalogPayload';
  data?: Maybe<ModelDatabaseCatalogPayload>;
  error?: Maybe<ModelDatabaseCatalogError>;
};

export type GetModelError = InvalidInput | ModelNotFound | ProjectNotFound;

export type GetModelPayload = {
  __typename?: 'GetModelPayload';
  error?: Maybe<GetModelError>;
  model?: Maybe<Model>;
};

export type GetMyUserProfileError = ProfileNotFound | UserNotFound;

export type GetMyUserProfilePayload = {
  __typename?: 'GetMyUserProfilePayload';
  error?: Maybe<GetMyUserProfileError>;
  user?: Maybe<User>;
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

export type GrantEndUserProjectAccessError = EndUserNotFound | EndUserPermissionBundleNotFound | EndUserProjectAccessAlreadyExists | InvalidInput | ProjectNotFound;

export type GrantEndUserProjectAccessInput = {
  endUserId: Scalars['ID']['input'];
  permissionBundleId: Scalars['ID']['input'];
};

export type GrantEndUserProjectAccessPayload = {
  __typename?: 'GrantEndUserProjectAccessPayload';
  access?: Maybe<EndUserProjectAccess>;
  error?: Maybe<GrantEndUserProjectAccessError>;
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

export enum HealthStatus {
  Broken = 'BROKEN',
  Healthy = 'HEALTHY',
  NeedsRepair = 'NEEDS_REPAIR'
}

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

export type InitPrivateDbError = Error & {
  __typename?: 'InitPrivateDBError';
  message: Scalars['String']['output'];
};

export type InitPrivateDbPayload = {
  __typename?: 'InitPrivateDBPayload';
  error?: Maybe<InitPrivateDbPayloadError>;
  success: Scalars['Boolean']['output'];
};

export type InitPrivateDbPayloadError = InitPrivateDbError | ProjectNotFound;

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

export type ListEndUserPermissionBundlesInput = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListEndUserPermissionsInput = {
  action?: InputMaybe<RbacAction>;
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  modelId?: InputMaybe<Scalars['ID']['input']>;
};

export type ListEndUserRolesInput = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  includeImplicit?: InputMaybe<Scalars['Boolean']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListEndUsersError = InvalidInput;

export type ListEndUsersInput = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListEndUsersPayload = {
  __typename?: 'ListEndUsersPayload';
  connection?: Maybe<EndUserConnection>;
  error?: Maybe<ListEndUsersError>;
};

export type ListProjectEndUserAccessError = InvalidInput | ProjectNotFound;

export type ListProjectEndUserAccessInput = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListProjectEndUserAccessPayload = {
  __typename?: 'ListProjectEndUserAccessPayload';
  connection?: Maybe<EndUserProjectAccessConnection>;
  error?: Maybe<ListProjectEndUserAccessError>;
};

export type ListProjectEndUsersError = ClusterNotFound | ProjectNotFound;

export type ListProjectEndUsersInput = {
  after?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ListProjectEndUsersPayload = {
  __typename?: 'ListProjectEndUsersPayload';
  connection?: Maybe<EndUserConnection>;
  error?: Maybe<ListProjectEndUsersError>;
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

export enum MembershipStatus {
  Active = 'ACTIVE',
  Invited = 'INVITED',
  Suspended = 'SUSPENDED'
}

export type Model = Node & {
  __typename?: 'Model';
  createdAt: Scalars['String']['output'];
  databaseName: Scalars['String']['output'];
  dbTable?: Maybe<DbTableStatus>;
  description: Scalars['String']['output'];
  displayField?: Maybe<Scalars['String']['output']>;
  fields: Array<Field>;
  group: ModelGroup;
  id: Scalars['ID']['output'];
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

export type ModelConnection = {
  __typename?: 'ModelConnection';
  edges: Array<ModelEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ModelDatabaseCatalogError = InvalidInput | ProjectNotFound;

export type ModelDatabaseCatalogInput = {
  page?: InputMaybe<Scalars['Int']['input']>;
  pageSize?: InputMaybe<Scalars['Int']['input']>;
  search?: InputMaybe<Scalars['String']['input']>;
};

export type ModelDatabaseCatalogPayload = {
  __typename?: 'ModelDatabaseCatalogPayload';
  databases: Array<DatabaseLite>;
  page: Scalars['Int']['output'];
  pageSize: Scalars['Int']['output'];
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
  addEndUserPermissionToBundle: AddEndUserPermissionToBundlePayload;
  addFields: AddFieldsPayload;
  addPermissionToRole: AddRolePermissionPayload;
  /**
   * 对指定模型应用预设策略。
   * 会替换该模型下所有 preset != null 的权限点，preset = null 的自定义权限点保留不变。
   */
  applyEndUserPresetPolicy: ApplyEndUserPresetPolicyPayload;
  assignBundleToEndUser: AssignBundleToEndUserPayload;
  assignBundleToEndUserRole: AssignBundleToEndUserRolePayload;
  assignEndUserRole: AssignEndUserRolePayload;
  assignRoleToUser: AssignRolePayload;
  createCustomRole: CreateCustomRolePayload;
  createEndUser: CreateEndUserPayload;
  createEndUserPermission: CreateEndUserPermissionPayload;
  createEndUserPermissionBundle: CreateEndUserPermissionBundlePayload;
  createEndUserRole: CreateEndUserRolePayload;
  createEnum: CreateEnumPayload;
  createGroup: CreateGroupPayload;
  createLogicalForeignKey: CreateLogicalForeignKeyPayload;
  createModel: CreateModelPayload;
  createModelFromSchema: CreateModelFromSchemaPayload;
  createProject: CreateProjectPayload;
  createRole: CreateRolePayload;
  deleteEndUser: DeleteEndUserPayload;
  deleteEndUserPermission: DeleteEndUserPermissionPayload;
  deleteEndUserPermissionBundle: DeleteEndUserPermissionBundlePayload;
  deleteEndUserRole: DeleteEndUserRolePayload;
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
  grantEndUserProjectAccess: GrantEndUserProjectAccessPayload;
  importModel: ImportModelPayload;
  moveModelToGroup: MoveModelToGroupPayload;
  pong: Scalars['String']['output'];
  removeEndUserPermissionFromBundle: RemoveEndUserPermissionFromBundlePayload;
  removeField: RemoveFieldPayload;
  removePermissionFromRole: RemoveRolePermissionPayload;
  renameGroup: RenameGroupPayload;
  reorderGroup: ReorderGroupPayload;
  repairModel: RepairModelPayload;
  revokeBundleFromEndUser: RevokeBundleFromEndUserPayload;
  revokeBundleFromEndUserRole: RevokeBundleFromEndUserRolePayload;
  revokeEndUserProjectAccess: RevokeEndUserProjectAccessPayload;
  revokeEndUserRole: RevokeEndUserRolePayload;
  revokeRoleFromUser: RevokeRolePayload;
  /**
   * 设置 Model RLS 策略
   * 支持完整的五件套 JSON 表达式，不限于 Preset
   */
  setModelRLSPolicy: SetModelRlsPolicyPayload;
  /** 设置当前项目 auth_schema */
  setProjectAuthSchema: SetProjectAuthSchemaPayload;
  syncModelSchema: SyncModelSchemaPayload;
  testDatabaseConnection: TestConnectionPayload;
  /**
   * 解除字段的废弃状态，恢复为正常可用。
   * 若字段未处于废弃状态，幂等返回成功。
   * 状态转换：DEPRECATED → ACTIVE
   */
  undeprecateField?: Maybe<Model>;
  updateEndUserPermission: UpdateEndUserPermissionPayload;
  updateEndUserPermissionBundle: UpdateEndUserPermissionBundlePayload;
  updateEndUserProjectAccess: UpdateEndUserProjectAccessPayload;
  updateEndUserRole: UpdateEndUserRolePayload;
  updateEndUserStatus: UpdateEndUserStatusPayload;
  updateEnum: UpdateEnumPayload;
  updateField: UpdateFieldPayload;
  updateModelMeta: UpdateModelMetaPayload;
  updateMyProfile: UpdateMyProfilePayload;
  updateOrganization: UpdateOrganizationPayload;
  updatePermissionRole: UpdatePermissionRolePayload;
  updateProject: UpdateProjectPayload;
  updateProjectCluster: UpdateClusterPayload;
  /**
   * 校验 RLS 表达式合法性
   * 用于 Policy 配置页面的实时校验
   */
  validateRLSExpr: ValidateRlsExprPayload;
};


export type MutationAddEndUserPermissionToBundleArgs = {
  input: AddEndUserPermissionToBundleInput;
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


export type MutationApplyEndUserPresetPolicyArgs = {
  input: ApplyEndUserPresetPolicyInput;
};


export type MutationAssignBundleToEndUserArgs = {
  input: AssignBundleToEndUserInput;
};


export type MutationAssignBundleToEndUserRoleArgs = {
  input: AssignBundleToEndUserRoleInput;
};


export type MutationAssignEndUserRoleArgs = {
  input: AssignEndUserRoleInput;
};


export type MutationAssignRoleToUserArgs = {
  orgName: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
  userId: Scalars['String']['input'];
};


export type MutationCreateCustomRoleArgs = {
  input: CreateCustomRoleInput;
};


export type MutationCreateEndUserArgs = {
  input: CreateEndUserInput;
};


export type MutationCreateEndUserPermissionArgs = {
  input: CreateEndUserPermissionInput;
};


export type MutationCreateEndUserPermissionBundleArgs = {
  input: CreateEndUserPermissionBundleInput;
};


export type MutationCreateEndUserRoleArgs = {
  input: CreateEndUserRoleInput;
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


export type MutationDeleteEndUserArgs = {
  input: DeleteEndUserInput;
};


export type MutationDeleteEndUserPermissionArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteEndUserPermissionBundleArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteEndUserRoleArgs = {
  id: Scalars['ID']['input'];
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


export type MutationGrantEndUserProjectAccessArgs = {
  input: GrantEndUserProjectAccessInput;
};


export type MutationImportModelArgs = {
  input: ImportModelInput;
};


export type MutationMoveModelToGroupArgs = {
  input: MoveModelToGroupInput;
};


export type MutationRemoveEndUserPermissionFromBundleArgs = {
  input: RemoveEndUserPermissionFromBundleInput;
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


export type MutationRevokeBundleFromEndUserArgs = {
  input: RevokeBundleFromEndUserInput;
};


export type MutationRevokeBundleFromEndUserRoleArgs = {
  input: RevokeBundleFromEndUserRoleInput;
};


export type MutationRevokeEndUserProjectAccessArgs = {
  input: RevokeEndUserProjectAccessInput;
};


export type MutationRevokeEndUserRoleArgs = {
  input: RevokeEndUserRoleInput;
};


export type MutationRevokeRoleFromUserArgs = {
  orgName: Scalars['String']['input'];
  roleId: Scalars['Int']['input'];
  userId: Scalars['String']['input'];
};


export type MutationSetModelRlsPolicyArgs = {
  input: SetModelRlsPolicyInput;
};


export type MutationSetProjectAuthSchemaArgs = {
  input: SetProjectAuthSchemaInput;
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


export type MutationUpdateEndUserPermissionArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEndUserPermissionInput;
};


export type MutationUpdateEndUserPermissionBundleArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEndUserPermissionBundleInput;
};


export type MutationUpdateEndUserProjectAccessArgs = {
  input: UpdateEndUserProjectAccessInput;
};


export type MutationUpdateEndUserRoleArgs = {
  id: Scalars['ID']['input'];
  input: UpdateEndUserRoleInput;
};


export type MutationUpdateEndUserStatusArgs = {
  input: UpdateEndUserStatusInput;
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

export type OrganizationNotFound = Error & {
  __typename?: 'OrganizationNotFound';
  message: Scalars['String']['output'];
};

export enum OrganizationStatus {
  Active = 'ACTIVE',
  Deleted = 'DELETED',
  Suspended = 'SUSPENDED'
}

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

export type PresetRequiresOwnerField = Error & {
  __typename?: 'PresetRequiresOwnerField';
  /** 选择的预设依赖 END_USER_REF 字段，但模型中不存在该字段 */
  message: Scalars['String']['output'];
  preset: EndUserPermissionPreset;
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

export type ProfileNotFound = Error & {
  __typename?: 'ProfileNotFound';
  message: Scalars['String']['output'];
};

export type Project = Node & {
  __typename?: 'Project';
  /** 认证变量配置（用于 RLS 表达式中的 _auth 引用） */
  authSchema: ProjectAuthSchema;
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

export type ProjectAuthSchema = {
  __typename?: 'ProjectAuthSchema';
  /** 认证变量列表（不含内置 uid） */
  variables: Array<AuthVariable>;
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

export enum ProjectStatus {
  Active = 'ACTIVE',
  Archived = 'ARCHIVED'
}

export type Query = {
  __typename?: 'Query';
  databaseCluster: GetClusterPayload;
  effectivePermissions: GetEffectivePermissionsPayload;
  endUserBundleAssignments: Array<EndUserBundleAssignment>;
  endUserPermission?: Maybe<EndUserPermission>;
  endUserPermissionBundle?: Maybe<EndUserPermissionBundle>;
  endUserPermissionBundles: EndUserPermissionBundleConnection;
  endUserPermissions: EndUserPermissionConnection;
  endUserRole?: Maybe<EndUserRole>;
  endUserRoleAssignments: Array<EndUserRoleAssignment>;
  endUserRoles: EndUserRoleConnection;
  enum: GetEnumPayload;
  enumReferences: Array<Scalars['String']['output']>;
  enums: Array<EnumDefinition>;
  fields: Array<Field>;
  hello: Scalars['String']['output'];
  listDatabases: DatabaseConnection;
  listEndUsers: ListEndUsersPayload;
  listProjectEndUserAccess: ListProjectEndUserAccessPayload;
  listProjectEndUsers: ListProjectEndUsersPayload;
  listTables: TableListConnection;
  logicalForeignKeys: Array<LogicalForeignKey>;
  me: CurrentUser;
  model: GetModelPayload;
  modelByName: GetModelPayload;
  modelDatabaseCatalog: GetModelDatabaseCatalogPayload;
  modelGroups: Array<ModelGroup>;
  modelJsonSchema?: Maybe<ModelJsonSchema>;
  /** 获取 Model RLS 策略配置 */
  modelRLSPolicy?: Maybe<ModelRlsPolicy>;
  models: ModelConnection;
  myOrganizations: Array<Organization>;
  myUserProfile: GetMyUserProfilePayload;
  node?: Maybe<Node>;
  organizationMembers: Array<OrganizationMember>;
  permissionRole?: Maybe<PermissionRole>;
  permissionRoles: Array<PermissionRole>;
  ping: Scalars['String']['output'];
  project: GetProjectPayload;
  /** 获取当前项目 auth_schema */
  projectAuthSchema: ProjectAuthSchema;
  projects: Array<Project>;
  rolePermissionsList: Array<PermissionDef>;
  roles: Array<Role>;
  userRoleAssignments: Array<UserRoleAssignment>;
};


export type QueryDatabaseClusterArgs = {
  projectSlug: Scalars['String']['input'];
};


export type QueryEffectivePermissionsArgs = {
  input: GetEffectivePermissionsInput;
};


export type QueryEndUserBundleAssignmentsArgs = {
  endUserId: Scalars['ID']['input'];
};


export type QueryEndUserPermissionArgs = {
  id: Scalars['ID']['input'];
};


export type QueryEndUserPermissionBundleArgs = {
  id: Scalars['ID']['input'];
};


export type QueryEndUserPermissionBundlesArgs = {
  input?: InputMaybe<ListEndUserPermissionBundlesInput>;
};


export type QueryEndUserPermissionsArgs = {
  input?: InputMaybe<ListEndUserPermissionsInput>;
};


export type QueryEndUserRoleArgs = {
  id: Scalars['ID']['input'];
};


export type QueryEndUserRoleAssignmentsArgs = {
  endUserId: Scalars['ID']['input'];
};


export type QueryEndUserRolesArgs = {
  input?: InputMaybe<ListEndUserRolesInput>;
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


export type QueryListEndUsersArgs = {
  input?: InputMaybe<ListEndUsersInput>;
};


export type QueryListProjectEndUserAccessArgs = {
  input?: InputMaybe<ListProjectEndUserAccessInput>;
};


export type QueryListProjectEndUsersArgs = {
  input?: InputMaybe<ListProjectEndUsersInput>;
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


export type QueryModelDatabaseCatalogArgs = {
  input?: InputMaybe<ModelDatabaseCatalogInput>;
};


export type QueryModelJsonSchemaArgs = {
  id: Scalars['ID']['input'];
};


export type QueryModelRlsPolicyArgs = {
  modelId: Scalars['ID']['input'];
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

export type RlsCheckViolation = Error & {
  __typename?: 'RLSCheckViolation';
  message: Scalars['String']['output'];
  operation?: Maybe<Scalars['String']['output']>;
};

export enum RlsExprType {
  DeletePredicate = 'DELETE_PREDICATE',
  InsertCheck = 'INSERT_CHECK',
  SelectPredicate = 'SELECT_PREDICATE',
  UpdateCheck = 'UPDATE_CHECK',
  UpdatePredicate = 'UPDATE_PREDICATE'
}

export enum RlsPreset {
  /**
   * 无访问权限
   * 五件套均为 false
   */
  NoAccess = 'NO_ACCESS',
  /**
   * 只读全部
   * selectPredicate=true，其余为 false
   */
  ReadAll = 'READ_ALL',
  /**
   * 读取全部，写自己
   * selectPredicate=true，其余为 OWNER_EQUALS_USER
   */
  ReadAllWriteOwner = 'READ_ALL_WRITE_OWNER',
  /**
   * 读写全部（⚠️ 高危策略）
   * 五件套均为 true
   */
  ReadWriteAll = 'READ_WRITE_ALL',
  /**
   * 默认策略：读写自己
   * 五件套均为 {"owner":{"_eq":{"_auth":"uid"}}}
   */
  ReadWriteOwner = 'READ_WRITE_OWNER'
}

/** 数据操作动作：终端用户对数据表可执行的操作类型 */
export enum RbacAction {
  Delete = 'DELETE',
  Export = 'EXPORT',
  Insert = 'INSERT',
  Select = 'SELECT',
  Update = 'UPDATE'
}

export type RemoveEndUserPermissionFromBundleError = EndUserPermissionBundleNotFound | EndUserPermissionNotFound | ProjectNotFound;

export type RemoveEndUserPermissionFromBundleInput = {
  bundleId: Scalars['ID']['input'];
  permissionId: Scalars['ID']['input'];
};

export type RemoveEndUserPermissionFromBundlePayload = {
  __typename?: 'RemoveEndUserPermissionFromBundlePayload';
  bundle?: Maybe<EndUserPermissionBundle>;
  error?: Maybe<RemoveEndUserPermissionFromBundleError>;
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

export enum RepairMode {
  Additive = 'ADDITIVE',
  DryRun = 'DRY_RUN',
  FullSync = 'FULL_SYNC'
}

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

export type RevokeBundleFromEndUserError = EndUserNotFoundInProject | EndUserPermissionBundleNotFound | ProjectNotFound;

export type RevokeBundleFromEndUserInput = {
  bundleId: Scalars['ID']['input'];
  endUserId: Scalars['ID']['input'];
};

export type RevokeBundleFromEndUserPayload = {
  __typename?: 'RevokeBundleFromEndUserPayload';
  error?: Maybe<RevokeBundleFromEndUserError>;
  success: Scalars['Boolean']['output'];
};

export type RevokeBundleFromEndUserRoleError = EndUserPermissionBundleNotFound | EndUserRoleNotFound | ProjectNotFound;

export type RevokeBundleFromEndUserRoleInput = {
  bundleId: Scalars['ID']['input'];
  roleId: Scalars['ID']['input'];
};

export type RevokeBundleFromEndUserRolePayload = {
  __typename?: 'RevokeBundleFromEndUserRolePayload';
  error?: Maybe<RevokeBundleFromEndUserRoleError>;
  role?: Maybe<EndUserRole>;
};

export type RevokeEndUserProjectAccessError = EndUserProjectAccessNotFound | ProjectNotFound;

export type RevokeEndUserProjectAccessInput = {
  accessId: Scalars['ID']['input'];
};

export type RevokeEndUserProjectAccessPayload = {
  __typename?: 'RevokeEndUserProjectAccessPayload';
  error?: Maybe<RevokeEndUserProjectAccessError>;
  success: Scalars['Boolean']['output'];
};

export type RevokeEndUserRoleError = EndUserNotFoundInProject | EndUserRoleNotFound | ProjectNotFound;

export type RevokeEndUserRoleInput = {
  endUserId: Scalars['ID']['input'];
  roleId: Scalars['ID']['input'];
};

export type RevokeEndUserRolePayload = {
  __typename?: 'RevokeEndUserRolePayload';
  error?: Maybe<RevokeEndUserRoleError>;
  success: Scalars['Boolean']['output'];
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

export type RolePermissionError = InvalidInput | PermissionRoleNotFound | PermissionSystemRoleCannotBeModified;

export type RowScopeFieldMissing = Error & {
  __typename?: 'RowScopeFieldMissing';
  message: Scalars['String']['output'];
  missingField: Scalars['String']['output'];
  requiredByRowScope: RowScopeType;
  suggestion?: Maybe<Scalars['String']['output']>;
};

/**
 * 行策略范围：控制终端用户可见哪些数据行
 *
 * ALL                — 全部行，不过滤
 * SELF               — 仅当前用户自己的行
 * DEPT               — 仅当前用户所在部门的行
 * DEPT_AND_CHILDREN  — 当前部门及所有下级部门的行
 */
export enum RowScopeType {
  All = 'ALL',
  Dept = 'DEPT',
  DeptAndChildren = 'DEPT_AND_CHILDREN',
  Self = 'SELF'
}

export type SchemaIssue = {
  __typename?: 'SchemaIssue';
  description: Scalars['String']['output'];
  details?: Maybe<Scalars['String']['output']>;
  fieldName?: Maybe<Scalars['String']['output']>;
  tableName: Scalars['String']['output'];
  type: SchemaIssueType;
};

export enum SchemaIssueType {
  ClusterNotFound = 'CLUSTER_NOT_FOUND',
  DatabaseMissing = 'DATABASE_MISSING',
  FieldConstraintMismatch = 'FIELD_CONSTRAINT_MISMATCH',
  FieldHasDependencies = 'FIELD_HAS_DEPENDENCIES',
  FieldMissing = 'FIELD_MISSING',
  FieldTypeMismatch = 'FIELD_TYPE_MISMATCH',
  TableMissing = 'TABLE_MISSING'
}

export enum SchemaType {
  Array = 'ARRAY',
  Boolean = 'BOOLEAN',
  Number = 'NUMBER',
  Object = 'OBJECT',
  String = 'STRING'
}

export type SetModelRlsPolicyError = InvalidAuthVariable | InvalidRlsExpression | ModelHasNoOwnerField | ModelNotFound | ProjectNotFound;

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

export type SetProjectAuthSchemaError = InvalidInput | ProjectNotFound;

export type SetProjectAuthSchemaInput = {
  /** 项目 slug */
  projectSlug: Scalars['String']['input'];
  /** 认证变量列表（uid 内置，无需声明） */
  variables: Array<AuthVariableInput>;
};

export type SetProjectAuthSchemaPayload = {
  __typename?: 'SetProjectAuthSchemaPayload';
  authSchema?: Maybe<ProjectAuthSchema>;
  error?: Maybe<SetProjectAuthSchemaError>;
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

export type UpdateClusterConnectionInput = {
  connectionInfo?: InputMaybe<DatabaseConnectionInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  skipConnectionTest?: InputMaybe<Scalars['Boolean']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateClusterError = ClusterNotFound | DatabaseConnectionFailed | InvalidInput | ProjectNotFound;

export type UpdateClusterPayload = {
  __typename?: 'UpdateClusterPayload';
  cluster?: Maybe<DatabaseCluster>;
  error?: Maybe<UpdateClusterError>;
};

export type UpdateEndUserError = ClusterNotFound | EndUserNotFound | InvalidInput | ProjectNotFound;

export type UpdateEndUserPermissionBundleError = EndUserPermissionBundleAlreadyExists | EndUserPermissionBundleNotFound | InvalidInput | ProjectNotFound;

export type UpdateEndUserPermissionBundleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateEndUserPermissionBundlePayload = {
  __typename?: 'UpdateEndUserPermissionBundlePayload';
  bundle?: Maybe<EndUserPermissionBundle>;
  error?: Maybe<UpdateEndUserPermissionBundleError>;
};

export type UpdateEndUserPermissionError = EndUserPermissionNotFound | InvalidInput | ProjectNotFound | RowScopeFieldMissing;

export type UpdateEndUserPermissionInput = {
  columnPolicy?: InputMaybe<ColumnPolicyInput>;
  description?: InputMaybe<Scalars['String']['input']>;
  displayName?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateEndUserPermissionPayload = {
  __typename?: 'UpdateEndUserPermissionPayload';
  error?: Maybe<UpdateEndUserPermissionError>;
  permission?: Maybe<EndUserPermission>;
};

export type UpdateEndUserProjectAccessError = EndUserPermissionBundleNotFound | EndUserProjectAccessNotFound | InvalidInput | ProjectNotFound;

export type UpdateEndUserProjectAccessInput = {
  accessId: Scalars['ID']['input'];
  permissionBundleId: Scalars['ID']['input'];
};

export type UpdateEndUserProjectAccessPayload = {
  __typename?: 'UpdateEndUserProjectAccessPayload';
  access?: Maybe<EndUserProjectAccess>;
  error?: Maybe<UpdateEndUserProjectAccessError>;
};

export type UpdateEndUserRoleError = EndUserImplicitRoleCannotBeModified | EndUserRoleAlreadyExists | EndUserRoleNotFound | InvalidInput | ProjectNotFound;

export type UpdateEndUserRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateEndUserRolePayload = {
  __typename?: 'UpdateEndUserRolePayload';
  error?: Maybe<UpdateEndUserRoleError>;
  role?: Maybe<EndUserRole>;
};

export type UpdateEndUserStatusInput = {
  isForbidden: Scalars['Boolean']['input'];
  userId: Scalars['ID']['input'];
};

export type UpdateEndUserStatusPayload = {
  __typename?: 'UpdateEndUserStatusPayload';
  endUser?: Maybe<EndUser>;
  error?: Maybe<UpdateEndUserError>;
};

export type UpdateEnumError = EnumNotFound | InvalidInput | ProjectNotFound;

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

export type UpdateModelError = InvalidInput | ModelNotFound | ProjectNotFound;

export type UpdateModelMetaInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  displayField?: InputMaybe<Scalars['String']['input']>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateModelMetaPayload = {
  __typename?: 'UpdateModelMetaPayload';
  error?: Maybe<UpdateModelError>;
  model?: Maybe<Model>;
  success: Scalars['Boolean']['output'];
};

export type UpdateMyProfileError = InvalidInput | ProfileNotFound;

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

export type UpdatePermissionRoleError = InvalidInput | PermissionRoleAlreadyExists | PermissionRoleNotFound | PermissionSystemRoleCannotBeModified;

export type UpdatePermissionRolePayload = {
  __typename?: 'UpdatePermissionRolePayload';
  error?: Maybe<UpdatePermissionRoleError>;
  role?: Maybe<PermissionRole>;
};

export type UpdateProjectError = InvalidInput | ProjectNotFound;

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

export type UserBundleAlreadyAssigned = Error & {
  __typename?: 'UserBundleAlreadyAssigned';
  message: Scalars['String']['output'];
};

export type UserNotFound = Error & {
  __typename?: 'UserNotFound';
  message: Scalars['String']['output'];
};

export type UserRoleAlreadyAssigned = Error & {
  __typename?: 'UserRoleAlreadyAssigned';
  message: Scalars['String']['output'];
};

export type UserRoleAssignment = {
  __typename?: 'UserRoleAssignment';
  createdAt: Scalars['Time']['output'];
  id: Scalars['Int']['output'];
  orgName: Scalars['String']['output'];
  roleId: Scalars['Int']['output'];
  userId: Scalars['String']['output'];
};

export enum UserStatus {
  Active = 'ACTIVE',
  Registered = 'REGISTERED',
  Suspended = 'SUSPENDED'
}

export type ValidateRlsExprError = InvalidAuthVariable | InvalidRlsExpression | ModelNotFound | ProjectNotFound;

export type ValidateRlsExprInput = {
  /** 要校验的谓词类型 */
  exprType: RlsExprType;
  /** 表达式 JSON 字符串 */
  expression: Scalars['String']['input'];
  /** 所属模型 ID（用于字段名白名单校验） */
  modelId: Scalars['ID']['input'];
};

export type ValidateRlsExprPayload = {
  __typename?: 'ValidateRLSExprPayload';
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


export type GetClusterQuery = { __typename?: 'Query', databaseCluster: { __typename?: 'GetClusterPayload', cluster?: { __typename?: 'DatabaseCluster', id: string, title: string, description: string, status: ClusterStatus, createdAt: string, updatedAt: string, connectionInfo: { __typename?: 'DatabaseConnectionInfo', host: string, port: any, username: string, password: string } } | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type ListDatabasesQueryVariables = Exact<{
  input: ListDatabasesInput;
}>;


export type ListDatabasesQuery = { __typename?: 'Query', listDatabases: { __typename?: 'DatabaseConnection', totalCount: number, edges: Array<{ __typename?: 'DatabaseEdge', node: { __typename?: 'Database', name: string } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } };

export type ModelDatabaseCatalogQueryVariables = Exact<{
  input?: InputMaybe<ModelDatabaseCatalogInput>;
}>;


export type ModelDatabaseCatalogQuery = { __typename?: 'Query', modelDatabaseCatalog: { __typename?: 'GetModelDatabaseCatalogPayload', data?: { __typename?: 'ModelDatabaseCatalogPayload', totalCount: number, page: number, pageSize: number, databases: Array<{ __typename?: 'DatabaseLite', name: string }> } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type TestClusterConnectionMutationVariables = Exact<{
  input: TestDatabaseConnectionInput;
}>;


export type TestClusterConnectionMutation = { __typename?: 'Mutation', testDatabaseConnection: { __typename?: 'TestConnectionPayload', success: boolean, connectionTime?: number | null, error?:
      | { __typename: 'ClusterNotFound', message: string }
      | { __typename: 'DatabaseConnectionFailed', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type ListEndUsersQueryVariables = Exact<{
  input?: InputMaybe<ListEndUsersInput>;
}>;


export type ListEndUsersQuery = { __typename?: 'Query', listEndUsers: { __typename?: 'ListEndUsersPayload', connection?: { __typename?: 'EndUserConnection', totalCount: number, nodes: Array<{ __typename?: 'EndUser', id: string, username: string, isForbidden: boolean, createdBy: string, createdAt: any, updatedAt: any }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } | null, error?: { __typename: 'InvalidInput', message: string, suggestion?: string | null } | null } };

export type CreateEndUserMutationVariables = Exact<{
  input: CreateEndUserInput;
}>;


export type CreateEndUserMutation = { __typename?: 'Mutation', createEndUser: { __typename?: 'CreateEndUserPayload', endUser?: { __typename?: 'EndUser', id: string, username: string, isForbidden: boolean, createdBy: string, createdAt: any, updatedAt: any } | null, error?:
      | { __typename: 'ClusterNotFound' }
      | { __typename: 'EndUserAlreadyExists', message: string }
      | { __typename: 'EndUserPasswordTooWeak', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound' }
     | null } };

export type UpdateEndUserStatusMutationVariables = Exact<{
  input: UpdateEndUserStatusInput;
}>;


export type UpdateEndUserStatusMutation = { __typename?: 'Mutation', updateEndUserStatus: { __typename?: 'UpdateEndUserStatusPayload', endUser?: { __typename?: 'EndUser', id: string, username: string, isForbidden: boolean, updatedAt: any } | null, error?:
      | { __typename: 'ClusterNotFound' }
      | { __typename: 'EndUserNotFound', message: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound' }
     | null } };

export type DeleteEndUserMutationVariables = Exact<{
  input: DeleteEndUserInput;
}>;


export type DeleteEndUserMutation = { __typename?: 'Mutation', deleteEndUser: { __typename?: 'DeleteEndUserPayload', success: boolean, error?:
      | { __typename: 'ClusterNotFound' }
      | { __typename: 'EndUserNotFound', message: string }
      | { __typename: 'ProjectNotFound' }
     | null } };

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

export type CreateEnumMutationVariables = Exact<{
  input: CreateEnumInput;
}>;


export type CreateEnumMutation = { __typename?: 'Mutation', createEnum: { __typename?: 'CreateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateEnumMutationVariables = Exact<{
  name: Scalars['String']['input'];
  input: UpdateEnumInput;
}>;


export type UpdateEnumMutation = { __typename?: 'Mutation', updateEnum: { __typename?: 'UpdateEnumPayload', enum?: { __typename?: 'EnumDefinition', id: string, projectSlug: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, createdAt: string, updatedAt: string, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, error?:
      | { __typename: 'EnumNotFound', message: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
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

export type GetModelEnumSourceFieldsQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetModelEnumSourceFieldsQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, enum?: { __typename?: 'EnumDefinition', name: string } | null }> } | null } };

export type GetModelsQueryVariables = Exact<{
  input?: InputMaybe<ModelQueryInput>;
}>;


export type GetModelsQuery = { __typename?: 'Query', models: { __typename?: 'ModelConnection', totalCount: number, edges: Array<{ __typename?: 'ModelEdge', cursor: string, node: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null } } };

export type GetModelsForRelationQueryVariables = Exact<{
  input?: InputMaybe<ModelQueryInput>;
}>;


export type GetModelsForRelationQuery = { __typename?: 'Query', models: { __typename?: 'ModelConnection', edges: Array<{ __typename?: 'ModelEdge', node: { __typename?: 'Model', id: string, name: string, title: string, databaseName: string } }> } };

export type GetModelQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  withActualSchema?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetModelQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, displayField?: string | null, databaseName: string, storageType: string, jsonSchema?: string | null, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, isDeprecated: boolean, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, dbColumn?: { __typename?: 'DbColumnInfo', columnType: string, unique: boolean, nonNull: boolean, defaultValue?: string | null, constraints: Array<ActualConstraintType>, foreignKey?: { __typename?: 'ActualForeignKey', referencedTable: string, referencedColumn: string, constraintName: string } | null, conflicts: Array<{ __typename?: 'FieldConflict', aspect: FieldConflictAspect, expected: string, actual: string }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type GetModelRecordWorkspaceQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetModelRecordWorkspaceQuery = { __typename?: 'Query', model: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, name: string, title: string, description: string, databaseName: string, jsonSchema?: string | null, fields: Array<{ __typename?: 'Field', name: string, isDeprecated: boolean }> } | null, error?:
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound' }
     | null } };

export type GetModelByNameQueryVariables = Exact<{
  name: Scalars['String']['input'];
  databaseName: Scalars['String']['input'];
}>;


export type GetModelByNameQuery = { __typename?: 'Query', modelByName: { __typename?: 'GetModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
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

export type CreateModelMutationVariables = Exact<{
  input: CreateModelInput;
}>;


export type CreateModelMutation = { __typename?: 'Mutation', createModel: { __typename?: 'CreateModelPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'ModelAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'ModelTableAlreadyExists', message: string, suggestion?: string | null }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateModelMetaMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateModelMetaInput;
}>;


export type UpdateModelMetaMutation = { __typename?: 'Mutation', updateModelMeta: { __typename?: 'UpdateModelMetaPayload', success: boolean, model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
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


export type AddFieldsMutation = { __typename?: 'Mutation', addFields: { __typename?: 'AddFieldsPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, isDeprecated: boolean, isArray: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?: { __typename: 'InvalidInput', message: string, suggestion?: string | null } | null } };

export type UpdateFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
  input: UpdateFieldInput;
}>;


export type UpdateFieldMutation = { __typename?: 'Mutation', updateField: { __typename?: 'UpdateFieldPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, isDeprecated: boolean, isArray: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string, enum?: { __typename?: 'EnumDefinition', id: string, name: string, displayName: string, description?: string | null, isMultiSelect: boolean, options: Array<{ __typename?: 'EnumOption', code: string, label: string, order: number, description?: string | null }> } | null, validationConfig?: { __typename?: 'ValidationConfig', minLength?: any | null, maxLength?: any | null, pattern?: string | null, minimum?: number | null, maximum?: number | null } | null }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'FieldFormatImmutable', message: string, code: string }
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
     | null } };

export type RemoveFieldMutationVariables = Exact<{
  modelID: Scalars['ID']['input'];
  fieldName: Scalars['String']['input'];
}>;


export type RemoveFieldMutation = { __typename?: 'Mutation', removeField: { __typename?: 'RemoveFieldPayload', model?: { __typename?: 'Model', id: string, projectSlug: string, name: string, title: string, description: string, databaseName: string, storageType: string, dbTable?: DbTableStatus | null, createdAt: string, updatedAt: string, fields: Array<{ __typename?: 'Field', name: string, title: string, format: FormatType, schemaType: SchemaType, storageHint: string, nonNull: boolean, required: boolean, isPrimary: boolean, isUnique: boolean, description?: string | null, relateFkId?: string | null, belongsToFkId?: string | null, createdAt: string, updatedAt: string }>, group: { __typename?: 'ModelGroup', id: string, name: string, isVirtual: boolean, displayOrder: string } } | null, error?:
      | { __typename: 'FieldReferenceInUse', message: string, code: string, suggestion?: string | null }
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


export type GetProjectQuery = { __typename?: 'Query', project: { __typename?: 'GetProjectPayload', project?: { __typename?: 'Project', id: string, slug: string, title: string, description: string, status: ProjectStatus, orgName: string, createdAt: string, updatedAt: string } | null, error?: { __typename: 'ProjectNotFound', message: string } | null } };

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
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
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

export type GetEndUserPermissionsQueryVariables = Exact<{ [key: string]: never; }>;


export type GetEndUserPermissionsQuery = { __typename?: 'Query', endUserPermissions: { __typename?: 'EndUserPermissionConnection', totalCount: number, edges: Array<{ __typename?: 'EndUserPermissionEdge', node: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType, displayName?: string | null, description?: string | null, createdAt: any, updatedAt: any, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type GetEndUserBundlesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetEndUserBundlesQuery = { __typename?: 'Query', endUserPermissionBundles: { __typename?: 'EndUserPermissionBundleConnection', totalCount: number, edges: Array<{ __typename?: 'EndUserPermissionBundleEdge', node: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, createdAt: any, updatedAt: any, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } } }> } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type GetEndUserBundleQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetEndUserBundleQuery = { __typename?: 'Query', endUserPermissionBundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, createdAt: any, updatedAt: any, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType, displayName?: string | null, description?: string | null, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } } }> } | null };

export type GetEndUserRolesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetEndUserRolesQuery = { __typename?: 'Query', endUserRoles: { __typename?: 'EndUserRoleConnection', totalCount: number, edges: Array<{ __typename?: 'EndUserRoleEdge', node: { __typename?: 'EndUserRole', id: string, name: string, description?: string | null, isImplicit: boolean, createdAt: any, updatedAt: any, permissionBundles: Array<{ __typename?: 'EndUserRoleBundleEntry', bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null } }> } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, endCursor?: string | null } } };

export type GetEndUserRoleQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetEndUserRoleQuery = { __typename?: 'Query', endUserRole?: { __typename?: 'EndUserRole', id: string, name: string, description?: string | null, isImplicit: boolean, createdAt: any, updatedAt: any, permissionBundles: Array<{ __typename?: 'EndUserRoleBundleEntry', bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType } }> } }> } | null };

export type GetEndUserEffectivePermissionsQueryVariables = Exact<{
  endUserId: Scalars['ID']['input'];
  modelId: Scalars['ID']['input'];
}>;


export type GetEndUserEffectivePermissionsQuery = { __typename?: 'Query', effectivePermissions: { __typename?: 'GetEffectivePermissionsPayload', effectivePermissions?: { __typename?: 'EffectivePermissions', endUserId: string, modelId: string, grants: Array<{ __typename?: 'EffectiveGrant', action: RbacAction, rowScope: RowScopeType, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } }> } | null, error?:
      | { __typename: 'EndUserNotFoundInProject', message: string }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound' }
     | null } };

export type GetEndUserRoleAssignmentsQueryVariables = Exact<{
  endUserId: Scalars['ID']['input'];
}>;


export type GetEndUserRoleAssignmentsQuery = { __typename?: 'Query', endUserRoleAssignments: Array<{ __typename?: 'EndUserRoleAssignment', endUserId: string, assignedAt: any, role: { __typename?: 'EndUserRole', id: string, name: string, description?: string | null, isImplicit: boolean } }> };

export type GetEndUserBundleAssignmentsQueryVariables = Exact<{
  endUserId: Scalars['ID']['input'];
}>;


export type GetEndUserBundleAssignmentsQuery = { __typename?: 'Query', endUserBundleAssignments: Array<{ __typename?: 'EndUserBundleAssignment', endUserId: string, assignedAt: any, bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType } }> } }> };

export type CreateEndUserPermissionMutationVariables = Exact<{
  input: CreateEndUserPermissionInput;
}>;


export type CreateEndUserPermissionMutation = { __typename?: 'Mutation', createEndUserPermission: { __typename?: 'CreateEndUserPermissionPayload', permission?: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType, displayName?: string | null, description?: string | null, createdAt: any, updatedAt: any, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } } | null, error?:
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
      | { __typename: 'RowScopeFieldMissing', message: string, missingField: string, requiredByRowScope: RowScopeType }
     | null } };

export type DeleteEndUserPermissionMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteEndUserPermissionMutation = { __typename?: 'Mutation', deleteEndUserPermission: { __typename?: 'DeleteEndUserPermissionPayload', success: boolean, error?:
      | { __typename: 'EndUserPermissionInUse', message: string }
      | { __typename: 'EndUserPermissionNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type CreateEndUserBundleMutationVariables = Exact<{
  input: CreateEndUserPermissionBundleInput;
}>;


export type CreateEndUserBundleMutation = { __typename?: 'Mutation', createEndUserPermissionBundle: { __typename?: 'CreateEndUserPermissionBundlePayload', bundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, createdAt: any, updatedAt: any, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType } }> } | null, error?:
      | { __typename: 'EndUserPermissionBundleAlreadyExists', message: string }
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type UpdateEndUserBundleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateEndUserPermissionBundleInput;
}>;


export type UpdateEndUserBundleMutation = { __typename?: 'Mutation', updateEndUserPermissionBundle: { __typename?: 'UpdateEndUserPermissionBundlePayload', bundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null, updatedAt: any } | null, error?:
      | { __typename: 'EndUserPermissionBundleAlreadyExists', message: string }
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type DeleteEndUserBundleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteEndUserBundleMutation = { __typename?: 'Mutation', deleteEndUserPermissionBundle: { __typename?: 'DeleteEndUserPermissionBundlePayload', success: boolean, error?:
      | { __typename: 'EndUserPermissionBundleInUse', message: string }
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type AddPermissionToBundleMutationVariables = Exact<{
  input: AddEndUserPermissionToBundleInput;
}>;


export type AddPermissionToBundleMutation = { __typename?: 'Mutation', addEndUserPermissionToBundle: { __typename?: 'AddEndUserPermissionToBundlePayload', bundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType } }> } | null, error?:
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'EndUserPermissionNotFound', message: string }
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type RemovePermissionFromBundleMutationVariables = Exact<{
  input: RemoveEndUserPermissionFromBundleInput;
}>;


export type RemovePermissionFromBundleMutation = { __typename?: 'Mutation', removeEndUserPermissionFromBundle: { __typename?: 'RemoveEndUserPermissionFromBundlePayload', bundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string, permissions: Array<{ __typename?: 'EndUserBundlePermissionEntry', sortOrder: number, permission: { __typename?: 'EndUserPermission', id: string, modelId: string, action: RbacAction, rowScope: RowScopeType } }> } | null, error?:
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'EndUserPermissionNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type CreateEndUserRoleMutationVariables = Exact<{
  input: CreateEndUserRoleInput;
}>;


export type CreateEndUserRoleMutation = { __typename?: 'Mutation', createEndUserRole: { __typename?: 'CreateEndUserRolePayload', role?: { __typename?: 'EndUserRole', id: string, name: string, description?: string | null, isImplicit: boolean, createdAt: any, updatedAt: any, permissionBundles: Array<{ __typename?: 'EndUserRoleBundleEntry', bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string } }> } | null, error?:
      | { __typename: 'EndUserRoleAlreadyExists', message: string }
      | { __typename: 'InvalidInput', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type DeleteEndUserRoleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteEndUserRoleMutation = { __typename?: 'Mutation', deleteEndUserRole: { __typename?: 'DeleteEndUserRolePayload', success: boolean, error?:
      | { __typename: 'EndUserImplicitRoleCannotBeModified', message: string }
      | { __typename: 'EndUserRoleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type AssignBundleToRoleMutationVariables = Exact<{
  input: AssignBundleToEndUserRoleInput;
}>;


export type AssignBundleToRoleMutation = { __typename?: 'Mutation', assignBundleToEndUserRole: { __typename?: 'AssignBundleToEndUserRolePayload', role?: { __typename?: 'EndUserRole', id: string, name: string, permissionBundles: Array<{ __typename?: 'EndUserRoleBundleEntry', bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null } }> } | null, error?:
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'EndUserRoleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type RevokeBundleFromRoleMutationVariables = Exact<{
  input: RevokeBundleFromEndUserRoleInput;
}>;


export type RevokeBundleFromRoleMutation = { __typename?: 'Mutation', revokeBundleFromEndUserRole: { __typename?: 'RevokeBundleFromEndUserRolePayload', role?: { __typename?: 'EndUserRole', id: string, name: string, permissionBundles: Array<{ __typename?: 'EndUserRoleBundleEntry', bundle: { __typename?: 'EndUserPermissionBundle', id: string, name: string, description?: string | null } }> } | null, error?:
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'EndUserRoleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type AssignEndUserRoleMutationVariables = Exact<{
  input: AssignEndUserRoleInput;
}>;


export type AssignEndUserRoleMutation = { __typename?: 'Mutation', assignEndUserRole: { __typename?: 'AssignEndUserRolePayload', endUserId: string, role?: { __typename?: 'EndUserRole', id: string, name: string } | null, error?:
      | { __typename: 'EndUserCannotAssignImplicitRole', message: string }
      | { __typename: 'EndUserNotFoundInProject', message: string }
      | { __typename: 'EndUserRoleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
      | { __typename: 'UserRoleAlreadyAssigned', message: string }
     | null } };

export type RevokeEndUserRoleMutationVariables = Exact<{
  input: RevokeEndUserRoleInput;
}>;


export type RevokeEndUserRoleMutation = { __typename?: 'Mutation', revokeEndUserRole: { __typename?: 'RevokeEndUserRolePayload', success: boolean, error?:
      | { __typename: 'EndUserNotFoundInProject', message: string }
      | { __typename: 'EndUserRoleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type AssignBundleToEndUserMutationVariables = Exact<{
  input: AssignBundleToEndUserInput;
}>;


export type AssignBundleToEndUserMutation = { __typename?: 'Mutation', assignBundleToEndUser: { __typename?: 'AssignBundleToEndUserPayload', endUserId: string, bundle?: { __typename?: 'EndUserPermissionBundle', id: string, name: string } | null, error?:
      | { __typename: 'EndUserNotFoundInProject', message: string }
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
      | { __typename: 'UserBundleAlreadyAssigned', message: string }
     | null } };

export type RevokeBundleFromEndUserMutationVariables = Exact<{
  input: RevokeBundleFromEndUserInput;
}>;


export type RevokeBundleFromEndUserMutation = { __typename?: 'Mutation', revokeBundleFromEndUser: { __typename?: 'RevokeBundleFromEndUserPayload', success: boolean, error?:
      | { __typename: 'EndUserNotFoundInProject', message: string }
      | { __typename: 'EndUserPermissionBundleNotFound', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

export type ApplyEndUserPresetPolicyMutationVariables = Exact<{
  input: ApplyEndUserPresetPolicyInput;
}>;


export type ApplyEndUserPresetPolicyMutation = { __typename?: 'Mutation', applyEndUserPresetPolicy: { __typename?: 'ApplyEndUserPresetPolicyPayload', permissions: Array<{ __typename?: 'EndUserPermission', id: string, modelId: string, preset?: EndUserPermissionPreset | null, displayName?: string | null, description?: string | null, createdAt: any, updatedAt: any, columnPolicy: { __typename?: 'ColumnPolicy', defaultMode: ColumnAccessMode, rules: Array<{ __typename?: 'ColumnRule', fieldName: string, mode: ColumnAccessMode, maskPattern?: string | null }> } }>, error?:
      | { __typename: 'ModelNotFound', message: string }
      | { __typename: 'PresetRequiresOwnerField', message: string }
      | { __typename: 'ProjectNotFound', message: string }
     | null } };

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

export type UpdateOrganizationMutationVariables = Exact<{
  input: UpdateOrganizationInput;
}>;


export type UpdateOrganizationMutation = { __typename?: 'Mutation', updateOrganization: { __typename?: 'UpdateOrganizationPayload', organization?: { __typename?: 'Organization', id: string, name: string, displayName?: string | null, ownerID: string, status: OrganizationStatus, createdAt: string, updatedAt: string } | null, error?: { __typename: 'OrganizationNotFound', message: string } | null } };

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
      | { __typename: 'RoleNotFound', message: string }
     | null } };

export type AddPermissionToRoleMutationVariables = Exact<{
  roleId: Scalars['Int']['input'];
  obj: Scalars['String']['input'];
  act: Scalars['String']['input'];
}>;


export type AddPermissionToRoleMutation = { __typename?: 'Mutation', addPermissionToRole: { __typename?: 'AddRolePermissionPayload', success: boolean, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'PermissionRoleNotFound', message: string, suggestion?: string | null }
      | { __typename: 'PermissionSystemRoleCannotBeModified', message: string, suggestion?: string | null }
     | null } };

export type RemovePermissionFromRoleMutationVariables = Exact<{
  roleId: Scalars['Int']['input'];
  obj: Scalars['String']['input'];
  act: Scalars['String']['input'];
}>;


export type RemovePermissionFromRoleMutation = { __typename?: 'Mutation', removePermissionFromRole: { __typename?: 'RemoveRolePermissionPayload', success: boolean, error?:
      | { __typename: 'InvalidInput', message: string, suggestion?: string | null }
      | { __typename: 'PermissionRoleNotFound', message: string, suggestion?: string | null }
      | { __typename: 'PermissionSystemRoleCannotBeModified', message: string, suggestion?: string | null }
     | null } };


/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetClusterQuery(
 *   ({ query, variables }) => {
 *     const { projectSlug } = variables;
 *     return HttpResponse.json({
 *       data: { databaseCluster }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetClusterQuery = (resolver: GraphQLResponseResolver<GetClusterQuery, GetClusterQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetClusterQuery, GetClusterQueryVariables>(
    'GetCluster',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockListDatabasesQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { listDatabases }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockListDatabasesQuery = (resolver: GraphQLResponseResolver<ListDatabasesQuery, ListDatabasesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<ListDatabasesQuery, ListDatabasesQueryVariables>(
    'ListDatabases',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockModelDatabaseCatalogQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { modelDatabaseCatalog }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockModelDatabaseCatalogQuery = (resolver: GraphQLResponseResolver<ModelDatabaseCatalogQuery, ModelDatabaseCatalogQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<ModelDatabaseCatalogQuery, ModelDatabaseCatalogQueryVariables>(
    'ModelDatabaseCatalog',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockTestClusterConnectionMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { testDatabaseConnection }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockTestClusterConnectionMutation = (resolver: GraphQLResponseResolver<TestClusterConnectionMutation, TestClusterConnectionMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<TestClusterConnectionMutation, TestClusterConnectionMutationVariables>(
    'TestClusterConnection',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockListEndUsersQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { listEndUsers }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockListEndUsersQuery = (resolver: GraphQLResponseResolver<ListEndUsersQuery, ListEndUsersQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<ListEndUsersQuery, ListEndUsersQueryVariables>(
    'ListEndUsers',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateEndUserMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createEndUser }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateEndUserMutation = (resolver: GraphQLResponseResolver<CreateEndUserMutation, CreateEndUserMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateEndUserMutation, CreateEndUserMutationVariables>(
    'CreateEndUser',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateEndUserStatusMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { updateEndUserStatus }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateEndUserStatusMutation = (resolver: GraphQLResponseResolver<UpdateEndUserStatusMutation, UpdateEndUserStatusMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateEndUserStatusMutation, UpdateEndUserStatusMutationVariables>(
    'UpdateEndUserStatus',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteEndUserMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { deleteEndUser }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteEndUserMutation = (resolver: GraphQLResponseResolver<DeleteEndUserMutation, DeleteEndUserMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteEndUserMutation, DeleteEndUserMutationVariables>(
    'DeleteEndUser',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEnumsQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { enums }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEnumsQuery = (resolver: GraphQLResponseResolver<GetEnumsQuery, GetEnumsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEnumsQuery, GetEnumsQueryVariables>(
    'GetEnums',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEnumQuery(
 *   ({ query, variables }) => {
 *     const { name } = variables;
 *     return HttpResponse.json({
 *       data: { enum }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEnumQuery = (resolver: GraphQLResponseResolver<GetEnumQuery, GetEnumQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEnumQuery, GetEnumQueryVariables>(
    'GetEnum',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEnumReferencesQuery(
 *   ({ query, variables }) => {
 *     const { name } = variables;
 *     return HttpResponse.json({
 *       data: { enumReferences }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEnumReferencesQuery = (resolver: GraphQLResponseResolver<GetEnumReferencesQuery, GetEnumReferencesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEnumReferencesQuery, GetEnumReferencesQueryVariables>(
    'GetEnumReferences',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateEnumMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createEnum }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateEnumMutation = (resolver: GraphQLResponseResolver<CreateEnumMutation, CreateEnumMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateEnumMutation, CreateEnumMutationVariables>(
    'CreateEnum',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateEnumMutation(
 *   ({ query, variables }) => {
 *     const { name, input } = variables;
 *     return HttpResponse.json({
 *       data: { updateEnum }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateEnumMutation = (resolver: GraphQLResponseResolver<UpdateEnumMutation, UpdateEnumMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateEnumMutation, UpdateEnumMutationVariables>(
    'UpdateEnum',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteEnumMutation(
 *   ({ query, variables }) => {
 *     const { name } = variables;
 *     return HttpResponse.json({
 *       data: { deleteEnum }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteEnumMutation = (resolver: GraphQLResponseResolver<DeleteEnumMutation, DeleteEnumMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteEnumMutation, DeleteEnumMutationVariables>(
    'DeleteEnum',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelEnumSourceFieldsQuery(
 *   ({ query, variables }) => {
 *     const { id, withActualSchema } = variables;
 *     return HttpResponse.json({
 *       data: { model }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelEnumSourceFieldsQuery = (resolver: GraphQLResponseResolver<GetModelEnumSourceFieldsQuery, GetModelEnumSourceFieldsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelEnumSourceFieldsQuery, GetModelEnumSourceFieldsQueryVariables>(
    'GetModelEnumSourceFields',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelsQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { models }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelsQuery = (resolver: GraphQLResponseResolver<GetModelsQuery, GetModelsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelsQuery, GetModelsQueryVariables>(
    'GetModels',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelsForRelationQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { models }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelsForRelationQuery = (resolver: GraphQLResponseResolver<GetModelsForRelationQuery, GetModelsForRelationQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelsForRelationQuery, GetModelsForRelationQueryVariables>(
    'GetModelsForRelation',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelQuery(
 *   ({ query, variables }) => {
 *     const { id, withActualSchema } = variables;
 *     return HttpResponse.json({
 *       data: { model }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelQuery = (resolver: GraphQLResponseResolver<GetModelQuery, GetModelQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelQuery, GetModelQueryVariables>(
    'GetModel',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelRecordWorkspaceQuery(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { model }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelRecordWorkspaceQuery = (resolver: GraphQLResponseResolver<GetModelRecordWorkspaceQuery, GetModelRecordWorkspaceQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelRecordWorkspaceQuery, GetModelRecordWorkspaceQueryVariables>(
    'GetModelRecordWorkspace',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelByNameQuery(
 *   ({ query, variables }) => {
 *     const { name, databaseName } = variables;
 *     return HttpResponse.json({
 *       data: { modelByName }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelByNameQuery = (resolver: GraphQLResponseResolver<GetModelByNameQuery, GetModelByNameQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelByNameQuery, GetModelByNameQueryVariables>(
    'GetModelByName',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelGroupsQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { modelGroups }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelGroupsQuery = (resolver: GraphQLResponseResolver<GetModelGroupsQuery, GetModelGroupsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelGroupsQuery, GetModelGroupsQueryVariables>(
    'GetModelGroups',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetModelJsonSchemaQuery(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { modelJsonSchema }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetModelJsonSchemaQuery = (resolver: GraphQLResponseResolver<GetModelJsonSchemaQuery, GetModelJsonSchemaQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetModelJsonSchemaQuery, GetModelJsonSchemaQueryVariables>(
    'GetModelJsonSchema',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetLogicalForeignKeysQuery(
 *   ({ query, variables }) => {
 *     const { modelId } = variables;
 *     return HttpResponse.json({
 *       data: { logicalForeignKeys }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetLogicalForeignKeysQuery = (resolver: GraphQLResponseResolver<GetLogicalForeignKeysQuery, GetLogicalForeignKeysQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetLogicalForeignKeysQuery, GetLogicalForeignKeysQueryVariables>(
    'GetLogicalForeignKeys',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateModelMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createModel }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateModelMutation = (resolver: GraphQLResponseResolver<CreateModelMutation, CreateModelMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateModelMutation, CreateModelMutationVariables>(
    'CreateModel',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateModelMetaMutation(
 *   ({ query, variables }) => {
 *     const { id, input } = variables;
 *     return HttpResponse.json({
 *       data: { updateModelMeta }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateModelMetaMutation = (resolver: GraphQLResponseResolver<UpdateModelMetaMutation, UpdateModelMetaMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateModelMetaMutation, UpdateModelMetaMutationVariables>(
    'UpdateModelMeta',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteModelMutation(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { deleteModel }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteModelMutation = (resolver: GraphQLResponseResolver<DeleteModelMutation, DeleteModelMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteModelMutation, DeleteModelMutationVariables>(
    'DeleteModel',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockSyncModelSchemaMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { syncModelSchema }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockSyncModelSchemaMutation = (resolver: GraphQLResponseResolver<SyncModelSchemaMutation, SyncModelSchemaMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<SyncModelSchemaMutation, SyncModelSchemaMutationVariables>(
    'SyncModelSchema',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRepairModelMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { repairModel }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRepairModelMutation = (resolver: GraphQLResponseResolver<RepairModelMutation, RepairModelMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RepairModelMutation, RepairModelMutationVariables>(
    'RepairModel',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateModelFromSchemaMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createModelFromSchema }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateModelFromSchemaMutation = (resolver: GraphQLResponseResolver<CreateModelFromSchemaMutation, CreateModelFromSchemaMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateModelFromSchemaMutation, CreateModelFromSchemaMutationVariables>(
    'CreateModelFromSchema',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockImportModelMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { importModel }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockImportModelMutation = (resolver: GraphQLResponseResolver<ImportModelMutation, ImportModelMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<ImportModelMutation, ImportModelMutationVariables>(
    'ImportModel',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateGroupMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createGroup }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateGroupMutation = (resolver: GraphQLResponseResolver<CreateGroupMutation, CreateGroupMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateGroupMutation, CreateGroupMutationVariables>(
    'CreateGroup',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRenameGroupMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { renameGroup }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRenameGroupMutation = (resolver: GraphQLResponseResolver<RenameGroupMutation, RenameGroupMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RenameGroupMutation, RenameGroupMutationVariables>(
    'RenameGroup',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteGroupMutation(
 *   ({ query, variables }) => {
 *     const { groupId } = variables;
 *     return HttpResponse.json({
 *       data: { deleteGroup }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteGroupMutation = (resolver: GraphQLResponseResolver<DeleteGroupMutation, DeleteGroupMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteGroupMutation, DeleteGroupMutationVariables>(
    'DeleteGroup',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockReorderGroupMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { reorderGroup }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockReorderGroupMutation = (resolver: GraphQLResponseResolver<ReorderGroupMutation, ReorderGroupMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<ReorderGroupMutation, ReorderGroupMutationVariables>(
    'ReorderGroup',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockMoveModelToGroupMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { moveModelToGroup }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockMoveModelToGroupMutation = (resolver: GraphQLResponseResolver<MoveModelToGroupMutation, MoveModelToGroupMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<MoveModelToGroupMutation, MoveModelToGroupMutationVariables>(
    'MoveModelToGroup',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAddFieldsMutation(
 *   ({ query, variables }) => {
 *     const { modelID, input } = variables;
 *     return HttpResponse.json({
 *       data: { addFields }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAddFieldsMutation = (resolver: GraphQLResponseResolver<AddFieldsMutation, AddFieldsMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AddFieldsMutation, AddFieldsMutationVariables>(
    'AddFields',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateFieldMutation(
 *   ({ query, variables }) => {
 *     const { modelID, fieldName, input } = variables;
 *     return HttpResponse.json({
 *       data: { updateField }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateFieldMutation = (resolver: GraphQLResponseResolver<UpdateFieldMutation, UpdateFieldMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateFieldMutation, UpdateFieldMutationVariables>(
    'UpdateField',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRemoveFieldMutation(
 *   ({ query, variables }) => {
 *     const { modelID, fieldName } = variables;
 *     return HttpResponse.json({
 *       data: { removeField }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRemoveFieldMutation = (resolver: GraphQLResponseResolver<RemoveFieldMutation, RemoveFieldMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RemoveFieldMutation, RemoveFieldMutationVariables>(
    'RemoveField',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeprecateFieldMutation(
 *   ({ query, variables }) => {
 *     const { modelID, fieldName } = variables;
 *     return HttpResponse.json({
 *       data: { deprecateField }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeprecateFieldMutation = (resolver: GraphQLResponseResolver<DeprecateFieldMutation, DeprecateFieldMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeprecateFieldMutation, DeprecateFieldMutationVariables>(
    'DeprecateField',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUndeprecateFieldMutation(
 *   ({ query, variables }) => {
 *     const { modelID, fieldName } = variables;
 *     return HttpResponse.json({
 *       data: { undeprecateField }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUndeprecateFieldMutation = (resolver: GraphQLResponseResolver<UndeprecateFieldMutation, UndeprecateFieldMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UndeprecateFieldMutation, UndeprecateFieldMutationVariables>(
    'UndeprecateField',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateLogicalForeignKeyMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createLogicalForeignKey }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateLogicalForeignKeyMutation = (resolver: GraphQLResponseResolver<CreateLogicalForeignKeyMutation, CreateLogicalForeignKeyMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateLogicalForeignKeyMutation, CreateLogicalForeignKeyMutationVariables>(
    'CreateLogicalForeignKey',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteLogicalForeignKeyMutation(
 *   ({ query, variables }) => {
 *     const { pairId } = variables;
 *     return HttpResponse.json({
 *       data: { deleteLogicalForeignKey }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteLogicalForeignKeyMutation = (resolver: GraphQLResponseResolver<DeleteLogicalForeignKeyMutation, DeleteLogicalForeignKeyMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteLogicalForeignKeyMutation, DeleteLogicalForeignKeyMutationVariables>(
    'DeleteLogicalForeignKey',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockNoopQueryQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { __typename }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockNoopQueryQuery = (resolver: GraphQLResponseResolver<NoopQueryQuery, NoopQueryQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<NoopQueryQuery, NoopQueryQueryVariables>(
    'NoopQuery',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockNoopMutationMutation(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { __typename }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockNoopMutationMutation = (resolver: GraphQLResponseResolver<NoopMutationMutation, NoopMutationMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<NoopMutationMutation, NoopMutationMutationVariables>(
    'NoopMutation',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockListTablesQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { listTables }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockListTablesQuery = (resolver: GraphQLResponseResolver<ListTablesQuery, ListTablesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<ListTablesQuery, ListTablesQueryVariables>(
    'ListTables',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetProjectsQuery(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { projects }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetProjectsQuery = (resolver: GraphQLResponseResolver<GetProjectsQuery, GetProjectsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetProjectsQuery, GetProjectsQueryVariables>(
    'GetProjects',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetProjectQuery(
 *   ({ query, variables }) => {
 *     const { slug } = variables;
 *     return HttpResponse.json({
 *       data: { project }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetProjectQuery = (resolver: GraphQLResponseResolver<GetProjectQuery, GetProjectQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetProjectQuery, GetProjectQueryVariables>(
    'GetProject',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateProjectMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createProject }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateProjectMutation = (resolver: GraphQLResponseResolver<CreateProjectMutation, CreateProjectMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateProjectMutation, CreateProjectMutationVariables>(
    'CreateProject',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateProjectMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { updateProject }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateProjectMutation = (resolver: GraphQLResponseResolver<UpdateProjectMutation, UpdateProjectMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateProjectMutation, UpdateProjectMutationVariables>(
    'UpdateProject',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteProjectMutation(
 *   ({ query, variables }) => {
 *     const { slug } = variables;
 *     return HttpResponse.json({
 *       data: { deleteProject }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteProjectMutation = (resolver: GraphQLResponseResolver<DeleteProjectMutation, DeleteProjectMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteProjectMutation, DeleteProjectMutationVariables>(
    'DeleteProject',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateProjectClusterMutation(
 *   ({ query, variables }) => {
 *     const { projectSlug, input } = variables;
 *     return HttpResponse.json({
 *       data: { updateProjectCluster }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateProjectClusterMutation = (resolver: GraphQLResponseResolver<UpdateProjectClusterMutation, UpdateProjectClusterMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateProjectClusterMutation, UpdateProjectClusterMutationVariables>(
    'UpdateProjectCluster',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockTestDatabaseConnectionMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { testDatabaseConnection }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockTestDatabaseConnectionMutation = (resolver: GraphQLResponseResolver<TestDatabaseConnectionMutation, TestDatabaseConnectionMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<TestDatabaseConnectionMutation, TestDatabaseConnectionMutationVariables>(
    'TestDatabaseConnection',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserPermissionsQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { endUserPermissions }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserPermissionsQuery = (resolver: GraphQLResponseResolver<GetEndUserPermissionsQuery, GetEndUserPermissionsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserPermissionsQuery, GetEndUserPermissionsQueryVariables>(
    'GetEndUserPermissions',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserBundlesQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { endUserPermissionBundles }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserBundlesQuery = (resolver: GraphQLResponseResolver<GetEndUserBundlesQuery, GetEndUserBundlesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserBundlesQuery, GetEndUserBundlesQueryVariables>(
    'GetEndUserBundles',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserBundleQuery(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { endUserPermissionBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserBundleQuery = (resolver: GraphQLResponseResolver<GetEndUserBundleQuery, GetEndUserBundleQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserBundleQuery, GetEndUserBundleQueryVariables>(
    'GetEndUserBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserRolesQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { endUserRoles }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserRolesQuery = (resolver: GraphQLResponseResolver<GetEndUserRolesQuery, GetEndUserRolesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserRolesQuery, GetEndUserRolesQueryVariables>(
    'GetEndUserRoles',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserRoleQuery(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { endUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserRoleQuery = (resolver: GraphQLResponseResolver<GetEndUserRoleQuery, GetEndUserRoleQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserRoleQuery, GetEndUserRoleQueryVariables>(
    'GetEndUserRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserEffectivePermissionsQuery(
 *   ({ query, variables }) => {
 *     const { endUserId, modelId } = variables;
 *     return HttpResponse.json({
 *       data: { effectivePermissions }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserEffectivePermissionsQuery = (resolver: GraphQLResponseResolver<GetEndUserEffectivePermissionsQuery, GetEndUserEffectivePermissionsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserEffectivePermissionsQuery, GetEndUserEffectivePermissionsQueryVariables>(
    'GetEndUserEffectivePermissions',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserRoleAssignmentsQuery(
 *   ({ query, variables }) => {
 *     const { endUserId } = variables;
 *     return HttpResponse.json({
 *       data: { endUserRoleAssignments }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserRoleAssignmentsQuery = (resolver: GraphQLResponseResolver<GetEndUserRoleAssignmentsQuery, GetEndUserRoleAssignmentsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserRoleAssignmentsQuery, GetEndUserRoleAssignmentsQueryVariables>(
    'GetEndUserRoleAssignments',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetEndUserBundleAssignmentsQuery(
 *   ({ query, variables }) => {
 *     const { endUserId } = variables;
 *     return HttpResponse.json({
 *       data: { endUserBundleAssignments }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetEndUserBundleAssignmentsQuery = (resolver: GraphQLResponseResolver<GetEndUserBundleAssignmentsQuery, GetEndUserBundleAssignmentsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetEndUserBundleAssignmentsQuery, GetEndUserBundleAssignmentsQueryVariables>(
    'GetEndUserBundleAssignments',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateEndUserPermissionMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createEndUserPermission }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateEndUserPermissionMutation = (resolver: GraphQLResponseResolver<CreateEndUserPermissionMutation, CreateEndUserPermissionMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateEndUserPermissionMutation, CreateEndUserPermissionMutationVariables>(
    'CreateEndUserPermission',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteEndUserPermissionMutation(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { deleteEndUserPermission }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteEndUserPermissionMutation = (resolver: GraphQLResponseResolver<DeleteEndUserPermissionMutation, DeleteEndUserPermissionMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteEndUserPermissionMutation, DeleteEndUserPermissionMutationVariables>(
    'DeleteEndUserPermission',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateEndUserBundleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createEndUserPermissionBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateEndUserBundleMutation = (resolver: GraphQLResponseResolver<CreateEndUserBundleMutation, CreateEndUserBundleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateEndUserBundleMutation, CreateEndUserBundleMutationVariables>(
    'CreateEndUserBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateEndUserBundleMutation(
 *   ({ query, variables }) => {
 *     const { id, input } = variables;
 *     return HttpResponse.json({
 *       data: { updateEndUserPermissionBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateEndUserBundleMutation = (resolver: GraphQLResponseResolver<UpdateEndUserBundleMutation, UpdateEndUserBundleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateEndUserBundleMutation, UpdateEndUserBundleMutationVariables>(
    'UpdateEndUserBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteEndUserBundleMutation(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { deleteEndUserPermissionBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteEndUserBundleMutation = (resolver: GraphQLResponseResolver<DeleteEndUserBundleMutation, DeleteEndUserBundleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteEndUserBundleMutation, DeleteEndUserBundleMutationVariables>(
    'DeleteEndUserBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAddPermissionToBundleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { addEndUserPermissionToBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAddPermissionToBundleMutation = (resolver: GraphQLResponseResolver<AddPermissionToBundleMutation, AddPermissionToBundleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AddPermissionToBundleMutation, AddPermissionToBundleMutationVariables>(
    'AddPermissionToBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRemovePermissionFromBundleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { removeEndUserPermissionFromBundle }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRemovePermissionFromBundleMutation = (resolver: GraphQLResponseResolver<RemovePermissionFromBundleMutation, RemovePermissionFromBundleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RemovePermissionFromBundleMutation, RemovePermissionFromBundleMutationVariables>(
    'RemovePermissionFromBundle',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateEndUserRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateEndUserRoleMutation = (resolver: GraphQLResponseResolver<CreateEndUserRoleMutation, CreateEndUserRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateEndUserRoleMutation, CreateEndUserRoleMutationVariables>(
    'CreateEndUserRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteEndUserRoleMutation(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { deleteEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteEndUserRoleMutation = (resolver: GraphQLResponseResolver<DeleteEndUserRoleMutation, DeleteEndUserRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteEndUserRoleMutation, DeleteEndUserRoleMutationVariables>(
    'DeleteEndUserRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAssignBundleToRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { assignBundleToEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAssignBundleToRoleMutation = (resolver: GraphQLResponseResolver<AssignBundleToRoleMutation, AssignBundleToRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AssignBundleToRoleMutation, AssignBundleToRoleMutationVariables>(
    'AssignBundleToRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRevokeBundleFromRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { revokeBundleFromEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRevokeBundleFromRoleMutation = (resolver: GraphQLResponseResolver<RevokeBundleFromRoleMutation, RevokeBundleFromRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RevokeBundleFromRoleMutation, RevokeBundleFromRoleMutationVariables>(
    'RevokeBundleFromRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAssignEndUserRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { assignEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAssignEndUserRoleMutation = (resolver: GraphQLResponseResolver<AssignEndUserRoleMutation, AssignEndUserRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AssignEndUserRoleMutation, AssignEndUserRoleMutationVariables>(
    'AssignEndUserRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRevokeEndUserRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { revokeEndUserRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRevokeEndUserRoleMutation = (resolver: GraphQLResponseResolver<RevokeEndUserRoleMutation, RevokeEndUserRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RevokeEndUserRoleMutation, RevokeEndUserRoleMutationVariables>(
    'RevokeEndUserRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAssignBundleToEndUserMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { assignBundleToEndUser }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAssignBundleToEndUserMutation = (resolver: GraphQLResponseResolver<AssignBundleToEndUserMutation, AssignBundleToEndUserMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AssignBundleToEndUserMutation, AssignBundleToEndUserMutationVariables>(
    'AssignBundleToEndUser',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRevokeBundleFromEndUserMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { revokeBundleFromEndUser }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRevokeBundleFromEndUserMutation = (resolver: GraphQLResponseResolver<RevokeBundleFromEndUserMutation, RevokeBundleFromEndUserMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RevokeBundleFromEndUserMutation, RevokeBundleFromEndUserMutationVariables>(
    'RevokeBundleFromEndUser',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockApplyEndUserPresetPolicyMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { applyEndUserPresetPolicy }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockApplyEndUserPresetPolicyMutation = (resolver: GraphQLResponseResolver<ApplyEndUserPresetPolicyMutation, ApplyEndUserPresetPolicyMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<ApplyEndUserPresetPolicyMutation, ApplyEndUserPresetPolicyMutationVariables>(
    'ApplyEndUserPresetPolicy',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetMeQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { me }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetMeQuery = (resolver: GraphQLResponseResolver<GetMeQuery, GetMeQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetMeQuery, GetMeQueryVariables>(
    'GetMe',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetMyOrganizationsQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { myOrganizations }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetMyOrganizationsQuery = (resolver: GraphQLResponseResolver<GetMyOrganizationsQuery, GetMyOrganizationsQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetMyOrganizationsQuery, GetMyOrganizationsQueryVariables>(
    'GetMyOrganizations',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetOrganizationMembersQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { organizationMembers }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetOrganizationMembersQuery = (resolver: GraphQLResponseResolver<GetOrganizationMembersQuery, GetOrganizationMembersQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetOrganizationMembersQuery, GetOrganizationMembersQueryVariables>(
    'GetOrganizationMembers',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetRolesQuery(
 *   ({ query, variables }) => {
 *     return HttpResponse.json({
 *       data: { roles }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetRolesQuery = (resolver: GraphQLResponseResolver<GetRolesQuery, GetRolesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetRolesQuery, GetRolesQueryVariables>(
    'GetRoles',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetPermissionRolesQuery(
 *   ({ query, variables }) => {
 *     const { orgName, includeSystem } = variables;
 *     return HttpResponse.json({
 *       data: { permissionRoles }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetPermissionRolesQuery = (resolver: GraphQLResponseResolver<GetPermissionRolesQuery, GetPermissionRolesQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetPermissionRolesQuery, GetPermissionRolesQueryVariables>(
    'GetPermissionRoles',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockGetRolePermissionsListQuery(
 *   ({ query, variables }) => {
 *     const { roleId } = variables;
 *     return HttpResponse.json({
 *       data: { rolePermissionsList }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockGetRolePermissionsListQuery = (resolver: GraphQLResponseResolver<GetRolePermissionsListQuery, GetRolePermissionsListQueryVariables>, options?: RequestHandlerOptions) =>
  graphql.query<GetRolePermissionsListQuery, GetRolePermissionsListQueryVariables>(
    'GetRolePermissionsList',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockUpdateOrganizationMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { updateOrganization }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockUpdateOrganizationMutation = (resolver: GraphQLResponseResolver<UpdateOrganizationMutation, UpdateOrganizationMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<UpdateOrganizationMutation, UpdateOrganizationMutationVariables>(
    'UpdateOrganization',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockCreateRoleMutation(
 *   ({ query, variables }) => {
 *     const { input } = variables;
 *     return HttpResponse.json({
 *       data: { createRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockCreateRoleMutation = (resolver: GraphQLResponseResolver<CreateRoleMutation, CreateRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<CreateRoleMutation, CreateRoleMutationVariables>(
    'CreateRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockDeleteRoleMutation(
 *   ({ query, variables }) => {
 *     const { id } = variables;
 *     return HttpResponse.json({
 *       data: { deleteRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockDeleteRoleMutation = (resolver: GraphQLResponseResolver<DeleteRoleMutation, DeleteRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<DeleteRoleMutation, DeleteRoleMutationVariables>(
    'DeleteRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockAddPermissionToRoleMutation(
 *   ({ query, variables }) => {
 *     const { roleId, obj, act } = variables;
 *     return HttpResponse.json({
 *       data: { addPermissionToRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockAddPermissionToRoleMutation = (resolver: GraphQLResponseResolver<AddPermissionToRoleMutation, AddPermissionToRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<AddPermissionToRoleMutation, AddPermissionToRoleMutationVariables>(
    'AddPermissionToRole',
    resolver,
    options
  )

/**
 * @param resolver A function that accepts [resolver arguments](https://mswjs.io/docs/api/graphql#resolver-argument) and must always return the instruction on what to do with the intercepted request. ([see more](https://mswjs.io/docs/concepts/response-resolver#resolver-instructions))
 * @param options Options object to customize the behavior of the mock. ([see more](https://mswjs.io/docs/api/graphql#handler-options))
 * @see https://mswjs.io/docs/basics/response-resolver
 * @example
 * mockRemovePermissionFromRoleMutation(
 *   ({ query, variables }) => {
 *     const { roleId, obj, act } = variables;
 *     return HttpResponse.json({
 *       data: { removePermissionFromRole }
 *     })
 *   },
 *   requestOptions
 * )
 */
export const mockRemovePermissionFromRoleMutation = (resolver: GraphQLResponseResolver<RemovePermissionFromRoleMutation, RemovePermissionFromRoleMutationVariables>, options?: RequestHandlerOptions) =>
  graphql.mutation<RemovePermissionFromRoleMutation, RemovePermissionFromRoleMutationVariables>(
    'RemovePermissionFromRole',
    resolver,
    options
  )

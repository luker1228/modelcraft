export function buildDatabaseManagementPath(orgName: string, projectSlug: string): string {
  return `/org/${orgName}/project/${projectSlug}/databases`
}

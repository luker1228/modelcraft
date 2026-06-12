export type ModelEditorViewMode = 'schema' | 'data'

interface BuildModelEditorPathOptions {
  view?: ModelEditorViewMode
  databaseName?: string | null
}

export function buildModelEditorPath(
  orgName: string,
  projectSlug: string,
  options: BuildModelEditorPathOptions = {}
): string {
  const params = new URLSearchParams()

  if (options.view === 'data') {
    params.set('view', 'data')
  }

  if (options.databaseName) {
    params.set('db', options.databaseName)
  }

  const query = params.toString()
  return `/org/${orgName}/project/${projectSlug}/model-editor${query ? `?${query}` : ''}`
}

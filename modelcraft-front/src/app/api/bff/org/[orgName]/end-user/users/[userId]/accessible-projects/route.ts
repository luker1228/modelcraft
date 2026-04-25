// src/app/api/bff/org/[orgName]/end-user/users/[userId]/accessible-projects/route.ts
// BFF: 获取指定终端用户可访问的 Project 列表（GraphQL 聚合）

import { NextRequest, NextResponse } from 'next/server'

interface RouteParams {
  params: Promise<{ orgName: string; userId: string }>
}

interface GraphQLErrorItem {
  message?: string
}

interface OrgUsersData {
  listEndUsers?: {
    connection?: {
      nodes?: Array<{
        id: string
        username: string
      }>
    }
  }
}

interface OrgProjectsData {
  projects?: Array<{
    slug: string
    title: string
  }>
}

interface ProjectAccessData {
  listProjectEndUserAccess?: {
    connection?: {
      nodes?: Array<{
        endUser: {
          id: string
        }
      }>
    }
  }
}

const ORG_LIST_END_USERS_QUERY = `
  query ListOrgEndUsers($input: ListEndUsersInput) {
    listEndUsers(input: $input) {
      connection {
        nodes {
          id
          username
        }
      }
    }
  }
`

const ORG_LIST_PROJECTS_QUERY = `
  query ListOrgProjects($input: ListProjectsInput) {
    projects(input: $input) {
      slug
      title
    }
  }
`

const PROJECT_LIST_ACCESS_QUERY = `
  query ListProjectEndUserAccess($input: ListProjectEndUserAccessInput) {
    listProjectEndUserAccess(input: $input) {
      connection {
        nodes {
          endUser {
            id
          }
        }
      }
    }
  }
`

async function postGraphQL<TData>(
  req: NextRequest,
  graphqlPath: string,
  query: string,
  variables?: Record<string, unknown>
): Promise<TData> {
  const url = new URL(graphqlPath, req.nextUrl.origin)
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const cookie = req.headers.get('cookie')
  const authorization = req.headers.get('authorization')
  if (cookie) headers.cookie = cookie
  if (authorization) headers.authorization = authorization

  const response = await fetch(url.toString(), {
    method: 'POST',
    headers,
    body: JSON.stringify({ query, variables }),
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(`GraphQL request failed: ${response.status}`)
  }

  const payload = (await response.json()) as {
    data?: TData
    errors?: GraphQLErrorItem[]
  }

  if (payload.errors && payload.errors.length > 0) {
    throw new Error(payload.errors[0]?.message || 'GraphQL error')
  }

  if (!payload.data) {
    throw new Error('GraphQL response missing data')
  }

  return payload.data
}

export async function GET(req: NextRequest, { params }: RouteParams) {
  const { orgName, userId } = await params

  try {
    const usersData = await postGraphQL<OrgUsersData>(
      req,
      `/graphql/org/${orgName}/`,
      ORG_LIST_END_USERS_QUERY,
      { input: { first: 500 } }
    )

    const user = (usersData.listEndUsers?.connection?.nodes ?? []).find((item) => item.id === userId)
    if (!user) {
      return NextResponse.json({ projects: [] })
    }

    const projectsData = await postGraphQL<OrgProjectsData>(
      req,
      `/graphql/org/${orgName}/`,
      ORG_LIST_PROJECTS_QUERY,
      { input: null }
    )

    const projects = projectsData.projects ?? []
    const checks = await Promise.all(
      projects.map(async (project) => {
        try {
          const accessData = await postGraphQL<ProjectAccessData>(
            req,
            `/graphql/org/${orgName}/project/${project.slug}/`,
            PROJECT_LIST_ACCESS_QUERY,
            { input: { search: user.username, first: 100 } }
          )

          const nodes = accessData.listProjectEndUserAccess?.connection?.nodes ?? []
          const hasAccess = nodes.some((node) => node.endUser.id === userId)
          return hasAccess ? { slug: project.slug, title: project.title } : null
        } catch {
          // 某些项目可能无权限读取，忽略并继续其余项目
          return null
        }
      })
    )

    return NextResponse.json({ projects: checks.filter((item) => item !== null) })
  } catch {
    return NextResponse.json({ error: { message: '获取关联项目失败' } }, { status: 500 })
  }
}

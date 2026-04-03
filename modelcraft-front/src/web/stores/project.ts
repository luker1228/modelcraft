import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import type { Project } from '@/types'

interface ProjectState {
  // 状态
  projects: Project[]
  selectedProject: Project | null
  loading: boolean
  error: string | null

  // Actions
  setProjects: (projects: Project[]) => void
  addProject: (project: Project) => void
  updateProject: (identifier: string, updates: Partial<Project>) => void // identifier can be id or name
  removeProject: (identifier: string) => void // identifier can be id or name
  setSelectedProject: (project: Project | null) => void
  findProjectById: (id: string) => Project | undefined
  findProjectByName: (name: string) => Project | undefined
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearProjects: () => void
  clearSelectedProject: () => void
}

export const useProjectStore = create<ProjectState>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        projects: [],
        selectedProject: null,
        loading: false,
        error: null,

        // Actions
        setProjects: (projects) => set({ projects }),

        addProject: (project) => set((state) => ({
          projects: [...state.projects, project]
        })),

        updateProject: (identifier, updates) => set((state) => ({
          projects: state.projects.map(project => 
            project.id === identifier || project.slug === identifier 
              ? { ...project, ...updates } 
              : project
          ),
          selectedProject: (state.selectedProject?.id === identifier || state.selectedProject?.slug === identifier)
            ? { ...state.selectedProject, ...updates }
            : state.selectedProject
        })),

        removeProject: (identifier) => set((state) => ({
          projects: state.projects.filter(project => 
            project.id !== identifier && project.slug !== identifier
          ),
          selectedProject: (state.selectedProject?.id === identifier || state.selectedProject?.slug === identifier) 
            ? null 
            : state.selectedProject
        })),

        setSelectedProject: (project) => set({ selectedProject: project }),

        findProjectById: (id) => {
          const { projects } = get()
          return projects.find(project => project.id === id)
        },

        findProjectByName: (name) => {
          const { projects } = get()
          return projects.find(project => project.title === name || project.slug === name)
        },

        setLoading: (loading) => set({ loading }),

        setError: (error) => set({ error }),

        clearProjects: () => set({
          projects: [],
          loading: false,
          error: null,
        }),

        clearSelectedProject: () => set({ selectedProject: null }),
      }),
      {
        name: 'modelcraft-project-storage',
        partialize: (state) => ({
          selectedProject: state.selectedProject,
        }),
      }
    ),
    {
      name: 'project-store',
    }
  )
)
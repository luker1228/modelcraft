import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import type { Project, DatabaseCluster } from '@/types'

interface AppState {
  // 全局状态
  loading: boolean
  selectedProject: Project | null
  selectedCluster: DatabaseCluster | null
  selectedDatabase: string | null
  sidebarCollapsed: boolean

  // Actions
  setLoading: (loading: boolean) => void
  setSelectedProject: (project: Project | null) => void
  setSelectedCluster: (cluster: DatabaseCluster | null) => void
  setSelectedDatabase: (database: string | null) => void
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  clearSelection: () => void
}

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set) => ({
        // 初始状态
        loading: false,
        selectedProject: null,
        selectedCluster: null,
        selectedDatabase: null,
        sidebarCollapsed: false,

        // Actions
        setLoading: (loading) => set({ loading }),
        
        setSelectedProject: (project) => {
          set({ selectedProject: project })
          // 当切换项目时，清除集群和数据库选择
          if (project === null) {
            set({ selectedCluster: null, selectedDatabase: null })
          }
        },
        
        setSelectedCluster: (cluster) => {
          set({ selectedCluster: cluster })
          // 当切换集群时，清除数据库选择
          if (cluster === null) {
            set({ selectedDatabase: null })
          }
        },
        
        setSelectedDatabase: (database) => set({ selectedDatabase: database }),
        
        toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
        
        setSidebarCollapsed: (collapsed: boolean) => set({ sidebarCollapsed: collapsed }),
        
        clearSelection: () => set({
          selectedProject: null,
          selectedCluster: null,
          selectedDatabase: null,
        }),
      }),
      {
        name: 'modelcraft-app-storage',
        partialize: (state) => ({
          selectedProject: state.selectedProject,
          selectedCluster: state.selectedCluster,
          selectedDatabase: state.selectedDatabase,
          sidebarCollapsed: state.sidebarCollapsed,
        }),
      }
    ),
    {
      name: 'app-store',
    }
  )
)

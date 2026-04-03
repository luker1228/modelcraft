import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type { DatabaseCluster } from '@/types'

interface ClusterState {
  // 状态
  clusters: DatabaseCluster[]
  selectedCluster: DatabaseCluster | null
  loading: boolean
  error: string | null

  // Actions
  setClusters: (clusters: DatabaseCluster[]) => void
  addCluster: (cluster: DatabaseCluster) => void
  updateCluster: (id: string, updates: Partial<DatabaseCluster>) => void
  removeCluster: (id: string) => void
  setSelectedCluster: (cluster: DatabaseCluster | null) => void
  findClusterById: (id: string) => DatabaseCluster | undefined
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearClusters: () => void
}

export const useClusterStore = create<ClusterState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      clusters: [],
      selectedCluster: null,
      loading: false,
      error: null,

      // Actions
      setClusters: (clusters) => set({ clusters }),

      addCluster: (cluster) => set((state) => ({
        clusters: [...state.clusters, cluster]
      })),

      updateCluster: (id, updates) => set((state) => ({
        clusters: state.clusters.map(cluster => 
          cluster.id === id ? { ...cluster, ...updates } : cluster
        ),
        selectedCluster: state.selectedCluster?.id === id 
          ? { ...state.selectedCluster, ...updates }
          : state.selectedCluster
      })),

      removeCluster: (id) => set((state) => ({
        clusters: state.clusters.filter(cluster => cluster.id !== id),
        selectedCluster: state.selectedCluster?.id === id ? null : state.selectedCluster
      })),

      setSelectedCluster: (cluster) => set({ selectedCluster: cluster }),

      findClusterById: (id) => {
        const { clusters } = get()
        return clusters.find(cluster => cluster.id === id)
      },

      setLoading: (loading) => set({ loading }),

      setError: (error) => set({ error }),

      clearClusters: () => set({
        clusters: [],
        selectedCluster: null,
        loading: false,
        error: null,
      }),
    }),
    {
      name: 'cluster-store',
    }
  )
)

import { writable, derived } from 'svelte/store'

// Current repository
export const currentRepository = writable(null)

// Repository status
export const repositoryStatus = writable(null)

// Branches list
export const branches = writable([])

// Recent repositories list
export const recentRepositories = writable([])

// Loading states
export const isLoading = writable(false)

// Error state
export const error = writable(null)

// Derived store - current branch
export const currentBranch = derived(
  [currentRepository, branches],
  ([$repo, $branches]) => {
    if (!$repo || !$branches.length) return null
    return $branches.find((b) => b.isHead)
  }
)

// Actions
export const repositoryActions = {
  setCurrentRepository: (repo) => {
    currentRepository.set(repo)
  },

  setRepositoryStatus: (status) => {
    repositoryStatus.set(status)
  },

  setBranches: (branchList) => {
    branches.set(branchList)
  },

  setRecentRepositories: (repos) => {
    recentRepositories.set(repos)
  },

  setLoading: (loading) => {
    isLoading.set(loading)
  },

  setError: (err) => {
    error.set(err)
  },

  clearError: () => {
    error.set(null)
  },

  reset: () => {
    currentRepository.set(null)
    repositoryStatus.set(null)
    branches.set([])
    error.set(null)
  }
}

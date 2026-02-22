import { writable } from 'svelte/store'

// Commits list
export const commits = writable([])

// Selected commit
export const selectedCommit = writable(null)

// Commit detail
export const commitDetail = writable(null)

// Loading state for commits
export const commitsLoading = writable(false)

// Refresh trigger - increment to trigger refresh
export const commitsRefreshTrigger = writable(0)

// Actions
export const commitActions = {
  setCommits: (commitList) => {
    commits.set(commitList)
  },

  setSelectedCommit: (commit) => {
    selectedCommit.set(commit)
  },

  setCommitDetail: (detail) => {
    commitDetail.set(detail)
  },

  setLoading: (loading) => {
    commitsLoading.set(loading)
  },

  triggerRefresh: () => {
    commitsRefreshTrigger.update(n => n + 1)
  },

  reset: () => {
    commits.set([])
    selectedCommit.set(null)
    commitDetail.set(null)
  }
}

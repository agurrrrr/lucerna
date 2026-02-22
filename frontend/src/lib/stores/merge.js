import { writable } from 'svelte/store'

// Merge state stores
export const mergeInProgress = writable(false)
export const conflictFiles = writable([])
export const currentConflictFile = writable(null)
export const threeWayDiff = writable(null)
export const mergeSourceBranch = writable('')
export const mergeTargetBranch = writable('')

// Actions for managing merge state
export const mergeActions = {
  setMergeInProgress: (inProgress) => {
    mergeInProgress.set(inProgress)
  },

  setConflictFiles: (files) => {
    conflictFiles.set(files)
  },

  updateConflictFile: (path, updates) => {
    conflictFiles.update(files =>
      files.map(f => f.path === path ? { ...f, ...updates } : f)
    )
  },

  setCurrentConflictFile: (file) => {
    currentConflictFile.set(file)
  },

  setThreeWayDiff: (diff) => {
    threeWayDiff.set(diff)
  },

  setMergeSourceBranch: (branch) => {
    mergeSourceBranch.set(branch)
  },

  setMergeTargetBranch: (branch) => {
    mergeTargetBranch.set(branch)
  },

  reset: () => {
    mergeInProgress.set(false)
    conflictFiles.set([])
    currentConflictFile.set(null)
    threeWayDiff.set(null)
    mergeSourceBranch.set('')
    mergeTargetBranch.set('')
  }
}

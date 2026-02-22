import { writable } from 'svelte/store'

// Current file diff
export const currentDiff = writable(null)

// Selected file path
export const selectedFile = writable(null)

// Diff loading state
export const diffLoading = writable(false)

// Actions
export const diffActions = {
  setCurrentDiff: (diff) => {
    currentDiff.set(diff)
  },

  setSelectedFile: (filePath) => {
    selectedFile.set(filePath)
  },

  setLoading: (loading) => {
    diffLoading.set(loading)
  },

  reset: () => {
    currentDiff.set(null)
    selectedFile.set(null)
    diffLoading.set(false)
  }
}

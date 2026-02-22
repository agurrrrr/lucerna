<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { GetFileHistory } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let filePath = ''

  let history = []
  let isLoading = false
  let error = null

  async function loadHistory() {
    if (!filePath) return

    try {
      isLoading = true
      error = null

      const historyJSON = await GetFileHistory(filePath, 50)
      history = JSON.parse(historyJSON)
    } catch (err) {
      console.error('Failed to load file history:', err)
      error = err.message || 'Failed to load file history'
    } finally {
      isLoading = false
    }
  }

  onMount(() => {
    loadHistory()
  })

  function formatDate(timestamp) {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
    return date.toLocaleDateString()
  }

  function getChangeColor(change) {
    switch (change) {
      case 'Added':
        return '#4caf50'
      case 'Modified':
        return '#ff9800'
      case 'Deleted':
        return '#f44336'
      default:
        return '#888'
    }
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="history-dialog">
  <div class="dialog-header">
    <h3>File History</h3>
    <div class="file-path">{filePath}</div>
  </div>

  <div class="dialog-body">
    {#if isLoading}
      <div class="loading">Loading history...</div>
    {:else if error}
      <div class="error-message">{error}</div>
    {:else if history.length === 0}
      <div class="empty-state">
        <p>No history found for this file</p>
      </div>
    {:else}
      <div class="history-list">
        {#each history as entry}
          <div class="history-item">
            <div class="history-header">
              <span class="commit-hash">{entry.hash.substring(0, 7)}</span>
              <span class="change-badge" style="color: {getChangeColor(entry.changes)}">
                {entry.changes}
              </span>
              <span class="commit-time">{formatDate(entry.timestamp)}</span>
            </div>
            <div class="commit-message">{entry.message.split('\n')[0]}</div>
            <div class="commit-author">
              {entry.author} &lt;{entry.email}&gt;
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div class="dialog-footer">
    <button class="btn-close" on:click={close}>Close</button>
  </div>
</div>

<style>
  .history-dialog {
    background: #1e2a38;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    width: 700px;
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  .dialog-header {
    padding: 1.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .dialog-header h3 {
    margin: 0 0 0.5rem 0;
    color: white;
    font-size: 1.25rem;
  }

  .file-path {
    font-family: monospace;
    font-size: 0.875rem;
    color: #888;
  }

  .dialog-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: #888;
  }

  .error-message {
    padding: 0.75rem 1rem;
    background: rgba(244, 67, 54, 0.1);
    border-left: 3px solid #f44336;
    color: #ffcdd2;
    font-size: 0.875rem;
    border-radius: 4px;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #666;
  }

  .empty-state p {
    margin: 0;
  }

  .history-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .history-item {
    padding: 1rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    border-left: 3px solid rgba(97, 218, 251, 0.3);
    transition: all 0.2s;
  }

  .history-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-left-color: #61dafb;
  }

  .history-header {
    display: flex;
    gap: 1rem;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .commit-hash {
    font-family: monospace;
    color: #61dafb;
    font-weight: 600;
  }

  .change-badge {
    padding: 0.125rem 0.5rem;
    border-radius: 3px;
    font-size: 0.75rem;
    font-weight: 600;
    background: rgba(255, 255, 255, 0.1);
  }

  .commit-time {
    font-size: 0.75rem;
    color: #888;
    margin-left: auto;
  }

  .commit-message {
    font-size: 0.875rem;
    color: #ddd;
    margin-bottom: 0.5rem;
  }

  .commit-author {
    font-size: 0.75rem;
    color: #888;
  }

  .dialog-footer {
    padding: 1.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    justify-content: flex-end;
  }

  .btn-close {
    padding: 0.75rem 2rem;
    background: rgba(255, 255, 255, 0.1);
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-close:hover {
    background: rgba(255, 255, 255, 0.15);
  }
</style>

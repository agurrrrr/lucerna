<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { GetFileBlame } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let filePath = ''

  let blameLines = []
  let isLoading = false
  let error = null

  async function loadBlame() {
    if (!filePath) return

    try {
      isLoading = true
      error = null

      const blameJSON = await GetFileBlame(filePath)
      blameLines = JSON.parse(blameJSON)
    } catch (err) {
      console.error('Failed to load blame:', err)
      error = err.message || 'Failed to load blame'
    } finally {
      isLoading = false
    }
  }

  onMount(() => {
    loadBlame()
  })

  function formatDate(timestamp) {
    const date = new Date(timestamp)
    return date.toLocaleDateString()
  }

  function getColorForHash(hash) {
    // Generate a consistent color based on hash
    let hashNum = 0
    for (let i = 0; i < Math.min(hash.length, 7); i++) {
      hashNum = hash.charCodeAt(i) + ((hashNum << 5) - hashNum)
    }
    const hue = Math.abs(hashNum % 360)
    return `hsl(${hue}, 50%, 40%)`
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="blame-dialog">
  <div class="dialog-header">
    <h3>Git Blame</h3>
    <div class="file-path">{filePath}</div>
  </div>

  <div class="dialog-body">
    {#if isLoading}
      <div class="loading">Loading blame...</div>
    {:else if error}
      <div class="error-message">{error}</div>
    {:else if blameLines.length === 0}
      <div class="empty-state">
        <p>No blame information available</p>
      </div>
    {:else}
      <div class="blame-container">
        <div class="blame-list">
          {#each blameLines as line}
            <div class="blame-line">
              <div class="blame-info" style="border-left-color: {getColorForHash(line.hash)}">
                <div class="blame-meta">
                  <span class="commit-hash" title={line.hash}>{line.hash.substring(0, 7)}</span>
                  <span class="author" title={line.email}>{line.author}</span>
                  <span class="date">{formatDate(line.timestamp)}</span>
                </div>
              </div>
              <div class="line-content">
                <span class="line-number">{line.lineNumber}</span>
                <pre class="line-text">{line.content}</pre>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <div class="dialog-footer">
    <button class="btn-close" on:click={close}>Close</button>
  </div>
</div>

<style>
  .blame-dialog {
    background: #1e2a38;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    width: 90vw;
    max-width: 1200px;
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
    margin: 1.5rem;
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

  .blame-container {
    overflow-x: auto;
  }

  .blame-list {
    font-family: monospace;
    font-size: 0.875rem;
  }

  .blame-line {
    display: flex;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    transition: background 0.2s;
  }

  .blame-line:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .blame-info {
    min-width: 350px;
    padding: 0.5rem 1rem;
    background: rgba(0, 0, 0, 0.2);
    border-left: 3px solid;
    display: flex;
    align-items: center;
  }

  .blame-meta {
    display: flex;
    gap: 1rem;
    align-items: center;
    font-size: 0.75rem;
  }

  .commit-hash {
    color: #61dafb;
    font-weight: 600;
    cursor: pointer;
  }

  .commit-hash:hover {
    text-decoration: underline;
  }

  .author {
    color: #aaa;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .date {
    color: #666;
    margin-left: auto;
  }

  .line-content {
    flex: 1;
    display: flex;
    padding: 0.5rem 0;
  }

  .line-number {
    display: inline-block;
    width: 50px;
    text-align: right;
    padding: 0 1rem;
    color: #666;
    user-select: none;
  }

  .line-text {
    flex: 1;
    margin: 0;
    color: #ddd;
    white-space: pre;
    overflow-x: auto;
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

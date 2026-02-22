<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { GetStashes, CreateStash, ApplyStash, DropStash } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let mode = 'list' // 'list' or 'create'

  let stashes = []
  let newStashMessage = ''
  let includeUntracked = false
  let isProcessing = false
  let error = null

  async function loadStashes() {
    try {
      const stashesJSON = await GetStashes()
      stashes = JSON.parse(stashesJSON)
    } catch (err) {
      console.error('Failed to load stashes:', err)
      error = err.message || 'Failed to load stashes'
    }
  }

  onMount(() => {
    if (mode === 'list') {
      loadStashes()
    }
  })

  async function handleCreateStash() {
    try {
      isProcessing = true
      error = null

      await CreateStash(newStashMessage, includeUntracked)

      dispatch('success', { action: 'create' })
      close()
    } catch (err) {
      error = err.message || 'Failed to create stash'
      console.error('Create stash failed:', err)
    } finally {
      isProcessing = false
    }
  }

  async function handleApplyStash(index, pop = false) {
    try {
      isProcessing = true
      error = null

      await ApplyStash(index, pop)

      dispatch('success', { action: pop ? 'pop' : 'apply', index })
      await loadStashes()
    } catch (err) {
      error = err.message || `Failed to ${pop ? 'pop' : 'apply'} stash`
      console.error('Apply stash failed:', err)
    } finally {
      isProcessing = false
    }
  }

  async function handleDropStash(index) {
    if (!confirm(`Are you sure you want to delete stash@{${index}}?`)) {
      return
    }

    try {
      isProcessing = true
      error = null

      await DropStash(index)

      await loadStashes()
    } catch (err) {
      error = err.message || 'Failed to drop stash'
      console.error('Drop stash failed:', err)
    } finally {
      isProcessing = false
    }
  }

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

  function close() {
    dispatch('cancel')
  }
</script>

<div class="stash-dialog">
  <div class="dialog-header">
    <h3>{mode === 'create' ? 'Create Stash' : 'Stashes'}</h3>
  </div>

  <div class="dialog-body">
    {#if mode === 'create'}
      <div class="create-stash">
        <div class="form-group">
          <label for="stashMessage">Stash Message (optional)</label>
          <input
            id="stashMessage"
            type="text"
            bind:value={newStashMessage}
            placeholder="WIP: feature implementation"
            disabled={isProcessing}
          />
        </div>

        <div class="form-group checkbox-group">
          <label>
            <input
              type="checkbox"
              bind:checked={includeUntracked}
              disabled={isProcessing}
            />
            Include untracked files
          </label>
        </div>

        <div class="info-box">
          <p>Stashing will save your local modifications and revert the working directory to match HEAD.</p>
        </div>
      </div>
    {:else}
      <div class="stash-list">
        {#if stashes.length === 0}
          <div class="empty-state">
            <p>No stashes found</p>
          </div>
        {:else}
          {#each stashes as stash}
            <div class="stash-item">
              <div class="stash-info">
                <div class="stash-header">
                  <span class="stash-index">stash@{`{${stash.index}}`}</span>
                  <span class="stash-branch">{stash.branch}</span>
                  <span class="stash-time">{formatDate(stash.timestamp)}</span>
                </div>
                <div class="stash-message">{stash.message}</div>
                <div class="stash-hash">{stash.hash.substring(0, 7)}</div>
              </div>
              <div class="stash-actions">
                <button
                  class="btn-action btn-apply"
                  on:click={() => handleApplyStash(stash.index, false)}
                  disabled={isProcessing}
                  title="Apply this stash"
                >
                  Apply
                </button>
                <button
                  class="btn-action btn-pop"
                  on:click={() => handleApplyStash(stash.index, true)}
                  disabled={isProcessing}
                  title="Pop this stash (apply and remove)"
                >
                  Pop
                </button>
                <button
                  class="btn-action btn-delete"
                  on:click={() => handleDropStash(stash.index)}
                  disabled={isProcessing}
                  title="Delete this stash"
                >
                  Delete
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    {/if}

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={close} disabled={isProcessing}>
      {mode === 'create' ? 'Cancel' : 'Close'}
    </button>
    {#if mode === 'create'}
      <button
        class="btn-submit"
        on:click={handleCreateStash}
        disabled={isProcessing}
      >
        {#if isProcessing}
          Creating...
        {:else}
          Create Stash
        {/if}
      </button>
    {/if}
  </div>
</div>

<style>
  .stash-dialog {
    background: var(--bg-secondary);
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    width: 500px;
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--border-color);
  }

  .dialog-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .dialog-header h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 600;
  }

  .dialog-body {
    padding: 16px;
    overflow-y: auto;
    flex: 1;
  }

  .form-group {
    margin-bottom: 12px;
  }

  label {
    display: block;
    margin-bottom: 6px;
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 600;
  }

  input[type='text'] {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: monospace;
    font-size: 12px;
  }

  input[type='text']:focus {
    outline: none;
    border-color: var(--accent);
  }

  input[type='text']:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .checkbox-group label {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-primary);
    cursor: pointer;
  }

  input[type='checkbox'] {
    width: auto;
    cursor: pointer;
  }

  .info-box {
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: 4px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .info-box p {
    margin: 0;
  }

  .stash-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .empty-state {
    text-align: center;
    padding: 24px;
    color: var(--text-muted);
  }

  .empty-state p {
    margin: 0;
    font-size: 12px;
  }

  .stash-item {
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    transition: background 0.15s;
  }

  .stash-item:hover {
    background: var(--bg-primary);
  }

  .stash-info {
    flex: 1;
    min-width: 0;
  }

  .stash-header {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 4px;
  }

  .stash-index {
    font-family: monospace;
    color: var(--accent);
    font-weight: 600;
    font-size: 12px;
  }

  .stash-branch {
    font-size: 11px;
    color: var(--success);
  }

  .stash-time {
    font-size: 10px;
    color: var(--text-muted);
  }

  .stash-message {
    font-size: 12px;
    color: var(--text-primary);
    margin-bottom: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .stash-hash {
    font-family: monospace;
    font-size: 10px;
    color: var(--text-muted);
  }

  .stash-actions {
    display: flex;
    gap: 4px;
  }

  .btn-action {
    padding: 4px 8px;
    border: none;
    border-radius: 3px;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-action:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-apply {
    background: rgba(88, 166, 255, 0.15);
    color: var(--accent);
  }

  .btn-apply:hover:not(:disabled) {
    background: rgba(88, 166, 255, 0.25);
  }

  .btn-pop {
    background: rgba(63, 185, 80, 0.15);
    color: var(--success);
  }

  .btn-pop:hover:not(:disabled) {
    background: rgba(63, 185, 80, 0.25);
  }

  .btn-delete {
    background: rgba(248, 81, 73, 0.15);
    color: var(--danger);
  }

  .btn-delete:hover:not(:disabled) {
    background: rgba(248, 81, 73, 0.25);
  }

  .error-message {
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-left: 2px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
    border-radius: 4px;
    margin-top: 12px;
  }

  .dialog-footer {
    padding: 12px 16px;
    border-top: 1px solid var(--border-color);
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  button {
    padding: 6px 16px;
    border-radius: 4px;
    font-weight: 600;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
    border: none;
  }

  .btn-cancel {
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .btn-cancel:hover:not(:disabled) {
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .btn-submit {
    background: var(--accent);
    color: var(--bg-primary);
  }

  .btn-submit:hover:not(:disabled) {
    background: var(--accent-hover);
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>

<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { GetTags, CreateTag, DeleteTag } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let mode = 'list' // 'list' or 'create'
  export let commitHash = '' // Optional: specific commit to tag

  let tags = []
  let tagName = ''
  let tagMessage = ''
  let isAnnotated = false
  let isProcessing = false
  let error = null

  async function loadTags() {
    try {
      const tagsJSON = await GetTags()
      tags = JSON.parse(tagsJSON)
      // Sort tags by timestamp descending
      tags.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
    } catch (err) {
      console.error('Failed to load tags:', err)
      error = err.message || 'Failed to load tags'
    }
  }

  onMount(() => {
    if (mode === 'list') {
      loadTags()
    }
  })

  async function handleCreateTag() {
    if (!tagName.trim()) {
      error = 'Tag name is required'
      return
    }

    try {
      isProcessing = true
      error = null

      const message = isAnnotated ? tagMessage : ''
      await CreateTag(tagName.trim(), message, commitHash)

      dispatch('success', { action: 'create', tagName: tagName.trim() })
      close()
    } catch (err) {
      error = err.message || 'Failed to create tag'
      console.error('Create tag failed:', err)
    } finally {
      isProcessing = false
    }
  }

  async function handleDeleteTag(name) {
    if (!confirm(`Are you sure you want to delete tag "${name}"?`)) {
      return
    }

    try {
      isProcessing = true
      error = null

      await DeleteTag(name)

      await loadTags()
    } catch (err) {
      error = err.message || 'Failed to delete tag'
      console.error('Delete tag failed:', err)
    } finally {
      isProcessing = false
    }
  }

  function formatDate(timestamp) {
    const date = new Date(timestamp)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="tag-dialog">
  <div class="dialog-header">
    <h3>{mode === 'create' ? 'Create Tag' : 'Tags'}</h3>
  </div>

  <div class="dialog-body">
    {#if mode === 'create'}
      <div class="create-tag">
        <div class="form-group">
          <label for="tagName">Tag Name</label>
          <input
            id="tagName"
            type="text"
            bind:value={tagName}
            placeholder="v1.0.0"
            disabled={isProcessing}
            autofocus
          />
        </div>

        <div class="form-group checkbox-group">
          <label>
            <input
              type="checkbox"
              bind:checked={isAnnotated}
              disabled={isProcessing}
            />
            Create annotated tag
          </label>
        </div>

        {#if isAnnotated}
          <div class="form-group">
            <label for="tagMessage">Tag Message</label>
            <textarea
              id="tagMessage"
              bind:value={tagMessage}
              placeholder="Release notes..."
              rows="4"
              disabled={isProcessing}
            ></textarea>
          </div>
        {/if}

        <div class="info-box">
          {#if commitHash}
            <p>This tag will be created at commit <strong>{commitHash.substring(0, 7)}</strong></p>
          {:else}
            <p>This tag will be created at <strong>HEAD</strong></p>
          {/if}
          {#if isAnnotated}
            <p>Annotated tags are recommended for releases.</p>
          {/if}
        </div>
      </div>
    {:else}
      <div class="tag-list">
        {#if tags.length === 0}
          <div class="empty-state">
            <p>No tags found</p>
          </div>
        {:else}
          {#each tags as tag}
            <div class="tag-item">
              <div class="tag-info">
                <div class="tag-header">
                  <span class="tag-name">{tag.name}</span>
                  {#if tag.isAnnotated}
                    <span class="tag-badge">Annotated</span>
                  {/if}
                  <span class="tag-hash">{tag.hash.substring(0, 7)}</span>
                </div>
                {#if tag.message}
                  <div class="tag-message">{tag.message}</div>
                {/if}
                <div class="tag-meta">
                  <span class="tag-tagger">{tag.tagger}</span>
                  <span class="tag-time">{formatDate(tag.timestamp)}</span>
                </div>
              </div>
              <div class="tag-actions">
                <button
                  class="btn-action btn-delete"
                  on:click={() => handleDeleteTag(tag.name)}
                  disabled={isProcessing}
                  title="Delete this tag"
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
        on:click={handleCreateTag}
        disabled={isProcessing || !tagName.trim()}
      >
        {#if isProcessing}
          Creating...
        {:else}
          Create Tag
        {/if}
      </button>
    {/if}
  </div>
</div>

<style>
  .tag-dialog {
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

  input[type='text'],
  textarea {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: monospace;
    font-size: 12px;
  }

  textarea {
    font-family: inherit;
    resize: vertical;
  }

  input[type='text']:focus,
  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  input[type='text']:disabled,
  textarea:disabled {
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
    margin: 4px 0;
  }

  .info-box p:first-child {
    margin-top: 0;
  }

  .info-box p:last-child {
    margin-bottom: 0;
  }

  .info-box strong {
    color: var(--accent);
    font-family: monospace;
  }

  .tag-list {
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

  .tag-item {
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    transition: background 0.15s;
  }

  .tag-item:hover {
    background: var(--bg-primary);
  }

  .tag-info {
    flex: 1;
    min-width: 0;
  }

  .tag-header {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 4px;
  }

  .tag-name {
    font-family: monospace;
    color: var(--warning);
    font-weight: 600;
    font-size: 12px;
  }

  .tag-badge {
    padding: 1px 6px;
    background: rgba(88, 166, 255, 0.15);
    color: var(--accent);
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
  }

  .tag-hash {
    font-family: monospace;
    font-size: 10px;
    color: var(--text-muted);
  }

  .tag-message {
    font-size: 12px;
    color: var(--text-primary);
    margin-bottom: 4px;
    white-space: pre-wrap;
  }

  .tag-meta {
    display: flex;
    gap: 8px;
    font-size: 10px;
    color: var(--text-muted);
  }

  .tag-actions {
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

<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { repositoryStatus } from '../stores/repository.js'
  import { CreateCommit, StageFile } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  // 체크된 파일 목록 (FileChanges에서 전달)
  export let filesToStage = []

  let message = ''
  let isCommitting = false
  let error = null
  let isStaging = false

  // 커밋할 파일 목록 계산
  $: stagedFiles = $repositoryStatus?.stagedFiles || []
  $: filesToCommitDisplay = filesToStage.length > 0
    ? filesToStage.map(f => f.path)
    : stagedFiles
  $: fileCount = filesToCommitDisplay.length
  $: canCommit = message.trim().length > 0 && fileCount > 0

  async function commit() {
    if (!canCommit) return

    try {
      isCommitting = true
      error = null

      // 체크된 파일이 있으면 먼저 stage
      if (filesToStage.length > 0) {
        isStaging = true
        for (const file of filesToStage) {
          if (file.status !== 'staged') {
            await StageFile(file.path)
          }
        }
        isStaging = false
      }

      // Create commit (author info will be taken from git config)
      const hash = await CreateCommit(message, '', '')
      console.log('Commit created:', hash)

      // Reset form
      message = ''

      // Notify parent
      console.log('Dispatching committed event')
      dispatch('committed', { hash })
    } catch (err) {
      error = err.message || 'Failed to create commit'
      console.error('Failed to commit:', err)
    } finally {
      isCommitting = false
      isStaging = false
    }
  }

  function cancel() {
    message = ''
    error = null
    dispatch('cancel')
  }
</script>

<div class="commit-dialog">
  <div class="dialog-header">
    <h3>Create Commit</h3>
    <div class="staged-info">
      {fileCount} file{fileCount !== 1 ? 's' : ''} to commit
    </div>
  </div>

  <div class="dialog-body">
    <textarea
      bind:value={message}
      placeholder="Commit message..."
      rows="6"
      disabled={isCommitting}
    ></textarea>

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}

    <div class="staged-files">
      <h4>Files to commit:</h4>
      <div class="file-list">
        {#if fileCount === 0}
          <div class="empty">No files selected</div>
        {:else}
          {#each filesToCommitDisplay as file}
            <div class="file-item">{file}</div>
          {/each}
        {/if}
      </div>
    </div>
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={cancel} disabled={isCommitting}>
      Cancel
    </button>
    <button
      class="btn-commit"
      on:click={commit}
      disabled={!canCommit || isCommitting}
    >
      {#if isStaging}
        Staging...
      {:else if isCommitting}
        Committing...
      {:else}
        Commit
      {/if}
    </button>
  </div>
</div>

<style>
  .commit-dialog {
    background: var(--bg-secondary);
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    width: 450px;
    max-width: 90vw;
    border: 1px solid var(--border-color);
  }

  .dialog-header {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .dialog-header h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 600;
  }

  .staged-info {
    font-size: 11px;
    color: var(--text-muted);
  }

  .dialog-body {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  textarea {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 12px;
    resize: vertical;
    min-height: 80px;
  }

  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  textarea::placeholder {
    color: var(--text-muted);
  }

  textarea:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .error-message {
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-left: 2px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
    border-radius: 4px;
  }

  .staged-files h4 {
    margin: 0 0 6px 0;
    font-size: 12px;
    color: var(--text-secondary);
    font-weight: 600;
  }

  .file-list {
    max-height: 120px;
    overflow-y: auto;
    background: var(--bg-tertiary);
    border-radius: 4px;
    padding: 6px;
  }

  .empty {
    text-align: center;
    color: var(--text-muted);
    padding: 12px;
    font-size: 12px;
  }

  .file-item {
    padding: 4px 6px;
    font-family: monospace;
    font-size: 11px;
    color: var(--success);
    border-bottom: 1px solid var(--border-color);
  }

  .file-item:last-child {
    border-bottom: none;
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

  .btn-commit {
    background: var(--success);
    color: var(--bg-primary);
  }

  .btn-commit:hover:not(:disabled) {
    background: #2ea043;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>

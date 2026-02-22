<script>
  import { createEventDispatcher } from 'svelte'
  import { CreateBranch } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let mode = 'create' // 'create' or 'rename'
  export let currentName = ''

  let branchName = currentName
  let isProcessing = false
  let error = null

  $: canSubmit = branchName.trim().length > 0 && branchName !== currentName

  async function handleSubmit() {
    if (!canSubmit) return

    try {
      isProcessing = true
      error = null

      await CreateBranch(branchName.trim())

      dispatch('success', { branchName: branchName.trim() })
      close()
    } catch (err) {
      error = err.message || 'Failed to create branch'
      console.error('Branch operation failed:', err)
    } finally {
      isProcessing = false
    }
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="branch-dialog">
  <div class="dialog-header">
    <h3>{mode === 'create' ? 'Create New Branch' : 'Rename Branch'}</h3>
  </div>

  <div class="dialog-body">
    <div class="form-group">
      <label for="branchName">Branch Name</label>
      <input
        id="branchName"
        type="text"
        bind:value={branchName}
        placeholder="feature/my-branch"
        disabled={isProcessing}
        on:keydown={(e) => e.key === 'Enter' && handleSubmit()}
        autofocus
      />
    </div>

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}

    <div class="help-text">
      Branch names can contain letters, numbers, hyphens, and slashes.
    </div>
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={close} disabled={isProcessing}>
      Cancel
    </button>
    <button
      class="btn-submit"
      on:click={handleSubmit}
      disabled={!canSubmit || isProcessing}
    >
      {#if isProcessing}
        {mode === 'create' ? 'Creating...' : 'Renaming...'}
      {:else}
        {mode === 'create' ? 'Create Branch' : 'Rename'}
      {/if}
    </button>
  </div>
</div>

<style>
  .branch-dialog {
    background: var(--bg-secondary);
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    width: 400px;
    max-width: 90vw;
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

  input {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: monospace;
    font-size: 12px;
  }

  input:focus {
    outline: none;
    border-color: var(--accent);
  }

  input:disabled {
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
    margin-bottom: 12px;
  }

  .help-text {
    font-size: 11px;
    color: var(--text-muted);
    font-style: italic;
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

<script>
  import { createEventDispatcher } from 'svelte'
  import { currentBranch } from '../stores/repository.js'
  import { MergeBranch } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let sourceBranch = ''

  let isMerging = false
  let error = null

  $: targetBranch = $currentBranch?.name || 'current branch'

  async function handleMerge() {
    if (!sourceBranch) return

    try {
      isMerging = true
      error = null

      await MergeBranch(sourceBranch)

      dispatch('success', { sourceBranch })
      close()
    } catch (err) {
      error = err.message || 'Failed to merge branch'
      console.error('Merge failed:', err)
    } finally {
      isMerging = false
    }
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="merge-dialog">
  <div class="dialog-header">
    <h3>Merge Branch</h3>
  </div>

  <div class="dialog-body">
    <div class="merge-info">
      <div class="merge-flow">
        <div class="branch-badge source">
          <span class="branch-icon">🌿</span>
          <span class="branch-label">{sourceBranch}</span>
        </div>
        <div class="arrow">→</div>
        <div class="branch-badge target">
          <span class="branch-icon">🌿</span>
          <span class="branch-label">{targetBranch}</span>
        </div>
      </div>
      <p class="merge-description">
        This will merge <strong>{sourceBranch}</strong> into <strong>{targetBranch}</strong>.
      </p>
    </div>

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}

    <div class="warning-message">
      ⚠️ Make sure you have committed all your changes before merging.
    </div>
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={close} disabled={isMerging}>
      Cancel
    </button>
    <button
      class="btn-merge"
      on:click={handleMerge}
      disabled={isMerging}
    >
      {#if isMerging}
        Merging...
      {:else}
        Merge
      {/if}
    </button>
  </div>
</div>

<style>
  .merge-dialog {
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
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .merge-info {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .merge-flow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 16px;
    background: var(--bg-tertiary);
    border-radius: 4px;
  }

  .branch-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: 4px;
    font-weight: 600;
    font-size: 12px;
  }

  .branch-badge.source {
    background: rgba(88, 166, 255, 0.15);
    color: var(--accent);
  }

  .branch-badge.target {
    background: rgba(63, 185, 80, 0.15);
    color: var(--success);
  }

  .branch-icon {
    font-size: 14px;
  }

  .branch-label {
    font-family: monospace;
  }

  .arrow {
    font-size: 16px;
    color: var(--text-muted);
  }

  .merge-description {
    text-align: center;
    color: var(--text-secondary);
    margin: 0;
    font-size: 12px;
  }

  .merge-description strong {
    color: var(--text-primary);
    font-family: monospace;
  }

  .error-message {
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-left: 2px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
    border-radius: 4px;
  }

  .warning-message {
    padding: 8px 12px;
    background: rgba(210, 153, 34, 0.1);
    border-left: 2px solid var(--warning);
    color: var(--warning);
    font-size: 12px;
    border-radius: 4px;
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

  .btn-merge {
    background: var(--success);
    color: var(--bg-primary);
  }

  .btn-merge:hover:not(:disabled) {
    background: #2ea043;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>

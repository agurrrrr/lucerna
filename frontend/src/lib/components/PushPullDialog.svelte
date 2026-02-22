<script>
  import { createEventDispatcher, onMount } from 'svelte'
  import { currentBranch, branches } from '../stores/repository.js'
  import {
    Push,
    Pull,
    Fetch,
    GetRemotes,
    GetRemoteURL,
    GetCredentials,
    PushWithAuth,
    PullWithAuth,
    FetchWithAuth
  } from '../../../wailsjs/go/main/App'
  import AuthDialog from './AuthDialog.svelte'

  const dispatch = createEventDispatcher()

  export let mode = 'push' // 'push', 'pull', or 'fetch'

  let remotes = []
  let selectedRemote = 'origin'
  let branchName = ''
  let force = false
  let isProcessing = false
  let error = null

  // Auth dialog state
  let showAuthDialog = false
  let authRemoteURL = ''

  // 로컬 브랜치 목록 (remote 제외)
  $: localBranches = $branches.filter(b => !b.isRemote)

  // 컴포넌트 마운트 시 현재 브랜치 설정
  onMount(() => {
    if ($currentBranch?.name) {
      branchName = $currentBranch.name
    } else if (localBranches.length > 0) {
      branchName = localBranches[0].name
    }
  })
  $: actionText = mode === 'push' ? 'Push' : mode === 'pull' ? 'Pull' : 'Fetch'
  $: currentRemoteURL = remotes.find(r => r.name === selectedRemote)?.urls[0] || ''

  async function loadRemotes() {
    try {
      const remotesJSON = await GetRemotes()
      remotes = JSON.parse(remotesJSON)
      if (remotes.length > 0 && !selectedRemote) {
        selectedRemote = remotes[0].name
      }
    } catch (err) {
      console.error('Failed to load remotes:', err)
      error = err.message || 'Failed to load remotes'
    }
  }

  // Load remotes on mount
  loadRemotes()

  // Check if error is auth-related
  function isAuthError(err) {
    // Wails에서 Go 에러는 문자열로 전달될 수 있음
    const errStr = typeof err === 'string' ? err : (err?.message || String(err))
    return errStr && errStr.includes('AUTH_REQUIRED:')
  }

  // Extract error message from various error formats
  function getErrorMessage(err) {
    if (typeof err === 'string') return err
    return err?.message || String(err)
  }

  // Try to get saved credentials and execute with auth
  async function tryWithSavedCredentials() {
    try {
      const remoteURL = await GetRemoteURL(selectedRemote)
      const credsJSON = await GetCredentials(remoteURL)

      if (credsJSON) {
        const creds = JSON.parse(credsJSON)
        return await executeWithAuth(creds.username, creds.password)
      }
      return false
    } catch (err) {
      console.error('Failed to get saved credentials:', err)
      return false
    }
  }

  // Execute the operation with authentication
  async function executeWithAuth(username, password) {
    try {
      if (mode === 'push') {
        await PushWithAuth(selectedRemote, branchName, force, username, password)
      } else if (mode === 'pull') {
        await PullWithAuth(selectedRemote, branchName, username, password)
      } else if (mode === 'fetch') {
        await FetchWithAuth(selectedRemote, username, password)
      }
      return true
    } catch (err) {
      // If still auth error, credentials are wrong
      if (isAuthError(err)) {
        return false
      }
      throw err
    }
  }

  async function handleSubmit() {
    if (!selectedRemote) {
      error = 'Please select a remote'
      return
    }

    try {
      isProcessing = true
      error = null

      // First, try without auth (might have system credentials)
      try {
        if (mode === 'push') {
          await Push(selectedRemote, branchName, force)
        } else if (mode === 'pull') {
          await Pull(selectedRemote, branchName)
        } else if (mode === 'fetch') {
          await Fetch(selectedRemote)
        }

        console.log('PushPullDialog: dispatching success')
        dispatch('success', { mode, remote: selectedRemote, branch: branchName })
        return
      } catch (err) {
        console.log('Git operation error:', err, typeof err)
        // Check if it's an auth error
        if (isAuthError(err)) {
          // Try saved credentials first
          const remoteURL = await GetRemoteURL(selectedRemote)
          const success = await tryWithSavedCredentials()

          if (success) {
            dispatch('success', { mode, remote: selectedRemote, branch: branchName })
            return
          }

          // Show auth dialog
          authRemoteURL = remoteURL
          showAuthDialog = true
          return
        }
        throw err
      }
    } catch (err) {
      error = getErrorMessage(err) || `Failed to ${mode}`
      console.error(`${mode} failed:`, err)
    } finally {
      isProcessing = false
    }
  }

  async function handleAuthSubmit(event) {
    const { username, password } = event.detail
    showAuthDialog = false

    try {
      isProcessing = true
      error = null

      const success = await executeWithAuth(username, password)

      if (success) {
        dispatch('success', { mode, remote: selectedRemote, branch: branchName })
      } else {
        error = 'Authentication failed. Please check your credentials.'
      }
    } catch (err) {
      error = getErrorMessage(err) || `Failed to ${mode}`
      console.error(`${mode} failed:`, err)
    } finally {
      isProcessing = false
    }
  }

  function handleAuthCancel() {
    showAuthDialog = false
  }

  function close() {
    dispatch('cancel')
  }
</script>

<div class="pushpull-dialog">
  <div class="dialog-header">
    <h3>{actionText} {mode !== 'fetch' ? 'Branch' : 'Updates'}</h3>
  </div>

  <div class="dialog-body">
    {#if remotes.length === 0}
      <div class="warning-message">
        No remotes configured. Add a remote first.
      </div>
    {:else}
      <div class="form-group">
        <label for="remote">Remote</label>
        <select
          id="remote"
          bind:value={selectedRemote}
          disabled={isProcessing}
        >
          {#each remotes as remote}
            <option value={remote.name}>{remote.name}</option>
          {/each}
        </select>
        {#if selectedRemote}
          <div class="remote-url">
            {remotes.find(r => r.name === selectedRemote)?.urls[0] || ''}
          </div>
        {/if}
      </div>

      {#if mode !== 'fetch'}
        <div class="form-group">
          <label for="branch">Branch</label>
          <select
            id="branch"
            bind:value={branchName}
            disabled={isProcessing}
          >
            {#each localBranches as branch}
              <option value={branch.name}>{branch.name}{branch.isHead ? ' (current)' : ''}</option>
            {/each}
          </select>
        </div>
      {/if}

      {#if mode === 'push'}
        <div class="form-group checkbox-group">
          <label>
            <input
              type="checkbox"
              bind:checked={force}
              disabled={isProcessing}
            />
            Force push
          </label>
          {#if force}
            <div class="warning-message">
              ⚠️ Force push will overwrite remote history. Use with caution!
            </div>
          {/if}
        </div>
      {/if}

      {#if error}
        <div class="error-message">
          {error}
        </div>
      {/if}

      <div class="info-box">
        {#if mode === 'push'}
          <p>This will push your local commits to <strong>{selectedRemote}/{branchName}</strong>.</p>
        {:else if mode === 'pull'}
          <p>This will fetch and merge changes from <strong>{selectedRemote}/{branchName}</strong>.</p>
        {:else}
          <p>This will fetch updates from <strong>{selectedRemote}</strong> without merging.</p>
        {/if}
      </div>
    {/if}
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={close} disabled={isProcessing}>
      Cancel
    </button>
    <button
      class="btn-submit"
      on:click={handleSubmit}
      disabled={isProcessing || remotes.length === 0}
    >
      {#if isProcessing}
        {actionText}ing...
      {:else}
        {actionText}
      {/if}
    </button>
  </div>
</div>

{#if showAuthDialog}
  <div class="auth-overlay" on:click={handleAuthCancel}>
    <div on:click|stopPropagation>
      <AuthDialog
        remoteURL={authRemoteURL}
        remoteName={selectedRemote}
        on:submit={handleAuthSubmit}
        on:cancel={handleAuthCancel}
      />
    </div>
  </div>
{/if}

<style>
  .auth-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1001;
  }

  .pushpull-dialog {
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

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label {
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 600;
  }

  select,
  input[type='text'] {
    width: 100%;
    padding: 6px 10px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: monospace;
    font-size: 12px;
  }

  select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%238b949e' d='M3 4.5L6 7.5L9 4.5'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 8px center;
    padding-right: 28px;
    cursor: pointer;
  }

  select option {
    background: var(--bg-secondary);
    color: var(--text-primary);
    padding: 8px;
  }

  select:focus,
  input[type='text']:focus {
    outline: none;
    border-color: var(--accent);
  }

  select:disabled,
  input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .remote-url {
    font-size: 11px;
    color: var(--text-muted);
    font-family: monospace;
    margin-top: 4px;
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

  .info-box strong {
    color: var(--accent);
    font-family: monospace;
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

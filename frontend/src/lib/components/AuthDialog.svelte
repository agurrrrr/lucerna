<script>
  import { createEventDispatcher } from 'svelte'
  import { SaveCredentials } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  export let remoteURL = ''
  export let remoteName = ''

  let username = ''
  let password = ''
  let saveCredentials = true
  let isSubmitting = false
  let error = null

  $: canSubmit = username.trim().length > 0 && password.length > 0

  async function handleSubmit() {
    if (!canSubmit) return

    try {
      isSubmitting = true
      error = null

      // Save credentials if requested
      if (saveCredentials && remoteURL) {
        await SaveCredentials(remoteURL, username, password)
      }

      dispatch('submit', {
        username: username.trim(),
        password,
        saveCredentials
      })
    } catch (err) {
      error = err.message || 'Failed to save credentials'
      console.error('Auth error:', err)
    } finally {
      isSubmitting = false
    }
  }

  function close() {
    dispatch('cancel')
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && canSubmit) {
      handleSubmit()
    }
  }
</script>

<div class="auth-dialog">
  <div class="dialog-header">
    <h3>Authentication Required</h3>
  </div>

  <div class="dialog-body">
    <div class="info-box">
      <p>Enter credentials for:</p>
      <p class="remote-info">
        {#if remoteName}
          <strong>{remoteName}</strong>
        {/if}
        <span class="url">{remoteURL}</span>
      </p>
    </div>

    <div class="form-group">
      <label for="username">Username</label>
      <input
        id="username"
        type="text"
        bind:value={username}
        placeholder="Username or email"
        disabled={isSubmitting}
        on:keydown={handleKeydown}
        autofocus
      />
    </div>

    <div class="form-group">
      <label for="password">Password / Token</label>
      <input
        id="password"
        type="password"
        bind:value={password}
        placeholder="Password or personal access token"
        disabled={isSubmitting}
        on:keydown={handleKeydown}
      />
    </div>

    <div class="form-group checkbox-group">
      <label>
        <input
          type="checkbox"
          bind:checked={saveCredentials}
          disabled={isSubmitting}
        />
        Save credentials for this repository
      </label>
    </div>

    {#if error}
      <div class="error-message">
        {error}
      </div>
    {/if}

    <div class="help-text">
      For GitHub, GitLab, or Bitbucket, use a personal access token instead of your password.
    </div>
  </div>

  <div class="dialog-footer">
    <button class="btn-cancel" on:click={close} disabled={isSubmitting}>
      Cancel
    </button>
    <button
      class="btn-submit"
      on:click={handleSubmit}
      disabled={!canSubmit || isSubmitting}
    >
      {#if isSubmitting}
        Authenticating...
      {:else}
        Authenticate
      {/if}
    </button>
  </div>
</div>

<style>
  .auth-dialog {
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

  .remote-info {
    margin-top: 6px !important;
  }

  .remote-info strong {
    color: var(--accent);
    font-family: monospace;
  }

  .url {
    display: block;
    font-family: monospace;
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 4px;
    word-break: break-all;
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

  input[type='text'],
  input[type='password'] {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-family: inherit;
    font-size: 12px;
  }

  input[type='text']:focus,
  input[type='password']:focus {
    outline: none;
    border-color: var(--accent);
  }

  input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .checkbox-group label {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-primary);
    cursor: pointer;
    font-weight: normal;
  }

  input[type='checkbox'] {
    width: auto;
    cursor: pointer;
    accent-color: var(--accent);
  }

  .error-message {
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-left: 2px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
    border-radius: 4px;
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

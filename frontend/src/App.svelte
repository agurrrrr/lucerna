<script>
  import Sidebar from './lib/components/Sidebar.svelte'
  import RepositoryView from './lib/components/RepositoryView.svelte'
  import { error, isLoading } from './lib/stores/repository.js'
</script>

<main>
  <Sidebar />
  <div class="main-content">
    {#if $isLoading}
      <div class="loading-overlay">
        <span>Loading...</span>
      </div>
    {/if}

    {#if $error}
      <div class="error-banner">
        <span>{$error}</span>
        <button on:click={() => error.set(null)}>×</button>
      </div>
    {/if}

    <RepositoryView />
  </div>
</main>

<style>
  main {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
  }

  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    color: var(--text-primary);
    font-size: 14px;
  }

  .error-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-bottom: 1px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
  }

  .error-banner button {
    background: none;
    border: none;
    color: var(--danger);
    font-size: 16px;
    cursor: pointer;
    padding: 0 4px;
  }

  .error-banner button:hover {
    opacity: 0.7;
  }
</style>

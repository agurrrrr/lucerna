<script>
  import {
    currentRepository,
    recentRepositories,
    repositoryActions
  } from '../stores/repository.js'
  import {
    SelectRepositoryDirectory,
    OpenRepository,
    GetRecentRepositories,
    SaveRecentRepository,
    ReorderRecentRepositories
  } from '../../../wailsjs/go/main/App'

  let draggedIndex = null
  let dragOverIndex = null

  async function loadRecentRepositories() {
    try {
      const recentJSON = await GetRecentRepositories()
      const recent = JSON.parse(recentJSON)
      repositoryActions.setRecentRepositories(recent)
    } catch (err) {
      console.error('Failed to load recent repositories:', err)
    }
  }

  async function selectAndOpenRepository() {
    try {
      repositoryActions.setLoading(true)
      repositoryActions.clearError()

      const dir = await SelectRepositoryDirectory()
      if (!dir) {
        repositoryActions.setLoading(false)
        return
      }

      await openRepository(dir)
    } catch (err) {
      repositoryActions.setError(err.message)
      console.error('Failed to open repository:', err)
    } finally {
      repositoryActions.setLoading(false)
    }
  }

  async function openRepository(path) {
    if ($currentRepository?.path === path) {
      return
    }

    try {
      repositoryActions.setLoading(true)
      repositoryActions.clearError()

      const repoJSON = await OpenRepository(path)
      const repo = JSON.parse(repoJSON)

      repositoryActions.setCurrentRepository(repo)

      await SaveRecentRepository(repoJSON)
      await loadRecentRepositories()
    } catch (err) {
      repositoryActions.setError(err.message)
      console.error('Failed to open repository:', err)
    } finally {
      repositoryActions.setLoading(false)
    }
  }

  function handleDragStart(event, index) {
    draggedIndex = index
    event.dataTransfer.effectAllowed = 'move'
  }

  function handleDragOver(event, index) {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
    dragOverIndex = index
  }

  function handleDragLeave() {
    dragOverIndex = null
  }

  async function handleDrop(event, index) {
    event.preventDefault()
    if (draggedIndex === null || draggedIndex === index) {
      draggedIndex = null
      dragOverIndex = null
      return
    }

    const repos = [...$recentRepositories]
    const [removed] = repos.splice(draggedIndex, 1)
    repos.splice(index, 0, removed)

    repositoryActions.setRecentRepositories(repos)

    try {
      await ReorderRecentRepositories(JSON.stringify(repos))
    } catch (err) {
      console.error('Failed to save repository order:', err)
      await loadRecentRepositories()
    }

    draggedIndex = null
    dragOverIndex = null
  }

  function handleDragEnd() {
    draggedIndex = null
    dragOverIndex = null
  }

  loadRecentRepositories()
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <span class="logo">Lucerna</span>
    <button class="btn-open" on:click={selectAndOpenRepository} title="Open Repository">
      +
    </button>
  </div>

  <div class="repo-list">
    {#if $recentRepositories.length === 0}
      <div class="empty">No repositories</div>
    {:else}
      {#each $recentRepositories as repo, index (repo.path)}
        <button
          class="repo-item"
          class:active={$currentRepository?.path === repo.path}
          class:dragging={draggedIndex === index}
          class:drag-over={dragOverIndex === index && draggedIndex !== index}
          draggable="true"
          on:click={() => openRepository(repo.path)}
          on:dragstart={(e) => handleDragStart(e, index)}
          on:dragover={(e) => handleDragOver(e, index)}
          on:dragleave={handleDragLeave}
          on:drop={(e) => handleDrop(e, index)}
          on:dragend={handleDragEnd}
        >
          <span class="repo-name">{repo.name}</span>
        </button>
      {/each}
    {/if}
  </div>
</aside>

<style>
  .sidebar {
    width: 180px;
    height: 100vh;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
  }

  .logo {
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-open {
    width: 24px;
    height: 24px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .btn-open:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .repo-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }

  .empty {
    padding: 16px 12px;
    color: var(--text-muted);
    font-size: 12px;
  }

  .repo-item {
    width: 100%;
    display: block;
    padding: 8px 12px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    text-align: left;
    cursor: pointer;
    font-size: 13px;
    border-left: 2px solid transparent;
  }

  .repo-item:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .repo-item.active {
    background: var(--bg-tertiary);
    color: var(--accent);
    border-left-color: var(--accent);
  }

  .repo-item.dragging {
    opacity: 0.5;
  }

  .repo-item.drag-over {
    background: var(--bg-tertiary);
    border-top: 1px solid var(--accent);
  }

  .repo-name {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>

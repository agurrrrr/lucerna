<script>
  import { createEventDispatcher } from 'svelte'
  import { repositoryStatus, repositoryActions } from '../stores/repository.js'
  import { selectedFile, diffActions } from '../stores/diff.js'
  import {
    GetWorkingTreeDiff,
    StageFile,
    UnstageFile,
    StageAll,
    UnstageAll,
    GetRepositoryStatus
  } from '../../../wailsjs/go/main/App'
  import CommitDialog from './CommitDialog.svelte'
  import FileHistoryDialog from './FileHistoryDialog.svelte'
  import BlameViewDialog from './BlameViewDialog.svelte'

  const dispatch = createEventDispatcher()

  let showCommitDialog = false
  let showHistoryDialog = false
  let showBlameDialog = false
  let historyFilePath = ''
  let blameFilePath = ''
  let filesToCommit = []

  // 선택된 파일들 (체크박스용)
  let checkedFiles = new Set()

  function toggleFileCheck(filePath, event) {
    event.stopPropagation()
    if (checkedFiles.has(filePath)) {
      checkedFiles.delete(filePath)
    } else {
      checkedFiles.add(filePath)
    }
    checkedFiles = checkedFiles // trigger reactivity
  }

  function toggleAllFiles() {
    if (checkedFiles.size === changedFiles.length) {
      checkedFiles = new Set()
    } else {
      checkedFiles = new Set(changedFiles.map(f => f.path))
    }
  }

  async function stageSelectedFiles() {
    if (checkedFiles.size === 0) return
    try {
      for (const filePath of checkedFiles) {
        const file = changedFiles.find(f => f.path === filePath)
        if (file && file.status !== 'staged') {
          await StageFile(filePath)
        }
      }
      await refreshStatus()
      checkedFiles = new Set()
    } catch (err) {
      console.error('Failed to stage files:', err)
      repositoryActions.setError(err.message)
    }
  }

  async function unstageSelectedFiles() {
    if (checkedFiles.size === 0) return
    try {
      for (const filePath of checkedFiles) {
        const file = changedFiles.find(f => f.path === filePath)
        if (file && file.status === 'staged') {
          await UnstageFile(filePath)
        }
      }
      await refreshStatus()
      checkedFiles = new Set()
    } catch (err) {
      console.error('Failed to unstage files:', err)
      repositoryActions.setError(err.message)
    }
  }

  $: selectedStagedCount = [...checkedFiles].filter(path => {
    const file = changedFiles.find(f => f.path === path)
    return file && file.status === 'staged'
  }).length

  $: selectedUnstagedCount = [...checkedFiles].filter(path => {
    const file = changedFiles.find(f => f.path === path)
    return file && file.status !== 'staged'
  }).length

  $: allChecked = changedFiles.length > 0 && checkedFiles.size === changedFiles.length

  async function selectFile(filePath) {
    try {
      diffActions.setLoading(true)
      diffActions.setSelectedFile(filePath)

      const diffJSON = await GetWorkingTreeDiff(filePath)
      const diff = JSON.parse(diffJSON)
      diffActions.setCurrentDiff(diff)
    } catch (err) {
      console.error('Failed to load diff:', err)
    } finally {
      diffActions.setLoading(false)
    }
  }

  async function stageFile(filePath, event) {
    event.stopPropagation()
    try {
      await StageFile(filePath)
      await refreshStatus()
    } catch (err) {
      console.error('Failed to stage file:', err)
      repositoryActions.setError(err.message)
    }
  }

  async function unstageFile(filePath, event) {
    event.stopPropagation()
    try {
      await UnstageFile(filePath)
      await refreshStatus()
    } catch (err) {
      console.error('Failed to unstage file:', err)
      repositoryActions.setError(err.message)
    }
  }

  async function stageAllFiles() {
    try {
      await StageAll()
      await refreshStatus()
    } catch (err) {
      console.error('Failed to stage all:', err)
      repositoryActions.setError(err.message)
    }
  }

  async function unstageAllFiles() {
    try {
      await UnstageAll()
      await refreshStatus()
    } catch (err) {
      console.error('Failed to unstage all:', err)
      repositoryActions.setError(err.message)
    }
  }

  async function refreshStatus() {
    try {
      const statusJSON = await GetRepositoryStatus()
      const status = JSON.parse(statusJSON)
      repositoryActions.setRepositoryStatus(status)
    } catch (err) {
      console.error('Failed to refresh status:', err)
    }
  }

  async function handleCommitted() {
    console.log('FileChanges: handleCommitted called')
    showCommitDialog = false
    checkedFiles = new Set()
    filesToCommit = []
    console.log('FileChanges: calling refreshStatus')
    await refreshStatus()
    console.log('FileChanges: refreshStatus done, dispatching committed')
    dispatch('committed')
  }

  function openCommitDialog() {
    // 체크된 파일이 있으면 해당 파일들을, 없으면 이미 staged된 파일들을 커밋 대상으로 설정
    if (checkedFiles.size > 0) {
      filesToCommit = [...checkedFiles].map(path => {
        const file = changedFiles.find(f => f.path === path)
        return { path, status: file?.status || 'modified' }
      })
    } else {
      filesToCommit = []
    }
    showCommitDialog = true
  }

  function openHistory(filePath, event) {
    event.stopPropagation()
    historyFilePath = filePath
    showHistoryDialog = true
  }

  function closeHistory() {
    showHistoryDialog = false
    historyFilePath = ''
  }

  function openBlame(filePath, event) {
    event.stopPropagation()
    blameFilePath = filePath
    showBlameDialog = true
  }

  function closeBlame() {
    showBlameDialog = false
    blameFilePath = ''
  }

  function getAllChangedFiles() {
    if (!$repositoryStatus) return []

    const files = []

    if ($repositoryStatus.stagedFiles) {
      $repositoryStatus.stagedFiles.forEach((file) => {
        files.push({ path: file, status: 'staged' })
      })
    }

    if ($repositoryStatus.modifiedFiles) {
      $repositoryStatus.modifiedFiles.forEach((file) => {
        files.push({ path: file, status: 'modified' })
      })
    }

    if ($repositoryStatus.untrackedFiles) {
      $repositoryStatus.untrackedFiles.forEach((file) => {
        files.push({ path: file, status: 'untracked' })
      })
    }

    return files
  }

  function getStatusColor(status) {
    if (status === 'staged') return 'var(--success)'
    if (status === 'modified') return 'var(--warning)'
    return 'var(--accent)'
  }

  // $repositoryStatus를 명시적으로 참조하여 reactive하게 업데이트
  $: changedFiles = $repositoryStatus ? getAllChangedFiles() : []
  $: stagedCount = $repositoryStatus?.stagedFiles?.length || 0
  $: unstagedCount = ($repositoryStatus?.modifiedFiles?.length || 0) + ($repositoryStatus?.untrackedFiles?.length || 0)
</script>

<div class="file-changes">
  <div class="header">
    <div class="header-left">
      {#if changedFiles.length > 0}
        <input
          type="checkbox"
          checked={allChecked}
          on:change={toggleAllFiles}
          class="check-all"
          title="Select all"
        />
      {/if}
      <span>Changes ({changedFiles.length})</span>
    </div>
    <div class="actions">
      {#if checkedFiles.size > 0}
        {#if selectedUnstagedCount > 0}
          <button class="btn-icon stage" on:click={stageSelectedFiles} title="Stage selected">+{selectedUnstagedCount}</button>
        {/if}
        {#if selectedStagedCount > 0}
          <button class="btn-icon unstage" on:click={unstageSelectedFiles} title="Unstage selected">-{selectedStagedCount}</button>
        {/if}
        <button class="btn-commit" on:click={openCommitDialog}>Commit</button>
      {:else}
        {#if unstagedCount > 0}
          <button class="btn-icon" on:click={stageAllFiles} title="Stage all">+</button>
        {/if}
        {#if stagedCount > 0}
          <button class="btn-icon" on:click={unstageAllFiles} title="Unstage all">-</button>
          <button class="btn-commit" on:click={openCommitDialog}>Commit</button>
        {/if}
      {/if}
    </div>
  </div>

  <div class="file-list">
    {#if changedFiles.length === 0}
      <div class="empty">No changes</div>
    {:else}
      {#each changedFiles as file}
        <div class="file-row" class:checked={checkedFiles.has(file.path)}>
          <input
            type="checkbox"
            checked={checkedFiles.has(file.path)}
            on:change={(e) => toggleFileCheck(file.path, e)}
            class="file-checkbox"
          />
          <button
            class="file-item"
            class:selected={$selectedFile === file.path}
            on:click={() => selectFile(file.path)}
          >
            <span class="status" style="color: {getStatusColor(file.status)}">{file.status[0].toUpperCase()}</span>
            <span class="path">{file.path}</span>
          </button>
          <div class="file-actions">
            <button class="btn-sm" on:click={(e) => openHistory(file.path, e)} title="History">H</button>
            <button class="btn-sm" on:click={(e) => openBlame(file.path, e)} title="Blame">B</button>
            {#if file.status === 'staged'}
              <button class="btn-sm unstage" on:click={(e) => unstageFile(file.path, e)} title="Unstage">-</button>
            {:else}
              <button class="btn-sm stage" on:click={(e) => stageFile(file.path, e)} title="Stage">+</button>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

{#if showCommitDialog}
  <div class="modal-overlay" on:click={() => (showCommitDialog = false)}>
    <div on:click|stopPropagation>
      <CommitDialog
        filesToStage={filesToCommit}
        on:committed={handleCommitted}
        on:cancel={() => (showCommitDialog = false)}
      />
    </div>
  </div>
{/if}

{#if showHistoryDialog}
  <div class="modal-overlay" on:click={closeHistory}>
    <div on:click|stopPropagation>
      <FileHistoryDialog filePath={historyFilePath} on:cancel={closeHistory} />
    </div>
  </div>
{/if}

{#if showBlameDialog}
  <div class="modal-overlay" on:click={closeBlame}>
    <div on:click|stopPropagation>
      <BlameViewDialog filePath={blameFilePath} on:cancel={closeBlame} />
    </div>
  </div>
{/if}

<style>
  .file-changes {
    width: 240px;
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    color: var(--text-primary);
    font-weight: 600;
    font-size: 12px;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .check-all {
    width: 14px;
    height: 14px;
    cursor: pointer;
    accent-color: var(--accent);
  }

  .actions {
    display: flex;
    gap: 4px;
  }

  .btn-icon {
    min-width: 24px;
    height: 24px;
    padding: 0 6px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 11px;
    font-weight: 600;
  }

  .btn-icon:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-icon.stage {
    color: var(--success);
    border-color: var(--success);
  }

  .btn-icon.stage:hover {
    background: rgba(63, 185, 80, 0.15);
  }

  .btn-icon.unstage {
    color: var(--danger);
    border-color: var(--danger);
  }

  .btn-icon.unstage:hover {
    background: rgba(248, 81, 73, 0.15);
  }

  .btn-commit {
    padding: 4px 10px;
    background: var(--accent);
    border: none;
    border-radius: 4px;
    color: var(--bg-primary);
    font-size: 12px;
    cursor: pointer;
  }

  .btn-commit:hover {
    background: var(--accent-hover);
  }

  .file-list {
    flex: 1;
    overflow-y: auto;
  }

  .empty {
    padding: 16px 12px;
    color: var(--text-muted);
    font-size: 12px;
  }

  .file-row {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border-color);
  }

  .file-row.checked {
    background: rgba(88, 166, 255, 0.08);
  }

  .file-checkbox {
    width: 14px;
    height: 14px;
    margin-left: 8px;
    cursor: pointer;
    accent-color: var(--accent);
    flex-shrink: 0;
  }

  .file-item {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    background: transparent;
    border: none;
    text-align: left;
    cursor: pointer;
    border-left: 2px solid transparent;
  }

  .file-item:hover {
    background: var(--bg-tertiary);
  }

  .file-item.selected {
    background: var(--bg-tertiary);
    border-left-color: var(--accent);
  }

  .status {
    font-size: 11px;
    font-weight: 600;
    width: 14px;
    text-align: center;
  }

  .path {
    flex: 1;
    font-size: 12px;
    color: var(--text-primary);
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .file-actions {
    display: flex;
    align-items: center;
    padding: 0 4px;
    gap: 2px;
  }

  .btn-sm {
    width: 22px;
    height: 22px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 11px;
  }

  .btn-sm:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-sm.stage:hover {
    color: var(--success);
    border-color: var(--success);
  }

  .btn-sm.unstage:hover {
    color: var(--danger);
    border-color: var(--danger);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
</style>

<script>
  import {
    currentRepository,
    repositoryStatus,
    branches,
    currentBranch,
    repositoryActions
  } from '../stores/repository.js'
  import { commitActions, selectedCommit } from '../stores/commits.js'
  import {
    mergeInProgress,
    conflictFiles,
    mergeSourceBranch,
    mergeTargetBranch,
    mergeActions
  } from '../stores/merge.js'
  import {
    GetRepositoryStatus,
    GetBranches,
    CheckoutBranch,
    DeleteBranch,
    CreateBranch,
    GetAheadBehind,
    MergeBranchWithResult,
    IsMergeInProgress,
    GetConflictFiles,
    GetMergeSourceBranch
  } from '../../../wailsjs/go/main/App'
  import CommitHistory from './CommitHistory.svelte'
  import CommitDetail from './CommitDetail.svelte'
  import FileChanges from './FileChanges.svelte'
  import DiffViewer from './DiffViewer.svelte'
  import BranchDialog from './BranchDialog.svelte'
  import MergeDialog from './MergeDialog.svelte'
  import PushPullDialog from './PushPullDialog.svelte'
  import StashDialog from './StashDialog.svelte'
  import TagDialog from './TagDialog.svelte'
  import MergeConflictEditor from './MergeConflictEditor.svelte'

  let activeTab = 'commits' // 'status', 'commits', 'branches'
  let showBranchDialog = false
  let branchDialogMode = 'create'
  let selectedBranchName = ''
  let showMergeDialog = false
  let mergeBranchName = ''
  let showPushPullDialog = false
  let pushPullMode = 'push'
  let showStashDialog = false
  let stashDialogMode = 'list'
  let showTagDialog = false
  let tagDialogMode = 'list'
  let branchError = null
  let processingBranch = null
  let aheadCount = 0
  let behindCount = 0
  let showMergeConflictEditor = false

  async function loadAheadBehind() {
    if (!$currentRepository) return
    try {
      const result = await GetAheadBehind('origin', '')
      const data = JSON.parse(result)
      aheadCount = data.ahead || 0
      behindCount = data.behind || 0
    } catch (err) {
      console.error('Failed to get ahead/behind:', err)
      aheadCount = 0
      behindCount = 0
    }
  }

  async function loadRepositoryData() {
    if (!$currentRepository) return

    try {
      // Load status
      const statusJSON = await GetRepositoryStatus()
      const status = JSON.parse(statusJSON)
      repositoryActions.setRepositoryStatus(status)

      // Load branches
      const branchesJSON = await GetBranches()
      const branchList = JSON.parse(branchesJSON)
      repositoryActions.setBranches(branchList)
    } catch (err) {
      console.error('Failed to load repository data:', err)
      repositoryActions.setError(err.message)
    }
  }

  // Check if there's an ongoing merge when repository loads
  async function checkMergeStatus() {
    if (!$currentRepository) return

    try {
      const inProgress = await IsMergeInProgress()
      if (inProgress) {
        // Restore merge state
        mergeActions.setMergeInProgress(true)

        const filesJSON = await GetConflictFiles()
        const files = JSON.parse(filesJSON)
        mergeActions.setConflictFiles(files.map(f => ({ ...f, resolved: false })))

        try {
          const sourceBranch = await GetMergeSourceBranch()
          mergeActions.setMergeSourceBranch(sourceBranch)
        } catch (e) {
          mergeActions.setMergeSourceBranch('unknown')
        }

        mergeActions.setMergeTargetBranch($currentBranch?.name || 'HEAD')
        showMergeConflictEditor = true
      }
    } catch (err) {
      console.error('Failed to check merge status:', err)
    }
  }

  // Handle merge from CommitHistory context menu
  async function handleMergeFromHistory(event) {
    const { branchName, sourceLabel, targetBranch } = event.detail
    await performMerge(branchName, targetBranch, sourceLabel)
  }

  // Perform merge operation
  async function performMerge(sourceBranch, targetBranch, sourceLabel) {
    const displayLabel = sourceLabel || `'${sourceBranch}'`
    try {
      const resultJSON = await MergeBranchWithResult(sourceBranch)
      const result = JSON.parse(resultJSON)

      if (result.success) {
        alert(`${displayLabel} 머지가 완료되었습니다.`)
        await loadRepositoryData()
        await loadAheadBehind()
        commitActions.triggerRefresh()
      } else if (result.hasConflicts) {
        // Conflicts detected - open merge conflict editor
        mergeActions.setMergeInProgress(true)
        mergeActions.setMergeSourceBranch(sourceBranch)
        mergeActions.setMergeTargetBranch(targetBranch || $currentBranch?.name || 'HEAD')
        mergeActions.setConflictFiles(result.conflictFiles.map(f => ({ ...f, resolved: false })))
        showMergeConflictEditor = true
      }
    } catch (err) {
      alert(`머지 실패: ${err.message || err}`)
      console.error('Merge failed:', err)
    }
  }

  // Handle merge conflict editor events
  function handleMergeComplete() {
    showMergeConflictEditor = false
    mergeActions.reset()
    loadRepositoryData()
    loadAheadBehind()
    commitActions.triggerRefresh()
  }

  function handleMergeAbort() {
    showMergeConflictEditor = false
    mergeActions.reset()
    loadRepositoryData()
  }

  // Load data when repository changes
  $: if ($currentRepository) {
    loadRepositoryData()
    loadAheadBehind()
    checkMergeStatus()
  }

  function openCreateBranchDialog() {
    branchDialogMode = 'create'
    selectedBranchName = ''
    branchError = null
    showBranchDialog = true
  }

  function closeBranchDialog() {
    showBranchDialog = false
    branchError = null
  }

  async function handleBranchCreated(event) {
    showBranchDialog = false
    await loadRepositoryData()
  }

  async function handleCheckout(branchName) {
    if (processingBranch) return

    try {
      processingBranch = branchName
      branchError = null
      await CheckoutBranch(branchName)
      await loadRepositoryData()
    } catch (err) {
      branchError = err.message || 'Failed to checkout branch'
      console.error('Checkout failed:', err)
    } finally {
      processingBranch = null
    }
  }

  async function handleDelete(branchName) {
    if (processingBranch) return

    if (!confirm(`Are you sure you want to delete branch "${branchName}"?`)) {
      return
    }

    try {
      processingBranch = branchName
      branchError = null
      await DeleteBranch(branchName)
      await loadRepositoryData()
    } catch (err) {
      branchError = err.message || 'Failed to delete branch'
      console.error('Delete failed:', err)
    } finally {
      processingBranch = null
    }
  }

  function openMergeDialog(branchName) {
    mergeBranchName = branchName
    branchError = null
    showMergeDialog = true
  }

  function closeMergeDialog() {
    showMergeDialog = false
    mergeBranchName = ''
    branchError = null
  }

  async function handleBranchMerged(event) {
    showMergeDialog = false
    await loadRepositoryData()
  }

  function openPushDialog() {
    pushPullMode = 'push'
    branchError = null
    showPushPullDialog = true
  }

  function openPullDialog() {
    pushPullMode = 'pull'
    branchError = null
    showPushPullDialog = true
  }

  function openFetchDialog() {
    pushPullMode = 'fetch'
    branchError = null
    showPushPullDialog = true
  }

  function closePushPullDialog() {
    showPushPullDialog = false
    branchError = null
  }

  async function handlePushPullSuccess(event) {
    console.log('handlePushPullSuccess called', event)
    showPushPullDialog = false
    console.log('Before refresh - aheadCount:', aheadCount, 'behindCount:', behindCount)
    await loadRepositoryData()
    await loadAheadBehind()
    commitActions.triggerRefresh()
    console.log('After refresh - aheadCount:', aheadCount, 'behindCount:', behindCount)
  }

  function openCreateStashDialog() {
    stashDialogMode = 'create'
    branchError = null
    showStashDialog = true
  }

  function openStashListDialog() {
    stashDialogMode = 'list'
    branchError = null
    showStashDialog = true
  }

  function closeStashDialog() {
    showStashDialog = false
    branchError = null
  }

  async function handleStashSuccess(event) {
    showStashDialog = false
    await loadRepositoryData()
  }

  function openCreateTagDialog() {
    tagDialogMode = 'create'
    branchError = null
    showTagDialog = true
  }

  function openTagListDialog() {
    tagDialogMode = 'list'
    branchError = null
    showTagDialog = true
  }

  function closeTagDialog() {
    showTagDialog = false
    branchError = null
  }

  async function handleTagSuccess(event) {
    showTagDialog = false
    await loadRepositoryData()
  }

  async function handleCommitted() {
    await loadRepositoryData()
    await loadAheadBehind()
    commitActions.triggerRefresh()
  }

  async function handleCheckoutCommit() {
    await loadRepositoryData()
    await loadAheadBehind()
  }
</script>

<div class="repository-view">
  {#if !$currentRepository}
    <div class="empty-state">
      <p>Select a repository</p>
    </div>
  {:else}
    <div class="repo-header">
      <div class="header-left">
        <span class="repo-name">{$currentRepository.name}</span>
        {#if $currentBranch}
          <span class="branch-badge">{$currentBranch.name}</span>
        {/if}
      </div>
      <div class="header-actions">
        <button class="btn-action" on:click={openFetchDialog}>Fetch</button>
        <button class="btn-action" on:click={openPullDialog}>
          Pull
          {#if behindCount > 0}
            <span class="badge pull">{behindCount}</span>
          {/if}
        </button>
        <button class="btn-action primary" on:click={openPushDialog}>
          Push
          {#if aheadCount > 0}
            <span class="badge push">{aheadCount}</span>
          {/if}
        </button>
        <span class="divider"></span>
        <button class="btn-action" on:click={openStashListDialog}>Stash</button>
        <button class="btn-action" on:click={openTagListDialog}>Tags</button>
      </div>
    </div>

    <div class="tabs">
      <button class="tab" class:active={activeTab === 'commits'} on:click={() => (activeTab = 'commits')}>
        Commits
      </button>
      <button class="tab" class:active={activeTab === 'status'} on:click={() => (activeTab = 'status')}>
        Changes
      </button>
      <button class="tab" class:active={activeTab === 'branches'} on:click={() => (activeTab = 'branches')}>
        Branches
      </button>
    </div>

    <div class="tab-content">
      {#if activeTab === 'commits'}
        <div class="commits-view" class:with-detail={$selectedCommit}>
          <CommitHistory limit={100} on:checkout={handleCheckoutCommit} on:merge={handleMergeFromHistory} />
          {#if $selectedCommit}
            <CommitDetail />
          {/if}
        </div>
      {:else if activeTab === 'status'}
        <div class="status-view">
          <FileChanges on:committed={handleCommitted} />
          <DiffViewer />
        </div>
      {:else if activeTab === 'branches'}
        <div class="branches-view">
          <div class="branches-header">
            <span>Branches ({$branches?.length || 0})</span>
            <button class="btn-new" on:click={openCreateBranchDialog}>New</button>
          </div>

          {#if branchError}
            <div class="error-msg">{branchError}</div>
          {/if}

          <div class="branch-list">
            {#if $branches && $branches.length > 0}
              {#each $branches as branch}
                <div class="branch-item" class:active={branch.isHead}>
                  <div class="branch-info">
                    <span class="branch-name">
                      {#if branch.isHead}<span class="check">*</span>{/if}
                      {branch.name}
                    </span>
                    <span class="branch-hash">{branch.hash.substring(0, 7)}</span>
                  </div>
                  {#if !branch.isHead}
                    <div class="branch-actions">
                      <button class="btn-sm" on:click={() => handleCheckout(branch.name)} disabled={processingBranch === branch.name}>
                        Checkout
                      </button>
                      <button class="btn-sm" on:click={() => openMergeDialog(branch.name)} disabled={processingBranch === branch.name}>
                        Merge
                      </button>
                      <button class="btn-sm danger" on:click={() => handleDelete(branch.name)} disabled={processingBranch === branch.name}>
                        Delete
                      </button>
                    </div>
                  {:else}
                    <span class="current-label">current</span>
                  {/if}
                </div>
              {/each}
            {:else}
              <div class="empty-list">No branches</div>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  {/if}

  {#if showBranchDialog}
    <div class="modal-overlay" on:click={closeBranchDialog}>
      <div on:click|stopPropagation>
        <BranchDialog
          mode={branchDialogMode}
          currentName={selectedBranchName}
          on:success={handleBranchCreated}
          on:cancel={closeBranchDialog}
        />
      </div>
    </div>
  {/if}

  {#if showMergeDialog}
    <div class="modal-overlay" on:click={closeMergeDialog}>
      <div on:click|stopPropagation>
        <MergeDialog
          sourceBranch={mergeBranchName}
          on:success={handleBranchMerged}
          on:cancel={closeMergeDialog}
        />
      </div>
    </div>
  {/if}

  {#if showPushPullDialog}
    <div class="modal-overlay" on:click={closePushPullDialog}>
      <div on:click|stopPropagation>
        <PushPullDialog
          mode={pushPullMode}
          on:success={handlePushPullSuccess}
          on:cancel={closePushPullDialog}
        />
      </div>
    </div>
  {/if}

  {#if showStashDialog}
    <div class="modal-overlay" on:click={closeStashDialog}>
      <div on:click|stopPropagation>
        <StashDialog
          mode={stashDialogMode}
          on:success={handleStashSuccess}
          on:cancel={closeStashDialog}
        />
      </div>
    </div>
  {/if}

  {#if showTagDialog}
    <div class="modal-overlay" on:click={closeTagDialog}>
      <div on:click|stopPropagation>
        <TagDialog
          mode={tagDialogMode}
          on:success={handleTagSuccess}
          on:cancel={closeTagDialog}
        />
      </div>
    </div>
  {/if}

  {#if showMergeConflictEditor}
    <div class="merge-editor-overlay">
      <MergeConflictEditor
        on:complete={handleMergeComplete}
        on:abort={handleMergeAbort}
      />
    </div>
  {/if}
</div>

<style>
  .repository-view {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
    overflow: hidden;
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .repo-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .repo-name {
    font-weight: 600;
    color: var(--text-primary);
  }

  .branch-badge {
    padding: 2px 8px;
    background: var(--bg-tertiary);
    border-radius: 12px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .divider {
    width: 1px;
    height: 16px;
    background: var(--border-color);
    margin: 0 4px;
  }

  .btn-action {
    padding: 4px 10px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
  }

  .btn-action:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-action.primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--bg-primary);
  }

  .btn-action.primary:hover {
    background: var(--accent-hover);
    border-color: var(--accent-hover);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    margin-left: 4px;
    border-radius: 8px;
    font-size: 10px;
    font-weight: 600;
  }

  .badge.pull {
    background: var(--warning);
    color: #000;
  }

  .badge.push {
    background: rgba(255, 255, 255, 0.9);
    color: var(--accent);
  }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .tab {
    padding: 8px 16px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
  }

  .tab:hover {
    color: var(--text-primary);
  }

  .tab.active {
    color: var(--accent);
    border-bottom-color: var(--accent);
  }

  .tab-content {
    flex: 1;
    overflow: hidden;
    display: flex;
  }

  .commits-view {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .commits-view:not(.with-detail) :global(.commit-history) {
    flex: 1;
    height: auto;
    border-bottom: none;
  }

  .status-view {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .branches-view {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
  }

  .branches-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    color: var(--text-primary);
    font-weight: 600;
  }

  .btn-new {
    padding: 4px 12px;
    background: var(--accent);
    border: none;
    border-radius: 4px;
    color: var(--bg-primary);
    font-size: 12px;
    cursor: pointer;
  }

  .btn-new:hover {
    background: var(--accent-hover);
  }

  .error-msg {
    padding: 8px 12px;
    background: rgba(248, 81, 73, 0.1);
    border-left: 2px solid var(--danger);
    color: var(--danger);
    font-size: 12px;
    margin-bottom: 12px;
  }

  .branch-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .branch-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: var(--bg-secondary);
    border-radius: 4px;
    border-left: 2px solid transparent;
  }

  .branch-item.active {
    border-left-color: var(--accent);
    background: var(--bg-tertiary);
  }

  .branch-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .branch-name {
    color: var(--text-primary);
  }

  .check {
    color: var(--accent);
    margin-right: 4px;
  }

  .branch-hash {
    font-family: monospace;
    font-size: 12px;
    color: var(--text-muted);
  }

  .branch-actions {
    display: flex;
    gap: 4px;
  }

  .btn-sm {
    padding: 3px 8px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-secondary);
    font-size: 11px;
    cursor: pointer;
  }

  .btn-sm:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-sm.danger:hover {
    background: rgba(248, 81, 73, 0.1);
    border-color: var(--danger);
    color: var(--danger);
  }

  .btn-sm:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .current-label {
    font-size: 11px;
    color: var(--text-muted);
  }

  .empty-list {
    padding: 24px;
    text-align: center;
    color: var(--text-muted);
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

  .merge-editor-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: var(--bg-primary);
    z-index: 1001;
    display: flex;
    flex-direction: column;
  }
</style>

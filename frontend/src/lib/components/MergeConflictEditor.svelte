<script>
  import { createEventDispatcher } from 'svelte'
  import {
    conflictFiles,
    currentConflictFile,
    threeWayDiff,
    mergeSourceBranch,
    mergeTargetBranch,
    mergeActions
  } from '../stores/merge.js'
  import {
    GetThreeWayDiff,
    ResolveConflictFile,
    CompleteMerge,
    AbortMerge,
    GetConflictFiles
  } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  let resolvedContent = ''
  let selectedFileIndex = 0
  let isLoading = false
  let isSaving = false

  // Select a conflict file and load its 3-way diff
  async function selectFile(file, index) {
    if (isLoading) return

    selectedFileIndex = index
    mergeActions.setCurrentConflictFile(file)

    try {
      isLoading = true
      const diffJSON = await GetThreeWayDiff(file.path)
      const diff = JSON.parse(diffJSON)
      mergeActions.setThreeWayDiff(diff)
      resolvedContent = diff.merged
    } catch (err) {
      console.error('Failed to load 3-way diff:', err)
      alert(`Failed to load diff: ${err.message || err}`)
    } finally {
      isLoading = false
    }
  }

  // Use ours (current branch) content
  function useOurs() {
    if ($threeWayDiff) {
      resolvedContent = $threeWayDiff.ours
    }
  }

  // Use theirs (source branch) content
  function useTheirs() {
    if ($threeWayDiff) {
      resolvedContent = $threeWayDiff.theirs
    }
  }

  // Use base (common ancestor) content
  function useBase() {
    if ($threeWayDiff) {
      resolvedContent = $threeWayDiff.base
    }
  }

  // Check if content has conflict markers
  function hasConflictMarkers(content) {
    if (!content) return false
    return content.includes('<<<<<<<') ||
           content.includes('=======') ||
           content.includes('>>>>>>>')
  }

  // Save resolution for current file
  async function saveResolution() {
    if (!$currentConflictFile || isSaving) return

    // Warn if conflict markers still exist
    if (hasConflictMarkers(resolvedContent)) {
      const confirmSave = confirm(
        '충돌 마커가 아직 남아있습니다. 그래도 저장하시겠습니까?'
      )
      if (!confirmSave) return
    }

    try {
      isSaving = true
      await ResolveConflictFile($currentConflictFile.path, resolvedContent)

      // Update file as resolved in store
      mergeActions.updateConflictFile($currentConflictFile.path, { resolved: true })

      // Refresh conflict files list
      const filesJSON = await GetConflictFiles()
      const files = JSON.parse(filesJSON)

      // Mark resolved files
      const updatedFiles = $conflictFiles.map(f => ({
        ...f,
        resolved: !files.find(cf => cf.path === f.path)
      }))
      mergeActions.setConflictFiles(updatedFiles)

      // Move to next unresolved file
      const nextUnresolved = updatedFiles.findIndex((f, i) => i > selectedFileIndex && !f.resolved)
      if (nextUnresolved !== -1) {
        await selectFile(updatedFiles[nextUnresolved], nextUnresolved)
      } else {
        // Check if all resolved
        const allResolved = updatedFiles.every(f => f.resolved)
        if (allResolved) {
          alert('모든 충돌이 해결되었습니다. Complete Merge 버튼을 클릭하세요.')
        }
      }
    } catch (err) {
      console.error('Failed to save resolution:', err)
      alert(`Failed to save: ${err.message || err}`)
    } finally {
      isSaving = false
    }
  }

  // Complete the merge
  async function completeMerge() {
    const unresolvedCount = $conflictFiles.filter(f => !f.resolved).length
    if (unresolvedCount > 0) {
      alert(`아직 ${unresolvedCount}개의 파일이 해결되지 않았습니다.`)
      return
    }

    const message = `Merge branch '${$mergeSourceBranch}' into ${$mergeTargetBranch}`
    const confirmMerge = confirm(`머지를 완료하시겠습니까?\n\n커밋 메시지:\n${message}`)
    if (!confirmMerge) return

    try {
      isLoading = true
      await CompleteMerge(message)
      dispatch('complete')
    } catch (err) {
      console.error('Failed to complete merge:', err)
      alert(`Merge failed: ${err.message || err}`)
    } finally {
      isLoading = false
    }
  }

  // Abort the merge
  async function abortMerge() {
    const confirmAbort = confirm(
      '머지를 취소하시겠습니까?\n\n모든 변경사항이 삭제됩니다.'
    )
    if (!confirmAbort) return

    try {
      isLoading = true
      await AbortMerge()
      dispatch('abort')
    } catch (err) {
      console.error('Failed to abort merge:', err)
      alert(`Abort failed: ${err.message || err}`)
    } finally {
      isLoading = false
    }
  }

  // Get file name from path
  function getFileName(path) {
    if (!path) return ''
    return path.split('/').pop()
  }

  // Count resolved and total files
  $: resolvedCount = $conflictFiles.filter(f => f.resolved).length
  $: totalCount = $conflictFiles.length

  // Auto-select first file when files are loaded
  $: if ($conflictFiles.length > 0 && !$currentConflictFile) {
    selectFile($conflictFiles[0], 0)
  }
</script>

<div class="merge-editor">
  <!-- Header -->
  <div class="editor-header">
    <div class="header-title">
      <span class="merge-icon">⎇</span>
      <span>Merge Conflicts</span>
      <span class="branch-info">
        '{$mergeSourceBranch}' → '{$mergeTargetBranch}'
      </span>
    </div>
    <div class="header-progress">
      <span class="progress-text">{resolvedCount} / {totalCount} resolved</span>
    </div>
    <div class="header-actions">
      <button class="btn btn-abort" on:click={abortMerge} disabled={isLoading}>
        Abort Merge
      </button>
      <button
        class="btn btn-complete"
        on:click={completeMerge}
        disabled={isLoading || resolvedCount < totalCount}
      >
        Complete Merge
      </button>
    </div>
  </div>

  <div class="editor-content">
    <!-- Left: File List -->
    <div class="file-list-panel">
      <div class="panel-header">
        <span>Conflict Files</span>
        <span class="file-count">({totalCount})</span>
      </div>
      <div class="file-list">
        {#each $conflictFiles as file, i}
          <button
            class="file-item"
            class:selected={i === selectedFileIndex}
            class:resolved={file.resolved}
            on:click={() => selectFile(file, i)}
          >
            <span class="file-status">
              {#if file.resolved}
                <span class="status-icon resolved">✓</span>
              {:else}
                <span class="status-icon conflict">!</span>
              {/if}
            </span>
            <span class="file-name" title={file.path}>{getFileName(file.path)}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- Right: Diff Panels -->
    <div class="diff-panels">
      {#if isLoading}
        <div class="loading-overlay">
          <span>Loading...</span>
        </div>
      {:else if $threeWayDiff}
        <!-- Top: Ours / Theirs -->
        <div class="top-panels">
          <div class="diff-panel ours">
            <div class="panel-header">
              <span class="panel-title">Ours ({$mergeTargetBranch})</span>
              <button class="btn-use" on:click={useOurs}>Use This</button>
            </div>
            <div class="diff-content">
              <pre>{$threeWayDiff.ours || '(empty)'}</pre>
            </div>
          </div>
          <div class="diff-panel theirs">
            <div class="panel-header">
              <span class="panel-title">Theirs ({$mergeSourceBranch})</span>
              <button class="btn-use" on:click={useTheirs}>Use This</button>
            </div>
            <div class="diff-content">
              <pre>{$threeWayDiff.theirs || '(empty)'}</pre>
            </div>
          </div>
        </div>

        <!-- Bottom: Result -->
        <div class="result-panel">
          <div class="panel-header">
            <span class="panel-title">Result</span>
            <div class="result-actions">
              {#if $threeWayDiff.base}
                <button class="btn-use small" on:click={useBase}>Use Base</button>
              {/if}
              <button
                class="btn-save"
                on:click={saveResolution}
                disabled={isSaving}
              >
                {isSaving ? 'Saving...' : 'Mark as Resolved'}
              </button>
            </div>
          </div>
          <div class="diff-content editable">
            <textarea
              bind:value={resolvedContent}
              class:has-conflicts={hasConflictMarkers(resolvedContent)}
              spellcheck="false"
            ></textarea>
          </div>
        </div>
      {:else}
        <div class="no-file-selected">
          <span>Select a file from the list to resolve conflicts</span>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .merge-editor {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
    background: var(--bg-primary);
  }

  .editor-header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .merge-icon {
    font-size: 16px;
  }

  .branch-info {
    color: var(--text-muted);
    font-size: 12px;
    font-weight: normal;
  }

  .header-progress {
    margin-left: auto;
  }

  .progress-text {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }

  .btn {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    font-size: 12px;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-abort {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
  }

  .btn-abort:hover:not(:disabled) {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-complete {
    background: #3fb950;
    color: #fff;
  }

  .btn-complete:hover:not(:disabled) {
    background: #2ea043;
  }

  .editor-content {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .file-list-panel {
    width: 200px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border-color);
  }

  .file-count {
    color: var(--text-muted);
  }

  .file-list {
    flex: 1;
    overflow-y: auto;
  }

  .file-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 12px;
    background: transparent;
    border: none;
    border-bottom: 1px solid var(--border-color);
    text-align: left;
    cursor: pointer;
    color: var(--text-primary);
    font-size: 12px;
  }

  .file-item:hover {
    background: var(--bg-tertiary);
  }

  .file-item.selected {
    background: rgba(88, 166, 255, 0.15);
  }

  .file-item.resolved {
    opacity: 0.7;
  }

  .file-status {
    flex-shrink: 0;
  }

  .status-icon {
    display: inline-block;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    text-align: center;
    line-height: 16px;
    font-size: 10px;
    font-weight: bold;
  }

  .status-icon.conflict {
    background: #f85149;
    color: #fff;
  }

  .status-icon.resolved {
    background: #3fb950;
    color: #fff;
  }

  .file-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .diff-panels {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
  }

  .loading-overlay,
  .no-file-selected {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
    font-size: 14px;
  }

  .top-panels {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  .diff-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .diff-panel.ours {
    border-right: 1px solid var(--border-color);
  }

  .diff-panel .panel-header {
    background: var(--bg-tertiary);
  }

  .panel-title {
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-use {
    padding: 2px 8px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    color: var(--text-secondary);
    font-size: 10px;
    cursor: pointer;
  }

  .btn-use:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-use.small {
    padding: 2px 6px;
    font-size: 10px;
  }

  .diff-content {
    flex: 1;
    overflow: auto;
    background: var(--bg-primary);
  }

  .diff-content pre {
    margin: 0;
    padding: 12px;
    font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
    font-size: 12px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-all;
    color: var(--text-primary);
  }

  .result-panel {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 200px;
    border-top: 2px solid var(--border-color);
  }

  .result-panel .panel-header {
    background: var(--bg-tertiary);
  }

  .result-actions {
    display: flex;
    gap: 8px;
  }

  .btn-save {
    padding: 4px 12px;
    background: #3fb950;
    border: none;
    border-radius: 3px;
    color: #fff;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-save:hover:not(:disabled) {
    background: #2ea043;
  }

  .btn-save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .diff-content.editable {
    display: flex;
  }

  .diff-content.editable textarea {
    flex: 1;
    width: 100%;
    padding: 12px;
    border: none;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
    font-size: 12px;
    line-height: 1.5;
    resize: none;
    outline: none;
  }

  .diff-content.editable textarea.has-conflicts {
    background: rgba(248, 81, 73, 0.1);
  }
</style>

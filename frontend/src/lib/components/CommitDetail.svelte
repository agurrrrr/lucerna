<script>
  import { selectedCommit, commitDetail, commitActions } from '../stores/commits.js'
  import { currentRepository } from '../stores/repository.js'
  import { GetCommitDetail, GetCommitFileDiff } from '../../../wailsjs/go/main/App'

  let loading = false
  let expandedFile = null
  let fileDiff = null
  let diffLoading = false
  let prevRepoPath = null

  async function loadCommitDetail(hash) {
    try {
      loading = true
      expandedFile = null
      fileDiff = null
      const detailJSON = await GetCommitDetail(hash)
      const detail = JSON.parse(detailJSON)
      commitActions.setCommitDetail(detail)
    } catch (err) {
      console.error('Failed to load commit detail:', err)
    } finally {
      loading = false
    }
  }

  async function toggleFileDiff(stat) {
    if (expandedFile === stat.path) {
      expandedFile = null
      fileDiff = null
      return
    }

    try {
      diffLoading = true
      expandedFile = stat.path
      const diffJSON = await GetCommitFileDiff($selectedCommit.hash, stat.path)
      fileDiff = JSON.parse(diffJSON)
    } catch (err) {
      console.error('Failed to load file diff:', err)
      fileDiff = null
    } finally {
      diffLoading = false
    }
  }

  $: if ($selectedCommit) {
    loadCommitDetail($selectedCommit.hash)
  }

  // 레포지토리 변경 시 상세창 닫기
  $: if ($currentRepository) {
    if (prevRepoPath !== null && prevRepoPath !== $currentRepository.path) {
      commitActions.setSelectedCommit(null)
      commitActions.setCommitDetail(null)
    }
    prevRepoPath = $currentRepository.path
  }

  function closeDetail() {
    commitActions.setSelectedCommit(null)
    commitActions.setCommitDetail(null)
  }

  function getStatusIcon(status) {
    switch (status) {
      case 'added': return '+'
      case 'modified': return '~'
      case 'deleted': return '-'
      default: return '?'
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case 'added': return 'var(--success)'
      case 'modified': return 'var(--warning)'
      case 'deleted': return 'var(--danger)'
      default: return 'var(--text-muted)'
    }
  }

  function getLineTypeClass(type) {
    switch (type) {
      case 'add': return 'line-add'
      case 'delete': return 'line-delete'
      default: return 'line-context'
    }
  }
</script>

<div class="commit-detail">
  {#if !$selectedCommit}
    <div class="empty">Select a commit</div>
  {:else if loading}
    <div class="loading">Loading...</div>
  {:else if $commitDetail}
    <div class="detail-header">
      <span class="detail-title">Commit Detail</span>
      <button class="btn-close" on:click={closeDetail}>×</button>
    </div>
    <div class="content">
      <div class="section">
        <div class="title">{$commitDetail.message.split('\n')[0]}</div>
        <div class="meta">
          <span class="hash">{$commitDetail.hash.substring(0, 7)}</span>
          <span class="sep">·</span>
          <span>{$commitDetail.author}</span>
          <span class="sep">·</span>
          <span>{new Date($commitDetail.timestamp).toLocaleString()}</span>
        </div>
      </div>

      {#if $commitDetail.message.includes('\n')}
        <div class="section">
          <pre class="description">{$commitDetail.message.split('\n').slice(1).join('\n').trim()}</pre>
        </div>
      {/if}

      <div class="section">
        <div class="section-title">Files ({$commitDetail.stats.length})</div>
        <div class="file-list">
          {#each $commitDetail.stats as stat}
            <div class="file-entry">
              <button
                class="file-item"
                class:expanded={expandedFile === stat.path}
                on:click={() => toggleFileDiff(stat)}
              >
                <span class="expand-icon">{expandedFile === stat.path ? '▼' : '▶'}</span>
                <span class="status" style="color: {getStatusColor(stat.status)}">{getStatusIcon(stat.status)}</span>
                <span class="path">{stat.path}</span>
                <span class="stats">
                  {#if stat.additions > 0}<span class="add">+{stat.additions}</span>{/if}
                  {#if stat.deletions > 0}<span class="del">-{stat.deletions}</span>{/if}
                </span>
              </button>

              {#if expandedFile === stat.path}
                <div class="file-diff">
                  {#if diffLoading}
                    <div class="diff-loading">Loading diff...</div>
                  {:else if fileDiff && fileDiff.hunks}
                    {#each fileDiff.hunks as hunk}
                      <div class="hunk">
                        <div class="hunk-header">{hunk.header}</div>
                        {#each hunk.lines as line}
                          <div class="diff-line {getLineTypeClass(line.type)}">
                            <span class="line-num old">{line.oldLine || ''}</span>
                            <span class="line-num new">{line.newLine || ''}</span>
                            <span class="line-prefix">{line.type === 'add' ? '+' : line.type === 'delete' ? '-' : ' '}</span>
                            <span class="line-content">{line.content}</span>
                          </div>
                        {/each}
                      </div>
                    {/each}
                  {:else}
                    <div class="diff-empty">No diff available</div>
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .commit-detail {
    flex: 1;
    overflow-y: auto;
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
  }

  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
    flex-shrink: 0;
  }

  .detail-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-close {
    padding: 2px 8px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-secondary);
    font-size: 14px;
    cursor: pointer;
    line-height: 1;
  }

  .btn-close:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .empty,
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
    font-size: 12px;
  }

  .content {
    padding: 12px;
    flex: 1;
    overflow-y: auto;
  }

  .section {
    margin-bottom: 16px;
  }

  .title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 8px;
  }

  .meta {
    font-size: 12px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .hash {
    font-family: monospace;
    color: var(--accent);
  }

  .sep {
    color: var(--text-muted);
  }

  .description {
    font-size: 12px;
    color: var(--text-secondary);
    background: var(--bg-secondary);
    padding: 8px 12px;
    border-radius: 4px;
    white-space: pre-wrap;
    margin: 0;
    font-family: inherit;
  }

  .section-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 8px;
  }

  .file-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .file-entry {
    display: flex;
    flex-direction: column;
  }

  .file-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    background: var(--bg-secondary);
    border-radius: 4px;
    font-size: 12px;
    border: none;
    width: 100%;
    text-align: left;
    cursor: pointer;
  }

  .file-item:hover {
    background: var(--bg-tertiary);
  }

  .file-item.expanded {
    border-radius: 4px 4px 0 0;
    background: var(--bg-tertiary);
  }

  .expand-icon {
    font-size: 8px;
    color: var(--text-muted);
    width: 10px;
  }

  .status {
    font-family: monospace;
    font-weight: 600;
    width: 12px;
  }

  .path {
    flex: 1;
    color: var(--text-primary);
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .stats {
    display: flex;
    gap: 8px;
    font-family: monospace;
    font-size: 11px;
  }

  .add {
    color: var(--success);
  }

  .del {
    color: var(--danger);
  }

  .file-diff {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-top: none;
    border-radius: 0 0 4px 4px;
    max-height: 400px;
    overflow-y: auto;
  }

  .diff-loading,
  .diff-empty {
    padding: 16px;
    text-align: center;
    color: var(--text-muted);
    font-size: 12px;
  }

  .hunk {
    border-bottom: 1px solid var(--border-color);
  }

  .hunk:last-child {
    border-bottom: none;
  }

  .hunk-header {
    padding: 4px 8px;
    background: var(--bg-secondary);
    color: var(--text-muted);
    font-family: monospace;
    font-size: 11px;
  }

  .diff-line {
    display: flex;
    font-family: monospace;
    font-size: 11px;
    line-height: 1.5;
  }

  .line-num {
    width: 40px;
    padding: 0 4px;
    text-align: right;
    color: var(--text-muted);
    background: var(--bg-secondary);
    user-select: none;
    flex-shrink: 0;
  }

  .line-prefix {
    width: 16px;
    text-align: center;
    flex-shrink: 0;
  }

  .line-content {
    flex: 1;
    padding-right: 8px;
    white-space: pre;
    overflow-x: auto;
  }

  .line-context {
    background: var(--bg-primary);
    color: var(--text-secondary);
  }

  .line-add {
    background: rgba(63, 185, 80, 0.15);
    color: var(--success);
  }

  .line-add .line-num {
    background: rgba(63, 185, 80, 0.2);
  }

  .line-delete {
    background: rgba(248, 81, 73, 0.15);
    color: var(--danger);
  }

  .line-delete .line-num {
    background: rgba(248, 81, 73, 0.2);
  }
</style>

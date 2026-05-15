<script>
  import { createEventDispatcher } from 'svelte'
  import {
    commits,
    selectedCommit,
    commitsLoading,
    commitsRefreshTrigger,
    commitActions
  } from '../stores/commits.js'
  import { currentRepository, currentBranch } from '../stores/repository.js'
  import { GetCommits, CheckoutCommit, CherryPickCommit, RevertCommit, CreateBranchFromCommit, CreateTag } from '../../../wailsjs/go/main/App'

  const dispatch = createEventDispatcher()

  // Get merge target for a commit. Returns null only for the current HEAD commit.
  // Prefers a non-current local branch label, then a remote branch label,
  // and falls back to the commit hash so any non-HEAD commit can be merged.
  function getMergeTarget(commit) {
    if (!commit) return null
    if ($currentBranch?.hash && commit.hash === $currentBranch.hash) return null

    const refs = commit.refs || []
    const localRef = refs.find(ref =>
      ref.type === 'branch' && !ref.isRemote && ref.name !== $currentBranch?.name
    )
    if (localRef) {
      return { name: localRef.name, ref: localRef.name, label: `'${localRef.name}'` }
    }

    const remoteRef = refs.find(ref => ref.type === 'branch' && ref.isRemote)
    if (remoteRef) {
      return { name: remoteRef.name, ref: remoteRef.name, label: `'${remoteRef.name}'` }
    }

    return { name: commit.shortHash || commit.hash, ref: commit.hash, label: commit.shortHash || commit.hash }
  }

  let checkoutInProgress = false

  // Context menu state
  let contextMenu = { show: false, x: 0, y: 0, commit: null }

  export let limit = 100

  const COLORS = [
    '#f0883e', // orange (main)
    '#58a6ff', // blue
    '#3fb950', // green
    '#a371f7', // purple
    '#f85149', // red
    '#d29922', // yellow
    '#f778ba', // pink
    '#79c0ff', // light blue
  ]

  const ROW_HEIGHT = 32
  const LANE_WIDTH = 12
  const NODE_RADIUS = 4
  const PADDING_LEFT = 8

  // Graph layout computation - Fork style
  function computeGraphLayout(commitList) {
    if (!commitList || commitList.length === 0) return { data: [], maxLanes: 0 }

    const result = []
    const hashToIndex = new Map()

    // Build index map
    commitList.forEach((commit, index) => {
      hashToIndex.set(commit.hash, index)
    })

    // activeLanes[i] = hash of the commit we're waiting for at lane i
    // null means the lane is free
    const activeLanes = []

    for (let i = 0; i < commitList.length; i++) {
      const commit = commitList[i]
      const parents = commit.parents || []

      // Step 1: Find my lane (was I expected somewhere?)
      // Check ALL lanes for this commit hash (could be in multiple due to merges)
      let myLane = -1
      let hasIncomingFromAbove = false
      const incomingLanes = []  // All lanes where this commit was expected

      for (let l = 0; l < activeLanes.length; l++) {
        if (activeLanes[l] === commit.hash) {
          incomingLanes.push(l)
          if (myLane === -1) {
            myLane = l  // Use the first (leftmost) lane found
          }
          hasIncomingFromAbove = true
        }
      }

      if (myLane === -1) {
        // Not expected - I'm a new branch head, take first free lane
        myLane = activeLanes.indexOf(null)
        if (myLane === -1) {
          myLane = activeLanes.length
        }
      }

      // Ensure activeLanes array is large enough
      while (activeLanes.length <= myLane) {
        activeLanes.push(null)
      }

      // Step 2: Capture pass-through lanes BEFORE any changes
      // Pass-through = lanes with commits we're still waiting for (not this commit, not null)
      const passThroughLanes = []
      for (let l = 0; l < activeLanes.length; l++) {
        if (l !== myLane && activeLanes[l] !== null && activeLanes[l] !== commit.hash) {
          passThroughLanes.push(l)
        }
      }

      // Step 3: Process this commit - clear my lane first
      // Make a copy of current state for sizing
      const currentLaneCount = Math.max(activeLanes.length, myLane + 1)

      // Clear all instances of this commit hash (it's been reached)
      for (let l = 0; l < activeLanes.length; l++) {
        if (activeLanes[l] === commit.hash) {
          activeLanes[l] = null
        }
      }

      // Step 4: Set up connections to parents
      const connections = []
      let hasOutgoingBelow = false

      if (parents.length > 0) {
        const firstParent = parents[0]
        const firstParentInList = hashToIndex.has(firstParent)

        if (firstParentInList) {
          // First parent continues in my lane
          activeLanes[myLane] = firstParent
          hasOutgoingBelow = true
          connections.push({
            fromLane: myLane,
            toLane: myLane,
            type: 'straight',
            color: COLORS[myLane % COLORS.length]
          })
        }

        // Handle merge parents (2nd, 3rd, etc.)
        for (let p = 1; p < parents.length; p++) {
          const parentHash = parents[p]
          if (!hashToIndex.has(parentHash)) continue

          // Is this parent already expected in some lane?
          let parentLane = activeLanes.indexOf(parentHash)

          if (parentLane === -1) {
            // Parent not yet tracked - find a lane for it
            // Prefer lanes to the right of myLane
            parentLane = -1
            for (let l = myLane + 1; l < activeLanes.length; l++) {
              if (activeLanes[l] === null) {
                parentLane = l
                break
              }
            }
            if (parentLane === -1) {
              // No free lane to the right, add new
              parentLane = activeLanes.length
            }

            // Ensure array is big enough
            while (activeLanes.length <= parentLane) {
              activeLanes.push(null)
            }
            activeLanes[parentLane] = parentHash
          }

          connections.push({
            fromLane: myLane,
            toLane: parentLane,
            type: 'merge',
            color: COLORS[parentLane % COLORS.length]
          })
        }
      }

      // Step 5: Clean up trailing nulls
      while (activeLanes.length > 0 && activeLanes[activeLanes.length - 1] === null) {
        activeLanes.pop()
      }

      result.push({
        commit,
        index: i,
        lane: myLane,
        laneCount: Math.max(currentLaneCount, activeLanes.length, myLane + 1),
        connections,
        passThroughLanes,
        incomingLanes,  // All lanes that have lines coming into this commit
        color: COLORS[myLane % COLORS.length],
        isMerge: parents.length > 1,
        hasIncomingFromAbove,
        hasOutgoingBelow
      })
    }

    const maxLanes = Math.max(...result.map(r => r.laneCount), 1)
    return { data: result, maxLanes }
  }

  function formatDateTime(timestamp) {
    const date = new Date(timestamp)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    const seconds = String(date.getSeconds()).padStart(2, '0')
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
  }

  function getInitials(name) {
    if (!name) return '?'
    const parts = name.trim().split(/\s+/)
    if (parts.length === 1) {
      return parts[0].charAt(0).toUpperCase()
    }
    return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase()
  }

  function getAvatarColor(name) {
    if (!name) return COLORS[0]
    let hash = 0
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    return COLORS[Math.abs(hash) % COLORS.length]
  }

  async function loadCommits() {
    if (!$currentRepository) return

    try {
      commitActions.setLoading(true)
      const commitsJSON = await GetCommits(limit)
      const commitList = JSON.parse(commitsJSON)
      commitActions.setCommits(commitList)
    } catch (err) {
      console.error('Failed to load commits:', err)
    } finally {
      commitActions.setLoading(false)
    }
  }

  function selectCommit(commit) {
    commitActions.setSelectedCommit(commit)
  }

  async function handleDoubleClick(commit) {
    if (checkoutInProgress) return

    const confirmMsg = `커밋 "${commit.shortHash}"로 체크아웃하시겠습니까?\n\n이 작업은 detached HEAD 상태가 됩니다.`
    if (!confirm(confirmMsg)) return

    try {
      checkoutInProgress = true
      await CheckoutCommit(commit.hash)
      await loadCommits()
      dispatch('checkout')
    } catch (err) {
      alert(`체크아웃 실패: ${err.message || err}`)
      console.error('Checkout failed:', err)
    } finally {
      checkoutInProgress = false
    }
  }

  // Context menu handlers
  function handleContextMenu(event, commit) {
    event.preventDefault()
    contextMenu = {
      show: true,
      x: event.clientX,
      y: event.clientY,
      commit
    }
  }

  function closeContextMenu() {
    contextMenu = { show: false, x: 0, y: 0, commit: null }
  }

  async function handleCheckoutCommit() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const confirmMsg = `커밋 "${commit.shortHash}"로 체크아웃하시겠습니까?\n\n이 작업은 detached HEAD 상태가 됩니다.`
    if (!confirm(confirmMsg)) return

    try {
      await CheckoutCommit(commit.hash)
      await loadCommits()
      dispatch('checkout')
    } catch (err) {
      alert(`체크아웃 실패: ${err.message || err}`)
    }
  }

  async function handleCherryPick() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const confirmMsg = `커밋 "${commit.shortHash}"을 체리픽하시겠습니까?`
    if (!confirm(confirmMsg)) return

    try {
      await CherryPickCommit(commit.hash)
      await loadCommits()
      alert('체리픽 완료')
    } catch (err) {
      alert(`체리픽 실패: ${err.message || err}`)
    }
  }

  async function handleRevert() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const confirmMsg = `커밋 "${commit.shortHash}"을 리버트하시겠습니까?`
    if (!confirm(confirmMsg)) return

    try {
      await RevertCommit(commit.hash)
      await loadCommits()
      alert('리버트 완료')
    } catch (err) {
      alert(`리버트 실패: ${err.message || err}`)
    }
  }

  async function handleNewBranch() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const branchName = prompt('새 브랜치 이름을 입력하세요:')
    if (!branchName) return

    try {
      await CreateBranchFromCommit(branchName, commit.hash)
      await loadCommits()
      alert(`브랜치 "${branchName}" 생성 완료`)
    } catch (err) {
      alert(`브랜치 생성 실패: ${err.message || err}`)
    }
  }

  async function handleNewTag() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const tagName = prompt('새 태그 이름을 입력하세요:')
    if (!tagName) return

    const tagMessage = prompt('태그 메시지를 입력하세요 (선택사항):', '')

    try {
      await CreateTag(tagName, tagMessage || '', commit.hash)
      await loadCommits()
      alert(`태그 "${tagName}" 생성 완료`)
    } catch (err) {
      alert(`태그 생성 실패: ${err.message || err}`)
    }
  }

  function handleMerge() {
    const commit = contextMenu.commit
    closeContextMenu()
    if (!commit) return

    const target = getMergeTarget(commit)
    if (!target) return

    // Dispatch merge event to parent (RepositoryView will handle it).
    // `branchName` is the committish passed to `git merge` — may be a branch
    // name, remote ref, or commit hash. `sourceLabel` is for display only.
    dispatch('merge', {
      branchName: target.ref,
      sourceLabel: target.label,
      targetBranch: $currentBranch?.name
    })
  }

  $: if ($currentRepository) {
    loadCommits()
  }

  // Refresh when triggered (after push/pull/fetch)
  $: if ($commitsRefreshTrigger) {
    loadCommits()
  }

  $: graphLayout = computeGraphLayout($commits)
  $: graphWidth = PADDING_LEFT + graphLayout.maxLanes * LANE_WIDTH + 8
</script>

<div class="commit-history">
  <div class="header">
    <span>History</span>
    <button class="btn-refresh" on:click={loadCommits} disabled={$commitsLoading}>
      {$commitsLoading ? '...' : '↻'}
    </button>
  </div>

  {#if $commitsLoading}
    <div class="loading">Loading...</div>
  {:else if graphLayout.data.length === 0}
    <div class="empty">No commits</div>
  {:else}
    <div class="commit-list">
      {#each graphLayout.data as data (data.commit.hash)}
        <button
          class="commit-row"
          class:selected={$selectedCommit?.hash === data.commit.hash}
          on:click={() => selectCommit(data.commit)}
          on:dblclick={() => handleDoubleClick(data.commit)}
          on:contextmenu={(e) => handleContextMenu(e, data.commit)}
        >
          <!-- Inline Graph Column -->
          <svg
            class="graph-cell"
            width={graphWidth}
            height={ROW_HEIGHT}
            viewBox="0 0 {graphWidth} {ROW_HEIGHT}"
          >
            <!-- Pass-through vertical lines (branches that continue through this row) -->
            {#each data.passThroughLanes as lane}
              <line
                x1={PADDING_LEFT + lane * LANE_WIDTH}
                y1="0"
                x2={PADDING_LEFT + lane * LANE_WIDTH}
                y2={ROW_HEIGHT}
                stroke={COLORS[lane % COLORS.length]}
                stroke-width="2"
              />
            {/each}

            <!-- Lines from TOP to node (from all incoming lanes) -->
            {#each data.incomingLanes as incomingLane}
              {#if incomingLane === data.lane}
                <!-- Straight line from top to node (same lane) -->
                <line
                  x1={PADDING_LEFT + incomingLane * LANE_WIDTH}
                  y1="0"
                  x2={PADDING_LEFT + data.lane * LANE_WIDTH}
                  y2={ROW_HEIGHT / 2 - NODE_RADIUS}
                  stroke={COLORS[incomingLane % COLORS.length]}
                  stroke-width="2"
                />
              {:else}
                <!-- Curved line from different lane to node (merge target) -->
                <path
                  d="M {PADDING_LEFT + incomingLane * LANE_WIDTH} 0
                     L {PADDING_LEFT + incomingLane * LANE_WIDTH} {ROW_HEIGHT * 0.15}
                     C {PADDING_LEFT + incomingLane * LANE_WIDTH} {ROW_HEIGHT * 0.4},
                       {PADDING_LEFT + data.lane * LANE_WIDTH} {ROW_HEIGHT * 0.4},
                       {PADDING_LEFT + data.lane * LANE_WIDTH} {ROW_HEIGHT / 2 - NODE_RADIUS}"
                  stroke={COLORS[incomingLane % COLORS.length]}
                  stroke-width="2"
                  fill="none"
                />
              {/if}
            {/each}

            <!-- Connections to next row (parents) - lines going DOWN -->
            {#each data.connections as conn}
              {#if conn.fromLane === conn.toLane}
                <!-- Straight vertical line down -->
                <line
                  x1={PADDING_LEFT + conn.fromLane * LANE_WIDTH}
                  y1={ROW_HEIGHT / 2 + NODE_RADIUS}
                  x2={PADDING_LEFT + conn.toLane * LANE_WIDTH}
                  y2={ROW_HEIGHT}
                  stroke={conn.color}
                  stroke-width="2"
                />
              {:else}
                <!-- Curved merge line -->
                <path
                  d="M {PADDING_LEFT + conn.fromLane * LANE_WIDTH} {ROW_HEIGHT / 2}
                     C {PADDING_LEFT + conn.fromLane * LANE_WIDTH} {ROW_HEIGHT * 0.85},
                       {PADDING_LEFT + conn.toLane * LANE_WIDTH} {ROW_HEIGHT * 0.85},
                       {PADDING_LEFT + conn.toLane * LANE_WIDTH} {ROW_HEIGHT}"
                  stroke={conn.color}
                  stroke-width="2"
                  fill="none"
                />
              {/if}
            {/each}

            <!-- Node -->
            {#if data.isMerge}
              <circle
                cx={PADDING_LEFT + data.lane * LANE_WIDTH}
                cy={ROW_HEIGHT / 2}
                r={NODE_RADIUS + 2}
                fill="var(--bg-secondary)"
                stroke={data.color}
                stroke-width="2"
              />
              <circle
                cx={PADDING_LEFT + data.lane * LANE_WIDTH}
                cy={ROW_HEIGHT / 2}
                r={NODE_RADIUS - 1}
                fill={data.color}
              />
            {:else}
              <circle
                cx={PADDING_LEFT + data.lane * LANE_WIDTH}
                cy={ROW_HEIGHT / 2}
                r={NODE_RADIUS}
                fill={data.color}
              />
            {/if}
          </svg>

          <!-- Commit Info Column -->
          <div class="commit-info">
            <div class="commit-title">
              {#if data.commit.refs && data.commit.refs.length > 0}
                <span class="refs">
                  {#each data.commit.refs as ref}
                    <span class="ref-label {ref.type}" class:remote={ref.isRemote}>
                      {ref.name}
                    </span>
                  {/each}
                </span>
              {/if}
              <span class="commit-message">{data.commit.message.split('\n')[0]}</span>
            </div>
          </div>

          <!-- Commit Meta Column (Right Side) -->
          <div class="commit-meta-right">
            <div class="author-info">
              <span class="avatar" style="background-color: {getAvatarColor(data.commit.author)}">
                {getInitials(data.commit.author)}
              </span>
              <span class="author-name">{data.commit.author}</span>
            </div>
            <span class="hash">{data.commit.shortHash}</span>
            <span class="datetime">{formatDateTime(data.commit.timestamp)}</span>
          </div>
        </button>
      {/each}
    </div>
  {/if}
</div>

<!-- Context Menu -->
{#if contextMenu.show}
  <div class="context-menu-overlay" on:click={closeContextMenu}></div>
  <div class="context-menu" style="left: {contextMenu.x}px; top: {contextMenu.y}px;">
    <button class="context-menu-item" on:click={handleCheckoutCommit}>
      <span class="icon">⎇</span>
      Checkout Commit
    </button>
    <button class="context-menu-item" on:click={handleCherryPick}>
      <span class="icon">🍒</span>
      Cherry-pick Commit
    </button>
    <button class="context-menu-item" on:click={handleRevert}>
      <span class="icon">↩</span>
      Revert Commit
    </button>
    <div class="context-menu-divider"></div>
    <button class="context-menu-item" on:click={handleNewBranch}>
      <span class="icon">⑂</span>
      New Branch...
    </button>
    <button class="context-menu-item" on:click={handleNewTag}>
      <span class="icon">🏷</span>
      New Tag...
    </button>
    {#if getMergeTarget(contextMenu.commit)}
      <div class="context-menu-divider"></div>
      <button class="context-menu-item merge-item" on:click={handleMerge}>
        <span class="icon">⎇</span>
        Merge {getMergeTarget(contextMenu.commit).label} into '{$currentBranch?.name}'
      </button>
    {/if}
  </div>
{/if}

<style>
  .commit-history {
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    height: 50%;
    min-height: 200px;
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
    flex-shrink: 0;
  }

  .btn-refresh {
    padding: 4px 8px;
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 12px;
  }

  .btn-refresh:hover:not(:disabled) {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-refresh:disabled {
    opacity: 0.5;
  }

  .loading,
  .empty {
    padding: 16px 12px;
    color: var(--text-muted);
    font-size: 12px;
  }

  .commit-list {
    flex: 1;
    overflow-y: auto;
  }

  .commit-row {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 0;
    padding-right: 12px;
    background: transparent;
    border: none;
    border-bottom: 1px solid var(--border-color);
    text-align: left;
    cursor: pointer;
    height: 32px;
  }

  .commit-row:hover {
    background: var(--bg-tertiary);
  }

  .commit-row.selected {
    background: rgba(88, 166, 255, 0.1);
  }

  .graph-cell {
    flex-shrink: 0;
    display: block;
  }

  .commit-info {
    flex: 1;
    min-width: 0;
    padding-left: 8px;
    display: flex;
    align-items: center;
  }

  .commit-title {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
  }

  .refs {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  .ref-label {
    display: inline-flex;
    align-items: center;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 600;
    white-space: nowrap;
  }

  .ref-label.head {
    background: #f85149;
    color: #fff;
  }

  .ref-label.branch {
    background: #3fb950;
    color: #fff;
  }

  .ref-label.branch.remote {
    background: #a371f7;
    color: #fff;
  }

  .ref-label.remote {
    background: #6e7681;
    color: #fff;
  }

  .ref-label.tag {
    background: #d29922;
    color: #000;
  }

  .commit-message {
    font-size: 12px;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 1.2;
  }

  .commit-meta-right {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-shrink: 0;
    padding-left: 12px;
    margin-left: auto;
  }

  .author-info {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 140px;
  }

  .avatar {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 9px;
    font-weight: 600;
    color: #fff;
    flex-shrink: 0;
  }

  .author-name {
    font-size: 11px;
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .hash {
    font-family: monospace;
    font-size: 11px;
    color: var(--accent);
    width: 60px;
    text-align: center;
  }

  .datetime {
    font-family: monospace;
    font-size: 11px;
    color: var(--text-muted);
    width: 145px;
    text-align: right;
  }

  /* Context Menu Styles */
  .context-menu-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 999;
  }

  .context-menu {
    position: fixed;
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    padding: 4px 0;
    min-width: 180px;
    z-index: 1000;
  }

  .context-menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 12px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 12px;
    text-align: left;
    cursor: pointer;
  }

  .context-menu-item:hover {
    background: var(--bg-tertiary);
  }

  .context-menu-item .icon {
    width: 16px;
    text-align: center;
    font-size: 12px;
  }

  .context-menu-divider {
    height: 1px;
    background: var(--border-color);
    margin: 4px 0;
  }

  .context-menu-item.merge-item {
    color: var(--accent);
  }
</style>

<script>
  import { currentDiff, diffLoading } from '../stores/diff.js'

  function formatLineNumber(num) {
    return num > 0 ? num.toString() : ''
  }

  function groupLines(lines) {
    const groups = []
    let i = 0

    while (i < lines.length) {
      const line = lines[i]

      if (line.type === 'context') {
        groups.push({ type: 'context', lines: [line] })
        i++
      } else if (line.type === 'add') {
        const adds = []
        while (i < lines.length && lines[i].type === 'add') {
          adds.push(lines[i])
          i++
        }
        const deletes = []
        while (i < lines.length && lines[i].type === 'delete') {
          deletes.push(lines[i])
          i++
        }

        if (deletes.length > 0) {
          groups.push({ type: 'change', adds, deletes })
        } else {
          groups.push({ type: 'adds', lines: adds })
        }
      } else if (line.type === 'delete') {
        const deletes = []
        while (i < lines.length && lines[i].type === 'delete') {
          deletes.push(lines[i])
          i++
        }
        groups.push({ type: 'deletes', lines: deletes })
      } else {
        i++
      }
    }

    return groups
  }
</script>

<div class="diff-viewer">
  {#if $diffLoading}
    <div class="loading">Loading...</div>
  {:else if !$currentDiff}
    <div class="empty">Select a file</div>
  {:else}
    <div class="diff-content">
      <div class="diff-header">
        <span class="path">{$currentDiff.path}</span>
        <span class="stats">
          {#if $currentDiff.additions > 0}<span class="add">+{$currentDiff.additions}</span>{/if}
          {#if $currentDiff.deletions > 0}<span class="del">-{$currentDiff.deletions}</span>{/if}
        </span>
      </div>

      {#if $currentDiff.isBinary}
        <div class="binary">Binary file</div>
      {:else}
        <div class="diff-body">
          {#each $currentDiff.hunks as hunk}
            <div class="hunk">
              <div class="hunk-header">{hunk.header}</div>
              {#each groupLines(hunk.lines) as group}
                {#if group.type === 'context'}
                  {#each group.lines as line}
                    <div class="line">
                      <span class="ln old">{formatLineNumber(line.oldLine)}</span>
                      <span class="ln new">{formatLineNumber(line.newLine)}</span>
                      <span class="prefix"> </span>
                      <span class="code">{line.content}</span>
                    </div>
                  {/each}
                {:else if group.type === 'change'}
                  {#each group.adds as line}
                    <div class="line add">
                      <span class="ln old"></span>
                      <span class="ln new">{formatLineNumber(line.newLine)}</span>
                      <span class="prefix">+</span>
                      <span class="code">{line.content}</span>
                    </div>
                  {/each}
                  {#each group.deletes as line}
                    <div class="line del">
                      <span class="ln old">{formatLineNumber(line.oldLine)}</span>
                      <span class="ln new"></span>
                      <span class="prefix">-</span>
                      <span class="code">{line.content}</span>
                    </div>
                  {/each}
                {:else if group.type === 'deletes'}
                  {#each group.lines as line}
                    <div class="line del">
                      <span class="ln old">{formatLineNumber(line.oldLine)}</span>
                      <span class="ln new"></span>
                      <span class="prefix">-</span>
                      <span class="code">{line.content}</span>
                    </div>
                  {/each}
                {:else if group.type === 'adds'}
                  {#each group.lines as line}
                    <div class="line add">
                      <span class="ln old"></span>
                      <span class="ln new">{formatLineNumber(line.newLine)}</span>
                      <span class="prefix">+</span>
                      <span class="code">{line.content}</span>
                    </div>
                  {/each}
                {/if}
              {/each}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .diff-viewer {
    flex: 1;
    overflow-y: auto;
    background: var(--bg-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
  }

  .loading,
  .empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
  }

  .diff-content {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .diff-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
  }

  .path {
    font-weight: 600;
    color: var(--text-primary);
  }

  .stats {
    display: flex;
    gap: 8px;
  }

  .stats .add {
    color: var(--success);
  }

  .stats .del {
    color: var(--danger);
  }

  .binary {
    padding: 24px;
    text-align: center;
    color: var(--text-muted);
  }

  .diff-body {
    flex: 1;
    overflow-y: auto;
  }

  .hunk {
    margin-bottom: 8px;
  }

  .hunk-header {
    padding: 4px 12px;
    background: rgba(88, 166, 255, 0.1);
    color: var(--accent);
    font-size: 11px;
  }

  .line {
    display: flex;
    line-height: 20px;
  }

  .ln {
    width: 40px;
    padding: 0 8px;
    text-align: right;
    color: var(--text-muted);
    background: var(--bg-secondary);
    user-select: none;
    flex-shrink: 0;
  }

  .prefix {
    width: 20px;
    text-align: center;
    user-select: none;
    flex-shrink: 0;
  }

  .code {
    flex: 1;
    padding-left: 8px;
    white-space: pre;
    overflow-x: auto;
    color: var(--text-primary);
  }

  .line.add {
    background: rgba(63, 185, 80, 0.15);
  }

  .line.add .prefix {
    color: var(--success);
  }

  .line.del {
    background: rgba(248, 81, 73, 0.15);
  }

  .line.del .prefix {
    color: var(--danger);
  }
</style>

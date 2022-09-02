<script>

  let paste = ''
  let language = ''
  let author = ''
  let burn = true
  let created = false

  export async function post(paste, language, author, burn) {
    const resp = await fetch('/api/v1/paste', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        paste,
        language,
        author,
        burn,
      }),
    }).then((data) => {
      return data.json()
    })

    if (resp.success) {
      created = true

      window.location.replace('/paste/' + resp.data.paste_id)
    }
  }
</script>

<div class="Header">
  <div class="Header-item">
    <a href="/" class="Header-link f4 d-flex flex-items-center">
      <!-- <%= octicon "mark-github", class: "mr-2", height: 32 %> -->
      <svg
        height="32"
        class="octicon octicon-mark-github mr-2"
        viewBox="0 0 16 16"
        version="1.1"
        width="32"
        aria-hidden="true"
        ><path
          fill-rule="evenodd"
          d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"
        /></svg
      >
      <span>Wastebin</span>
    </a>
  </div>
</div>
<div data-color-mode="auto" data-light-theme="light" data-dark-theme="dark_dimmed" class="p-3">
  <form>
    <div class="form-group">
      <div class="form-group-header">
        <label for="paste-textarea">Paste</label>
      </div>
      <div class="form-group-body">
        <textarea
          class="form-control"
          id="Start typing…"
          style="font-family:monospace"
          bind:value={paste}
        />
      </div>
    </div>
    <div class="form-checkbox">
      <label>
        <input
          type="checkbox"
          checked="checked"
          aria-describedby="help-text-for-checkbox"
          bind:value={burn}
        />
        Burn?
      </label>
      <p class="note" id="help-text-for-checkbox">
        This will delete <strong>awesome</strong> and <strong>amazing</strong> things!
      </p>
    </div>
  </form>
  <button
    class="btn-mktg arrow-target-mktg btn-signup-mktg"
    type="button"
    on:click={() => {
      post(paste, language, author, burn)
    }}
    >Create Paste
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="octicon arrow-symbol-mktg"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
    >
      <path
        fill="currentColor"
        d="M7.28033 3.21967C6.98744 2.92678 6.51256 2.92678 6.21967 3.21967C5.92678 3.51256 5.92678 3.98744 6.21967 4.28033L7.28033 3.21967ZM11 8L11.5303 8.53033C11.8232 8.23744 11.8232 7.76256 11.5303 7.46967L11 8ZM6.21967 11.7197C5.92678 12.0126 5.92678 12.4874 6.21967 12.7803C6.51256 13.0732 6.98744 13.0732 7.28033 12.7803L6.21967 11.7197ZM6.21967 4.28033L10.4697 8.53033L11.5303 7.46967L7.28033 3.21967L6.21967 4.28033ZM10.4697 7.46967L6.21967 11.7197L7.28033 12.7803L11.5303 8.53033L10.4697 7.46967Z"
      />
      <path stroke="currentColor" d="M1.75 8H11" stroke-width="1.5" stroke-linecap="round" />
    </svg>
  </button>
</div>

{#if created}
  <div class="p-1">
    <div class="Toast Toast--success">
      <span class="Toast-icon">
        <!-- <%= octicon "check" %> -->
        <svg
          width="12"
          height="16"
          viewBox="0 0 12 16"
          class="octicon octicon-check"
          aria-hidden="true"
        >
          <path fill-rule="evenodd" d="M12 5l-8 8-4-4 1.5-1.5L4 10l6.5-6.5L12 5z" />
        </svg>
      </span>
      <span class="Toast-content">Paste created, redirecting</span>
    </div>
  </div>
{/if}

<style>
  .form-control {
    font-family: monospace;
    font-size: medium;
  }
</style>

<script>
  import { onMount } from 'svelte'
  import { HighlightAuto } from 'svelte-highlight'
  import githubDarkDimmed from 'svelte-highlight/styles/github-dark-dimmed'

  let paste = {}
  let errorMessage = ''

  onMount(async function () {
    // Get paste ID from URL
    const id = window.location.pathname.split('/')[2]
    // Send GET request to retrieve paste
    try {
      const response = await fetch(`/api/v1/paste/${id}`, {
        method: 'GET',
        'Content-Type': 'application/json',
      })
      if (!response.ok) {
        throw new Error(response.statusText)
      }
      paste = await response.json()
    } catch (error) {
      errorMessage = error.message
    }
  })


  function downloadPaste() {
    // Create a link to the paste and simulate a click on it to download the paste
    const link = document.createElement('a')
    link.href = `data:text/plain;charset=utf-8,${encodeURIComponent(paste.content)}`
    link.download = 'paste.txt'
    link.click()
  }
</script>

<svelte:head>
  {@html githubDarkDimmed}
</svelte:head>
<div class="buttons">
  <button id="download-button" class="button" on:click={downloadPaste}>Download</button>
  <button id="copy-button" data-clipboard-text={paste.content} disabled>Copy to clipboard</button>
</div>
{#if errorMessage}
  <div class="error-message">
    <pre>{errorMessage}</pre>
  </div>
{:else}
  <HighlightAuto code={paste.content} />
{/if}

<style>
  .buttons {
    display: flex;
    justify-content: center;
    margin: 1rem 0;
  }
  .error-message {
    color: red;
  }
  .button {
    margin: 0 0.5rem;
  }

  #download-button {
    background-color: #4caf50;
    color: white;
  }
  #copy-button {
    background-color: #2196f3;
    color: white;
  }
</style>

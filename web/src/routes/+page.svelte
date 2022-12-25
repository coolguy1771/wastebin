<script>
  import { onMount } from 'svelte'

  let created = false
  let burn = false
  let syntaxes = ['js,ts,jsx,tsx']

  function onDrop(event) {
    event.preventDefault()
    const file = event.dataTransfer.files[0]
    const reader = new FileReader()
    reader.onload = (e) => {
      paste = e.target.result
    }
    reader.readAsText(file)
  }
  function onDragOver(event) {
    event.preventDefault()
  }
  //onMount(async () => {
  //  const resp = await fetch('/api/v1/syntaxes').then((data) => {
  //    return data.json()
  //  })
  //  syntaxes = resp.data
  //})
</script>

<svelte:head>
  <title>Wastebin</title>
</svelte:head>

<form action="/api/v1/paste" method="post">
  <div class="container">
    <div class="content">
      <textarea
        id="text"
        name="text"
        autocorrect="off"
        autocomplete="off"
        spellcheck="false"
        placeholder="<paste text or drop file here>"
        ondrop="onDrop"
        ondragover="onDragOver"
      />
    </div>
    <div class="expiration-list">
      <select name="expires" size="5">
        <option selected="" value="">never</option>
        <option value="600">10 minutes</option>
        <option value="3600">1 hour</option>
        <option value="86400">1 day</option>
        <option value="604800">1 week</option>
        <option value="215308800">1 year</option>
        <option value="burn">🔥 after reading</option>
      </select>
    </div>
    <div class="extensions-list">
      <select name="extension" size="21">
        {#each syntaxes as syntax}
          {#if syntax.file_extensions.len() > 0}
            <option value="${syntax.file_extensions.first().unwrap()}">{syntax.name}</option>
          {/if}
        {/each}
      </select>
    </div>
    <div class="paste-button">
      <button type="submit" title="Paste">Paste</button>
    </div>
  </div>
</form>

<!--{#if burn}
  <div class="center">
    Copy <a href=`/paste/${id}`>this link</a>. It will be deleted after reading.
  </div>
{/if} -->
<style>
  /* use a css grid layout */

  textarea {
    width: 100%;
    height: 100%;
    resize: none;
    border: none;
    outline: none;
    background: #1e1e1e;
    color: #d4d4d4;
    font-family: 'Fira Code', monospace;
    font-size: 1.2rem;
    padding: 1rem;
    box-sizing: border-box;
  }
  textarea::placeholder {
    color: #6a9955;
  }
  .container {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    grid-template-rows: 1fr 1fr;
    grid-template-areas:
      'content content content'
      'expiration extensions paste';
    height: 100vh;
  }
  .expiration-list {
    grid-area: expiration;
  }
  .expiration-list select {
    width: 100%;
    height: 100%;
    border: none;
    outline: none;
    background: #1e1e1e;
    color: #d4d4d4;
    font-family: 'Fira Code', monospace;
    font-size: 1.2rem;
    padding: 1rem;
    box-sizing: border-box;
  }
</style>

<script>
  import { onMount } from 'svelte'

  let created = false
  let burn = false

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

  // Handle form submission
  async function onSubmit(event) {
    event.preventDefault()

    // Gather form data
    const formData = new FormData(event.target)
    const content = formData.get('text')
    const expires = formData.get('expires')
    const extension = formData.get('extension')

    // Send the POST request to create the paste
    const response = await fetch('/api/v1/paste', {
      method: 'POST',
      body: formData,
    })

    // If the request was successful, redirect the user to the URL of the created paste
    if (response.ok) {
      const pasteData = await response.json()
      const pasteURL = `/paste/${pasteData.uuid}`
      window.location = pasteURL
    }
  }
</script>

<div>
  <form on:submit={onSubmit}>
    <div class="container">
      <div class="content">
        <textarea
          id="text"
          name="text"
          autocorrect="off"
          autocomplete="off"
          spellcheck="false"
          placeholder="<paste text or drop file here>"
          on:drop={onDrop}
          on:dragover={onDragOver}
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
        </select>
        <!-- checbox for burning-->
        <input type="checkbox" name="burn" label="Burn?" bind:value={burn} />
      </div>
      <div class="paste-button">
        <button type="submit" title="Paste">Paste</button>
      </div>
    </div>
  </form>
</div>

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
    grid-template-columns: 80% 20%;
    grid-template-rows: 50% 50%;
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

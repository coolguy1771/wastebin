<script>
  let paste = {
    content: '',
    expires: '',
    extension: '',
    burn: false,
  }

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
    reader.readAsText(file)
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
          bind:value={paste.content}
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
      </div>
      <div class="burn-checkbox">
        <input type="checkbox" name="Burn" bind:value={paste.burn} />
        <label for="Burn">Burn after reading</label>
      </div>
      <div class="paste-button">
        <button type="submit" title="Paste">Paste</button>
      </div>
    </div>
  </form>
</div>

<style>
  .container {
    display: grid; /* use a grid layout */
    grid-template-columns: 2fr 1fr; /* divide the container into two columns, with the textarea taking up 2/3 of the width */
    grid-template-rows: 1fr 1fr 1fr 1fr; /* four rows, one for each element */
    height: 100vh; /* make the container fill the entire viewport */
    gap: 1rem; /* add a gap between the columns */
  }
  .content {
    grid-column: 1 / 2; /* make the content field span across the first column */
    grid-row: 1 / 5; /* make the content field span across all four rows */
  }
  .expiration-list {
    grid-column: 2;
    grid-row: 2; /* second row */
  }
  .burn-checkbox {
    grid-column: 2;
    grid-row: 3; /* third row */
  }
  .paste-button {
    grid-column: 2;
    grid-row: 4; /* fourth row */
  }
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

  .burn-checkbox {
    display: flex;
    align-items: left;
    justify-content: left;
    background-color: #1e1e1e;
    height: fit-content;
  }

  .burn-checkbox input {
    margin: 0.5em;
    background-color: #1e1e1e;
  }

  .burn-checkbox label {
    color: #d4d4d4;
    font-family: 'Fira Code', monospace;
    font-size: 1.2rem;
  }

  .expiration-list select,
  .paste-button button {
    width: 100%;
    height: 100%;
    border: none;
    outline: none;
    color: #d4d4d4;
    font-family: 'Fira Code', monospace;
    font-size: 1.2rem;
    padding: 1rem;
  }
  .expiration-list select {
    background: #1e1e1e;
    height: 100%;
  }
  .paste-button button {
    background: #6a9955;
  }
</style>

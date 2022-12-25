<script>
  let paste
  let code
  let language
  let id

  //onMount(async function () {
  //  id = window.location.pathname.split('/')[2]
  //  const response = await fetch(`/api/v1/paste/${id}`, {
  //    method: 'GET',
  //  })
  //  const data = await response.json()
  //  paste = data.data
  //  code = paste.paste
  //  language = paste.language
  //})

  // On event
  function onEvent(event) {
    if (event.key == 'n') {
      window.location.href = '/'
    } else if (event.key == 'r') {
      window.location.href = `/paste/${id}?raw=true`
    } else if (event.key == 'y') {
      navigator.clipboard.writeText(window.location.href)
    } else if (event.key == 'd') {
      window.location.href = `/download/${id}.${extension}`
    }
  }
</script>

<div>
  <div>
    <pre class="code">${paste.code}</pre>
  </div>
  <div class="paste-box">
    <!-- if deleteion is possible -->
    {#if paste.burn}
      <a class="punctuation definition tag" href="/api/v1/{id}">del</a> •
    {/if}

    <a class="punctuation definition tag" href="/paste/{id}?dl={paste.extension}">↓</a>
    <a class="punctuation definition tag" href="/paste/{id}?fmt=raw">raw</a>
    <a class="punctuation definition tag" href="/">new</a>
  </div>
</div>

<style>
</style>

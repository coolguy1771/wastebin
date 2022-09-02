<script>
  import { onMount } from 'svelte'

  let paste
  let code
  let language
  let id

  onMount(async function () {
    id = window.location.pathname.split('/')[2]
    const response = await fetch(`/api/v1/paste/${id}`, {
      method: 'GET',
    })
    const data = await response.json()
    paste = data.data
    code = paste.paste
    language = paste.language
  })
</script>

<pre>{code}</pre>

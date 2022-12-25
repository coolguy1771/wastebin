<script>
  import { onMount } from 'svelte'
  import { HighlightAuto } from 'svelte-highlight'
  import githubDarkDimmed from 'svelte-highlight/styles/github-dark-dimmed'

  let paste

  onMount(async function () {
    // Get paste ID from URL
    const id = window.location.pathname.split('/')[2]

    paste = await response.json()
  })
</script>

<svelte:head>
  {@html githubDarkDimmed}
</svelte:head>

<HighlightAuto code={paste.content} />

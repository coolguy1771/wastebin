<script>
  import '../app.css'
  import { fade } from 'svelte/transition'
  import { onMount } from 'svelte'

  let visible = false

  onMount(() => {
    visible = true
  })
</script>

{#if visible}
  <main id="container" transition:fade>
    <slot />
  </main>
{/if}

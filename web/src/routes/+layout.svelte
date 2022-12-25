<script>
  import '../app.css'
  import { fade } from 'svelte/transition'
  import { onMount } from 'svelte'

  let visible = false

  onMount(() => {
    visible = true
  })
</script>

<svelte:head>
  <title>Wastebin</title>
</svelte:head>

{#if visible}
  <main id="container" transition:fade>
    <slot />
  </main>
{/if}

<style>
</style>

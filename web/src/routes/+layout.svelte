<script>
  import '../app.css'
  import { fade } from 'svelte/transition'
</script>

<svelte:head>
  <title>Wastebin</title>
</svelte:head>

<main class="container" transition:fade duration={300} easing="ease-in-out">
  <slot />
</main>

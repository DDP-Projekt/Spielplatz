<script lang="ts">
	import favicon_ico from '$lib/assets/favicon/favicon.ico';
	import favicon_small from '$lib/assets/favicon/favicon-16x16.png';
	import favicon_medium from '$lib/assets/favicon/favicon-32x32.png';
	import apple from '$lib/assets/favicon/apple-touch-icon.png';
	import safari from '$lib/assets/favicon/safari-pinned-tab.svg';
	const themeColor = "#14437f"

	let { children } = $props();
</script>

<svelte:head>
	<meta name="description" content="Probiere DDP direkt in deinem Browser aus!">

	<!--favicon-->
	<link rel="shortcut icon" href={favicon_ico}>
	<link rel="icon" type="image/png" sizes="32x32" href={favicon_medium}>
	<link rel="icon" type="image/png" sizes="16x16" href={favicon_small}>
	<link rel="apple-touch-icon" sizes="180x180" href={apple}>
	<link rel="mask-icon" href={safari} color={themeColor}>
	<link rel="manifest" href="/static/site.webmanifest">
	<meta name="msapplication-TileColor" content={themeColor}>
	<meta name="theme-color" content={themeColor}>
</svelte:head>

{@render children()}

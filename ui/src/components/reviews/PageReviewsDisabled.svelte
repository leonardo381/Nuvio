<script>
    import { onMount } from "svelte";
    import { replace } from "svelte-spa-router";

    // NUVIO CUSTOM START: Reviews route is temporarily disabled in product UI.
    onMount(() => {
        replace("/cms");
    });
    // NUVIO CUSTOM END: Reviews route is temporarily disabled in product UI.
</script>


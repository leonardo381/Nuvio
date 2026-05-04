<script>
    import { replace } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";

    handler();

    function handler() {
        if (ApiClient.isAdminSuperuser()) {
            replace("/collections");
        } else if (ApiClient.isClientSuperuser()) {
            replace("/cms");
        } else {
            ApiClient.logout();
        }
    }
</script>

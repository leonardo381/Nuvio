<script>
    import { replace } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";

    handler();

    function handler() {
        if (ApiClient.isSuperuserAuth()) {
            replace("/collections");
        } else {
            ApiClient.logout();
        }
    }
</script>

<script setup lang="ts">
import { inject, computed, ref, onMounted, reactive } from "vue";
import { AuthKey } from "@/keys";
import { useRoute } from "vue-router";
import { useUser } from "@/services/auth";
import BreakglassSessionService from "@/services/breakglassSession";
import type { BreakglassSessionRequest } from "@/model/breakglassSession";

const route = useRoute()
const user = useUser();
const auth = inject(AuthKey);
const authenticated = computed(() => user.value && !user.value?.expired);
const service = new BreakglassSessionService(auth!);

const resourceName = ref(route.query.name?.toString() || "");

const state = reactive({
  breakglasses: new Array<any>(),
  loading: true,
  refreshing: false,
  search: "",
});

onMounted(async () => {
  const params: BreakglassSessionRequest = {uname: resourceName.value}
  console.log(params)
  state.breakglasses = await service.getSessionStatus(params);
  console.log(state.breakglasses);
  state.loading = false;
});

</script>


<template>
  <main>
    <div v-if="authenticated" class="center">
      Displaying resource {{ resourceName }}
    </div>
  </main>
</template>

<style scoped>
.center {
  text-align: center;
}

scale-data-grid {
  display: block;
  margin: 0 auto;
  max-width: 600px;
}

scale-card {
  display: block;
  margin: 0 auto;
  max-width: 500px;
}
</style>

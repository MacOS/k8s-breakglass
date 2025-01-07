<script setup lang="ts">
import ClusterAccessCard from "@/components/ClusterAccessCard.vue";
import { inject, onMounted, reactive } from "vue";
import { AuthKey } from "@/keys";
import BreakglassSessionService from "@/services/breakglassSession";
import type { ClusterAccessReview } from "@/model/cluster_access";
import useCurrentTime from "@/util/currentTime";
import { computed } from "@vue/reactivity";

const auth = inject(AuthKey);
const clusterAccessService = new BreakglassSessionService(auth!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
const time = useCurrentTime();

const state = reactive({
  reviews: new Array<ClusterAccessReview>(),
  loading: true,
  refreshing: false,
  search: "",
});

onMounted(async () => {
  state.reviews = await clusterAccessService.getClusterAccessReviews();
  console.log(state.reviews);
  console.log(state.reviews.length);
  state.loading = false;
});

async function refresh() {
  state.refreshing = true;
  state.reviews = await clusterAccessService.getClusterAccessReviews();
  state.refreshing = false;
}


async function onAccept(car: ClusterAccessReview) {
  await clusterAccessService.approveReview(car);
  state.reviews = await clusterAccessService.getClusterAccessReviews();
}

async function onReject(car: ClusterAccessReview) {
  await clusterAccessService.rejectReview(car);
  state.reviews = await clusterAccessService.getClusterAccessReviews();
}

</script>

<template>
  <main>
    <div v-if="state.loading" class="loading">
      <scale-loading-spinner size="large" />
    </div>
    <div v-else-if="state.reviews.length > 0">
      <div class="search">
        <scale-text-field
          label="Search"
          class="search-field"
          :value="state.search"
          @scaleChange="(ev: any) => state.search = ev.target.value"
        ></scale-text-field>
        <div class="refresh">
          <scale-loading-spinner v-if="state.refreshing"></scale-loading-spinner>
          <scale-button v-else icon-only="true" icon-position="before" variant="secondary" @click="refresh()">
            <scale-icon-action-refresh></scale-icon-action-refresh>
          </scale-button>
        </div>
      </div>
      <div class="cluster-access-list">
        <ClusterAccessCard
          v-for="rev in state.reviews"
          :key="rev.uid"
          class="card"
          :review="rev"
          :time="time"
          @accept="
            () => {
              onAccept(rev);
            }
          "
          @reject="
            () => {
              onReject(rev);
            }
          "
        >
        </ClusterAccessCard>
      </div>
    </div>
    <div v-else class="not-found">No cluster requests found.</div>
  </main>
</template>

<style scoped>
main {
  margin: 3rem auto;
  max-width: 1200px;
}

.loading {
  margin: 2rem auto;
  text-align: center;
}

.search {
  max-width: 400px;
  margin: 1rem auto;
  display: flex;
  align-items: center;
}

.search-field {
  flex-grow: 1;
  margin-right: 1rem;
}

.refresh {
  width: 48px;
}

.cluster-access-list {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  justify-content: center;
}

.not-found {
  text-align: center;
}

.card {
  flex-grow: 1;
  flex-shrink: 0;
}
</style>

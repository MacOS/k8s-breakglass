<script setup lang="ts">
import humanizeDuration from "humanize-duration";
import { computed, ref } from "vue";

import type { Breakglass } from "@/model/breakglass";

const humanizeConfig: humanizeDuration.Options = {
  round: true,
  largest: 2,
};

const props = defineProps<{
  breakglass: Breakglass;
  time: number;
}>();

const emit = defineEmits(["request", "drop"]);

const active = computed(() => props.breakglass.expiry > 0);
const lastRequested = ref(0);
const recentlyRequested = () => lastRequested.value + 600_000 > Date.now();

const expiryHumanized = computed(() => {
  if (!active.value) {
    return "";
  }
  const duration: number = props.time - props.breakglass.expiry * 1000;
  return humanizeDuration(duration, humanizeConfig);
});

const durationHumanized = computed(() => {
  return humanizeDuration(props.breakglass.duration * 1000, humanizeConfig);
});

const buttonText = computed(() => {
  if (active.value) {
    return "Already Active";
  } else {
    if (recentlyRequested()) {
      return "Already requested";
    } else {
      return "Request";
    }
  }
});

function request() {
  emit("request");
  lastRequested.value = Date.now();
}
</script>

<template>
  <scale-card>
    <h2 class="to">
      {{ breakglass.to }}
    </h2>
    <span v-if="!active">
      <p>
        From <b>{{ breakglass.from }}</b>
      </p>
      <p>
        For <b>{{ durationHumanized }}</b>
      </p>

      <p v-if="breakglass.approvalGroups && breakglass.approvalGroups.length > 0">
        Requires approval from {{ breakglass.approvalGroups.join(", ") }}
      </p>
      <p v-else>No approvers defined.</p>
    </span>
    <span v-else>
      <p class="expiry">
        Expires in<br />
        <b>{{ expiryHumanized }}</b>
      </p>
    </span>

    <p class="actions">
      <scale-button :disabled="active || recentlyRequested()" @click="request">{{ buttonText }} </scale-button>
      <scale-button v-if="active" variant="secondary" @click="emit('drop')">Drop</scale-button>
    </p>
  </scale-card>
</template>

<style scoped>
scale-card {
  display: inline-block;
  max-width: 300px;
}

scale-button {
  margin: 0 0.4rem;
}

.actions {
  margin-top: 1rem;
  text-align: center;
}

.to,
.expiry {
  text-align: center;
}
</style>

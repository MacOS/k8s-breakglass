<script setup lang="ts">
import { inject, computed, ref, onMounted } from "vue";

import { useRoute } from "vue-router";
import { AuthKey } from "@/keys";
import { useUser } from "@/services/auth";
import BreakglassSessionService from "@/services/breakglassSession";
import type { BreakglassSessionRequest } from "@/model/breakglassSession";

const auth = inject(AuthKey);
const breakglassSession = new BreakglassSessionService(auth!);
const route = useRoute()
const user = useUser();
const authenticated = computed(() => user.value && !user.value?.expired);

const userName = ref(route.query.username || "");
const clusterName = ref(route.query.cluster || "");
const clusterGroup = ref(route.query.group || "breakglass-create-all");
const alreadyRequested = ref(false);
const requestStatusMessage = ref("");

// TODO probably move to before create
const hasUsername = route.query.username ? true : false
const hasCluster = route.query.cluster ? true : false

// Function to handle the form submission
const handleSendButtonClick = async () => {
  const sessionRequest = {
    clustername: clusterName.value,
    username: userName.value,
    clustergroup: clusterGroup.value
  } as BreakglassSessionRequest

  await breakglassSession.requestSession(sessionRequest).then(response => {
    switch (response.status) {
      case 200:
        alreadyRequested.value = true
        requestStatusMessage.value = "Request already created"
        break
      case 201:
        alreadyRequested.value = true
        requestStatusMessage.value = "Successfully created request"
        break
      default:
        requestStatusMessage.value = "Failed to create breakglass session, please try again later"
    }
  }).catch(
    err => {
      switch (err.status) {
        case 401:
          requestStatusMessage.value = "No transition defined for requested group."
          break
      default:
        requestStatusMessage.value = "Failed to create breakglass session, please try again later"
        console.log(err)
      }
    }
  )
};

const onInput = () => {
  alreadyRequested.value = false
  requestStatusMessage.value = ""
}

onMounted(() => {
  console.log(`the component is now mounted.`)
})

</script>

<template>
  <main>
    <scale-card class="centered">
      <div v-if="authenticated" class="center">
        <p>Request for group assignment</p>
        <form @submit.prevent="handleSendButtonClick">
          <div>
            <label for="user_name">Username:</label>
            <input type="text" id="user_name" v-model="userName" :disabled="hasUsername" placeholder="Enter user name"
              required />
          </div>
          <div>
            <label for="cluster_name">Cluster name:</label>
            <input type="text" id="cluster_name" v-model="clusterName" :disabled="hasCluster"
              placeholder="Enter cluster name" required />
          </div>
          <div style="margin-bottom: 5px;">
            <label for="cluser_group">Cluster group:</label>
            <input type="text" id="" v-model="clusterGroup" placeholder="Enter cluster group" v-on:input="onInput"
              required />
          </div>

          <div>
            <scale-button type="submit" :disabled="alreadyRequested" size="small">Send</scale-button>
          </div>

          <p v-if="requestStatusMessage !== ''">{{ requestStatusMessage }}</p>

        </form>

      </div>
    </scale-card>
  </main>
</template>

<style scoped>
.center {
  text-align: center;
}

input {
  margin-left: 5px;
}

label {
  display: inline-block;
  width: 110px;
  text-align: right;
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

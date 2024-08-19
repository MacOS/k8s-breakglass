import ApproveView from "@/views/ApproveView.vue";
import { createRouter, createWebHistory } from "vue-router";
import BreakglassView from "../views/BreakglassView.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "home",
      component: BreakglassView,
    },
    {
      path: "/approve",
      name: "approve",
      component: ApproveView,
    },
  ],
});

export default router;

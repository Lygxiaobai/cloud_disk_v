import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import { pinia } from "@/stores";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/disk",
  },
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/auth/LoginView.vue"),
    meta: { guestOnly: true },
  },
  {
    path: "/register",
    name: "register",
    component: () => import("@/views/auth/RegisterView.vue"),
    meta: { guestOnly: true },
  },
  {
    path: "/disk",
    name: "disk",
    component: () => import("@/views/disk/DiskView.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/share/:identity",
    name: "share",
    component: () => import("@/views/share/ShareView.vue"),
    props: true,
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  const authStore = useAuthStore(pinia);

  if (authStore.token && !authStore.bootstrapFinished && !authStore.isBootstrapping) {
    await authStore.bootstrap();
  }

  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return {
      name: "login",
      query: {
        redirect: to.fullPath,
      },
    };
  }

  if (to.meta.guestOnly && authStore.isLoggedIn) {
    return { name: "disk" };
  }

  return true;
});
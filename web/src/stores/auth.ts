import { computed, ref } from "vue";
import { defineStore } from "pinia";

import * as authApi from "@/api/auth";
import type { LoginRequest, RegisterRequest, UserDetailResponse } from "@/types/api";
import { clearStoredAuth, readStoredAuth, writeStoredAuth } from "@/utils/auth-storage";
import { getDisplayNameFromToken, getIdentityFromToken } from "@/utils/jwt";

const storedAuth = readStoredAuth();

export const useAuthStore = defineStore("auth", () => {
  const token = ref(storedAuth.token);
  const refreshToken = ref(storedAuth.refreshToken);
  const profile = ref<UserDetailResponse | null>(null);
  const bootstrapFinished = ref(false);
  const isBootstrapping = ref(false);

  const isLoggedIn = computed(() => Boolean(token.value));
  const displayName = computed(() => profile.value?.name || getDisplayNameFromToken(token.value) || "Guest");
  const identity = computed(() => getIdentityFromToken(token.value));

  function persist(): void {
    writeStoredAuth({
      refreshToken: refreshToken.value,
      token: token.value,
    });
  }

  function setTokens(nextToken: string, nextRefreshToken: string): void {
    token.value = nextToken;
    refreshToken.value = nextRefreshToken;
    persist();
  }

  function setProfile(nextProfile: UserDetailResponse | null): void {
    profile.value = nextProfile;
  }

  function clearAuth(): void {
    token.value = "";
    refreshToken.value = "";
    profile.value = null;
    bootstrapFinished.value = false;
    clearStoredAuth();
  }

  async function hydrateProfile(): Promise<UserDetailResponse | null> {
    if (!token.value) {
      return null;
    }

    const data = await authApi.fetchCurrentUser();
    profile.value = data;
    return data;
  }

  async function bootstrap(): Promise<void> {
    if (!token.value || bootstrapFinished.value || isBootstrapping.value) {
      return;
    }

    isBootstrapping.value = true;
    try {
      await hydrateProfile();
    } catch {
      profile.value = profile.value ?? null;
    } finally {
      bootstrapFinished.value = true;
      isBootstrapping.value = false;
    }
  }

  async function login(payload: LoginRequest): Promise<void> {
    const data = await authApi.login(payload);
    setTokens(data.token, data.refresh_token);
    bootstrapFinished.value = false;
    await bootstrap();
  }

  async function sendRegisterCode(email: string): Promise<void> {
    await authApi.sendRegisterCode({ email });
  }

  async function registerAccount(payload: RegisterRequest): Promise<void> {
    await authApi.register(payload);
  }

  function logout(): void {
    clearAuth();
  }

  return {
    bootstrap,
    bootstrapFinished,
    clearAuth,
    displayName,
    hydrateProfile,
    identity,
    isBootstrapping,
    isLoggedIn,
    login,
    logout,
    profile,
    refreshToken,
    registerAccount,
    sendRegisterCode,
    setProfile,
    setTokens,
    token,
  };
});

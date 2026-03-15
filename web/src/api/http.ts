import axios, { type AxiosError, type InternalAxiosRequestConfig } from "axios";

import type { RefreshTokenResponse } from "@/types/api";
import { pinia } from "@/stores";
import { useAuthStore } from "@/stores/auth";

interface RequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
  skipAuth?: boolean;
  skipRefresh?: boolean;
}

const baseURL = "/api";

export const rawClient = axios.create({
  baseURL,
  timeout: 20000,
});

const httpClient = axios.create({
  baseURL,
  timeout: 20000,
});

let isRefreshing = false;
let waitingQueue: Array<(token: string | null) => void> = [];

function flushQueue(token: string | null): void {
  waitingQueue.forEach((callback) => callback(token));
  waitingQueue = [];
}

function redirectToLogin(): void {
  const redirect = encodeURIComponent(`${window.location.pathname}${window.location.search}`);
  window.location.replace(`/login?redirect=${redirect}`);
}

httpClient.interceptors.request.use((config) => {
  const requestConfig = config as RequestConfig;
  const authStore = useAuthStore(pinia);

  if (!requestConfig.skipAuth && authStore.token) {
    requestConfig.headers = requestConfig.headers ?? {};
    requestConfig.headers.Authorization = authStore.token;
  }

  return requestConfig;
});

httpClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const authStore = useAuthStore(pinia);
    const requestConfig = error.config as RequestConfig | undefined;
    const status = error.response?.status;

    if (!requestConfig || requestConfig.skipRefresh || requestConfig._retry) {
      return Promise.reject(error);
    }

    if (status !== 401 || !authStore.refreshToken) {
      return Promise.reject(error);
    }

    requestConfig._retry = true;

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        waitingQueue.push((nextToken) => {
          if (!nextToken) {
            reject(error);
            return;
          }

          requestConfig.headers = requestConfig.headers ?? {};
          requestConfig.headers.Authorization = nextToken;
          resolve(httpClient(requestConfig));
        });
      });
    }

    isRefreshing = true;

    try {
      const { data } = await rawClient.put<RefreshTokenResponse>("/refresh/token", {
        refresh_token: authStore.refreshToken,
      });

      authStore.setTokens(data.token, data.refresh_token);
      flushQueue(data.token);

      requestConfig.headers = requestConfig.headers ?? {};
      requestConfig.headers.Authorization = data.token;
      return httpClient(requestConfig);
    } catch (refreshError) {
      const refreshStatus = axios.isAxiosError(refreshError) ? refreshError.response?.status : undefined;

      flushQueue(null);

      if (refreshStatus === 400 || refreshStatus === 401 || refreshStatus === 403) {
        authStore.clearAuth();
        redirectToLogin();
      }

      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);

export default httpClient;

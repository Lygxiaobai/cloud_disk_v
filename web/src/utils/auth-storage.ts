const STORAGE_KEY = "cloud-disk-auth";

export interface StoredAuth {
  refreshToken: string;
  token: string;
}

export function readStoredAuth(): StoredAuth {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return { refreshToken: "", token: "" };
    }

    const parsed = JSON.parse(raw) as Partial<StoredAuth>;
    return {
      refreshToken: parsed.refreshToken ?? "",
      token: parsed.token ?? "",
    };
  } catch {
    return { refreshToken: "", token: "" };
  }
}

export function writeStoredAuth(auth: StoredAuth): void {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(auth));
}

export function clearStoredAuth(): void {
  window.localStorage.removeItem(STORAGE_KEY);
}

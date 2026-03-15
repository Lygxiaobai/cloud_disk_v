import type { JwtPayload } from "@/types/api";

function decodeBase64Url(value: string): string {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/").padEnd(Math.ceil(value.length / 4) * 4, "=");
  return atob(padded);
}

export function parseJwt(token: string): JwtPayload | null {
  if (!token) {
    return null;
  }

  try {
    const parts = token.split(".");
    if (parts.length < 2) {
      return null;
    }
    return JSON.parse(decodeBase64Url(parts[1])) as JwtPayload;
  } catch {
    return null;
  }
}

export function getDisplayNameFromToken(token: string): string {
  const payload = parseJwt(token);
  return payload?.Name ?? payload?.name ?? "";
}

export function getIdentityFromToken(token: string): string {
  const payload = parseJwt(token);
  return payload?.Identity ?? payload?.identity ?? "";
}

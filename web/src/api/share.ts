import httpClient, { rawClient } from "@/api/http";
import type { ShareFileDetailResponse, ShareFileSavePayload } from "@/types/api";

export async function fetchShareDetail(identity: string): Promise<ShareFileDetailResponse> {
  const encodedIdentity = encodeURIComponent(identity);
  const { data } = await rawClient.get<ShareFileDetailResponse>(`/share/file/detail/${encodedIdentity}`);
  return data;
}

export async function saveSharedFile(payload: ShareFileSavePayload): Promise<void> {
  await httpClient.post("/share/file/save", payload);
}
import httpClient, { rawClient } from "@/api/http";
import type {
  ShareDeletePayload,
  ShareFileDetailResponse,
  ShareFileSavePayload,
  ShareListParams,
  ShareListResponse,
} from "@/types/api";

export async function fetchShareDetail(identity: string, accessCode?: string): Promise<ShareFileDetailResponse> {
  const encodedIdentity = encodeURIComponent(identity);
  const { data } = await rawClient.get<ShareFileDetailResponse>(`/share/file/detail/${encodedIdentity}`, {
    params: accessCode ? { access_code: accessCode } : undefined,
  });
  return data;
}

export async function saveSharedFile(payload: ShareFileSavePayload): Promise<void> {
  await httpClient.post("/share/file/save", payload);
}

export async function listShares(params: ShareListParams): Promise<ShareListResponse> {
  const { data } = await httpClient.get<ShareListResponse>("/share/basic/list", { params });
  return data;
}

export async function deleteShares(payload: ShareDeletePayload): Promise<void> {
  await httpClient.delete("/share/basic/delete", { data: payload });
}

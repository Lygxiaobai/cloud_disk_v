import httpClient from "@/api/http";
import type {
  BatchRenamePayload,
  BatchRenameResponse,
  BatchDeletePayload,
  BatchFavoritePayload,
  BatchMovePayload,
  CreateFolderPayload,
  DeleteFilePayload,
  FavoritePayload,
  FileVersionListResponse,
  FileVersionRestorePayload,
  FileVersionRestoreResponse,
  FilePreviewResponse,
  FolderChildrenResponse,
  FolderPathResponse,
  MoveFilePayload,
  RecentFilesResponse,
  RecycleDeletePayload,
  RecycleListParams,
  RecycleListResponse,
  RecycleRestorePayload,
  RenameFilePayload,
  RepositorySavePayload,
  ShareCreatePayload,
  ShareCreateResponse,
  UploadCompletePayload,
  UploadCompleteResponse,
  UploadInitPayload,
  UploadInitResponse,
  UploadStsRefreshPayload,
  UploadStsRefreshResponse,
  UserFileListParams,
  UserFileListResponse,
} from "@/types/api";

export async function listFiles(params: UserFileListParams): Promise<UserFileListResponse> {
  const { data } = await httpClient.get<UserFileListResponse>("/user/file/list", { params });
  return data;
}

export async function listFolderChildren(folderId: number): Promise<FolderChildrenResponse> {
  const { data } = await httpClient.get<FolderChildrenResponse>(`/user/folder/children/${folderId}`);
  return data;
}

export async function fetchFolderPath(identity: string): Promise<FolderPathResponse> {
  const encodedIdentity = encodeURIComponent(identity);
  const { data } = await httpClient.get<FolderPathResponse>(`/user/folder/path/${encodedIdentity}`);
  return data;
}

export async function createFolder(payload: CreateFolderPayload): Promise<void> {
  await httpClient.put("/user/folder/create", payload);
}

export async function uploadInit(payload: UploadInitPayload): Promise<UploadInitResponse> {
  const { data } = await httpClient.post<UploadInitResponse>("/file/upload/init", payload);
  return data;
}

export async function uploadComplete(payload: UploadCompletePayload): Promise<UploadCompleteResponse> {
  const { data } = await httpClient.post<UploadCompleteResponse>("/file/upload/complete", payload);
  return data;
}

export async function refreshUploadSTS(payload: UploadStsRefreshPayload): Promise<UploadStsRefreshResponse> {
  const { data } = await httpClient.post<UploadStsRefreshResponse>("/file/upload/sts/refresh", payload);
  return data;
}

export async function saveUploadedRepository(payload: RepositorySavePayload): Promise<void> {
  await httpClient.post("/user/repository/save", payload);
}

export async function renameFile(payload: RenameFilePayload): Promise<void> {
  await httpClient.put("/user/file/name/update", payload);
}

export async function batchRenameFiles(payload: BatchRenamePayload): Promise<BatchRenameResponse> {
  const { data } = await httpClient.put<BatchRenameResponse>("/user/file/batch/rename", payload);
  return data;
}

export async function deleteFile(payload: DeleteFilePayload): Promise<void> {
  await httpClient.delete("/user/file/delete", { data: payload });
}

export async function moveFile(payload: MoveFilePayload): Promise<void> {
  await httpClient.put("/user/file/move", payload);
}

export async function previewFile(identity: string): Promise<FilePreviewResponse> {
  const encodedIdentity = encodeURIComponent(identity);
  const { data } = await httpClient.get<FilePreviewResponse>(`/user/file/preview/${encodedIdentity}`);
  return data;
}

export async function listFileVersions(identity: string): Promise<FileVersionListResponse> {
  const encodedIdentity = encodeURIComponent(identity);
  const { data } = await httpClient.get<FileVersionListResponse>(`/user/file/version/list/${encodedIdentity}`);
  return data;
}

export async function restoreFileVersion(payload: FileVersionRestorePayload): Promise<FileVersionRestoreResponse> {
  const { data } = await httpClient.put<FileVersionRestoreResponse>("/user/file/version/restore", payload);
  return data;
}

export async function listRecentFiles(limit = 10): Promise<RecentFilesResponse> {
  const { data } = await httpClient.get<RecentFilesResponse>("/user/file/recent", { params: { limit } });
  return data;
}

export async function favoriteFile(payload: FavoritePayload): Promise<void> {
  await httpClient.put("/user/file/favorite", payload);
}

export async function batchDeleteFiles(payload: BatchDeletePayload): Promise<void> {
  await httpClient.delete("/user/file/batch/delete", { data: payload });
}

export async function batchMoveFiles(payload: BatchMovePayload): Promise<void> {
  await httpClient.put("/user/file/batch/move", payload);
}

export async function batchFavoriteFiles(payload: BatchFavoritePayload): Promise<void> {
  await httpClient.put("/user/file/batch/favorite", payload);
}

export async function listRecycleFiles(params: RecycleListParams): Promise<RecycleListResponse> {
  const { data } = await httpClient.get<RecycleListResponse>("/user/recycle/list", { params });
  return data;
}

export async function restoreRecycleFiles(payload: RecycleRestorePayload): Promise<void> {
  await httpClient.put("/user/recycle/restore", payload);
}

export async function deleteRecycleFiles(payload: RecycleDeletePayload): Promise<void> {
  await httpClient.delete("/user/recycle/delete", { data: payload });
}

export async function createShare(payload: ShareCreatePayload): Promise<ShareCreateResponse> {
  const { data } = await httpClient.post<ShareCreateResponse>("/share/basic/create", payload);
  return data;
}

import httpClient from "@/api/http";
import type {
  CreateFolderPayload,
  DeleteFilePayload,
  FolderChildrenResponse,
  FolderPathResponse,
  MoveFilePayload,
  RenameFilePayload,
  RepositorySavePayload,
  ShareCreatePayload,
  ShareCreateResponse,
  UploadMultipartResponse,
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

export async function uploadFile(file: File): Promise<UploadMultipartResponse> {
  const formData = new FormData();
  formData.append("file", file);

  const { data } = await httpClient.post<UploadMultipartResponse>("/file/upload/multipart", formData);
  return data;
}

export async function saveUploadedRepository(payload: RepositorySavePayload): Promise<void> {
  await httpClient.post("/user/repository/save", payload);
}

export async function renameFile(payload: RenameFilePayload): Promise<void> {
  await httpClient.put("/user/file/name/update", payload);
}

export async function deleteFile(payload: DeleteFilePayload): Promise<void> {
  await httpClient.delete("/user/file/delete", { data: payload });
}

export async function moveFile(payload: MoveFilePayload): Promise<void> {
  await httpClient.put("/user/file/move", payload);
}

export async function createShare(payload: ShareCreatePayload): Promise<ShareCreateResponse> {
  const { data } = await httpClient.post<ShareCreateResponse>("/share/basic/create", payload);
  return data;
}
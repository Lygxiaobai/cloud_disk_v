export interface LoginRequest {
  name: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
}

export interface RegisterRequest {
  code: string;
  email: string;
  name: string;
  password: string;
}

export interface MailCodeRequest {
  email: string;
}

export interface UserDetailResponse {
  email: string;
  name: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  token: string;
  refresh_token: string;
}

export interface UserFileListParams {
  id?: number;
  identity?: string;
  page?: number;
  size?: number;
}

export interface UserFile {
  ext: string;
  id: number;
  identity: string;
  is_dir: number;
  name: string;
  path: string;
  repository_identity: string;
  size: number;
}

export interface UserFileListResponse {
  count: number;
  list: UserFile[];
}

export interface FolderTreeNode {
  has_children: number;
  id: number;
  identity: string;
  name: string;
  parent_id: number;
}

export interface FolderChildrenResponse {
  list: FolderTreeNode[];
}

export interface FolderPathItem {
  id: number;
  identity: string;
  name: string;
}

export interface FolderPathResponse {
  list: FolderPathItem[];
}

export interface CreateFolderPayload {
  name: string;
  parentId: number;
}

export interface RenameFilePayload {
  identity: string;
  name: string;
}

export interface DeleteFilePayload {
  identity: string;
}

export interface MoveFilePayload {
  identity: string;
  parent_identity: string;
}

export interface UploadMultipartResponse {
  identity: string;
}

export interface RepositorySavePayload {
  ext: string;
  name: string;
  parentId: number;
  repositoryIdentity: string;
}

export interface ShareCreatePayload {
  expired_time: number;
  user_repository_identity: string;
}

export interface ShareCreateResponse {
  identity: string;
}

export interface ShareFileDetailResponse {
  ext: string;
  name: string;
  path: string;
  repository_identity: string;
  size: number;
}

export interface ShareFileSavePayload {
  parent_id: number;
  repository_identity: string;
}

export interface JwtPayload {
  Identity?: string;
  Name?: string;
  exp?: number;
  identity?: string;
  name?: string;
}
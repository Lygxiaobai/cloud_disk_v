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
  query?: string;
  file_type?: string;
  favorite_only?: boolean;
  order_by?: string;
  order_dir?: "asc" | "desc";
  scope?: "folder" | "all";
  view?: "duplicates" | "large";
  min_size_mb?: number;
}

export interface UserFile {
  created_at: string;
  deleted_at?: string;
  duplicate_count?: number;
  duplicate_group_size?: number;
  ext: string;
  id: number;
  identity: string;
  is_dir: number;
  is_favorite: number;
  last_accessed_at?: string;
  name: string;
  path: string;
  repository_identity: string;
  size: number;
  updated_at: string;
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

export interface BatchRenamePayload {
  identities: string[];
  prefix?: string;
  suffix?: string;
  find_text?: string;
  replace_text?: string;
  apply_sequence?: boolean;
  start_index?: number;
  step?: number;
  padding?: number;
  keep_ext?: boolean;
}

export interface BatchRenameItem {
  identity: string;
  old_name: string;
  new_name: string;
}

export interface BatchRenameResponse {
  list: BatchRenameItem[];
}

export interface DeleteFilePayload {
  identity: string;
}

export interface MoveFilePayload {
  identity: string;
  parent_identity: string;
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
  access_code?: string;
  allow_download?: number;
}

export interface ShareCreateResponse {
  identity: string;
  access_code_set: boolean;
}

export interface ShareFileDetailResponse {
  ext: string;
  allow_download: number;
  need_code: boolean;
  name: string;
  path: string;
  repository_identity: string;
  size: number;
}

export interface ShareFileSavePayload {
  share_identity?: string;
  parent_id: number;
  repository_identity?: string;
  access_code?: string;
}

export interface ShareListParams {
  page?: number;
  size?: number;
  query?: string;
}

export interface ShareListItem {
  identity: string;
  user_file_identity: string;
  name: string;
  ext: string;
  size: number;
  click_num: number;
  allow_download: number;
  access_code_set: boolean;
  created_at: string;
  expires_at: string;
  expired: boolean;
}

export interface ShareListResponse {
  count: number;
  list: ShareListItem[];
}

export interface ShareDeletePayload {
  identities: string[];
}

export interface UploadSTS {
  access_key_id: string;
  access_key_secret: string;
  expiration: string;
  security_token: string;
}

export interface UploadInitPayload {
  ext?: string;
  hash: string;
  name: string;
  parent_id: number;
  parent_identity?: string;
  target_file_identity?: string;
  size: number;
}

export interface UploadInitResponse {
  instant_hit: boolean;
  file_identity?: string;
  object_key?: string;
  oss_bucket?: string;
  oss_endpoint?: string;
  oss_region?: string;
  repository_identity?: string;
  session_identity?: string;
  sts?: UploadSTS;
}

export interface UploadCompletePayload {
  session_identity: string;
}

export interface UploadCompleteResponse {
  file_identity: string;
  repository_identity: string;
}

export interface UploadStsRefreshPayload {
  session_identity: string;
}

export interface UploadStsRefreshResponse {
  object_key: string;
  oss_bucket: string;
  oss_endpoint: string;
  oss_region: string;
  session_identity: string;
  sts: UploadSTS;
}

export interface FilePreviewResponse {
  ext: string;
  kind: "image" | "video" | "audio" | "pdf" | "text" | "download";
  name: string;
  size: number;
  text?: string;
  truncated: boolean;
  url?: string;
}

export interface RecentFilesResponse {
  list: UserFile[];
}

export interface FavoritePayload {
  identity: string;
  is_favorite: number;
}

export interface BatchDeletePayload {
  identities: string[];
}

export interface BatchMovePayload {
  identities: string[];
  parent_identity: string;
}

export interface BatchFavoritePayload {
  identities: string[];
  is_favorite: number;
}

export interface FileVersionItem {
  identity: string;
  file_identity: string;
  repository_identity: string;
  name: string;
  ext: string;
  size: number;
  hash: string;
  action: string;
  is_current: number;
  created_at: string;
}

export interface FileVersionListResponse {
  list: FileVersionItem[];
}

export interface FileVersionRestorePayload {
  file_identity: string;
  version_identity: string;
}

export interface FileVersionRestoreResponse {
  file_identity: string;
  repository_identity: string;
}

export interface RecycleListParams {
  order_by?: string;
  order_dir?: "asc" | "desc";
  page?: number;
  query?: string;
  size?: number;
}

export interface RecycleListResponse {
  count: number;
  list: UserFile[];
}

export interface RecycleRestorePayload {
  identities: string[];
}

export interface RecycleDeletePayload {
  identities: string[];
}

export interface JwtPayload {
  Identity?: string;
  Name?: string;
  exp?: number;
  identity?: string;
  name?: string;
}

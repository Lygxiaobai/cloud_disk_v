package types

type FileUploadMultipartRequest struct {
	Hash string `json:"hash,optional"`
	Name string `json:"name,optional"`
	Ext  string `json:"ext,optional"`
	Size int64  `json:"size,optional"`
	Path string `json:"path,optional"`
}

type FileUploadMultipartResponse struct {
	Identity string `json:"identity"`
}

type FileUploadRequest struct {
	Hash string `json:"hash,optional"`
	Name string `json:"name,optional"`
	Ext  string `json:"ext,optional"`
	Size int64  `json:"size,optional"`
	Path string `json:"path,optional"`
}

type FileUploadResponse struct {
	Identity string `json:"identity"`
	Ext      string `json:"ext"`
	Name     string `json:"name"`
}

type UploadSTS struct {
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Expiration      string `json:"expiration"`
}

type UploadInitRequest struct {
	ParentId       int64  `json:"parent_id,optional"`
	ParentIdentity string `json:"parent_identity,optional"`
	Name           string `json:"name"`
	Ext            string `json:"ext,optional"`
	Hash           string `json:"hash"`
	Size           int64  `json:"size"`
}

type UploadInitResponse struct {
	InstantHit         bool       `json:"instant_hit"`
	SessionIdentity    string     `json:"session_identity,optional"`
	FileIdentity       string     `json:"file_identity,optional"`
	RepositoryIdentity string     `json:"repository_identity,optional"`
	ObjectKey          string     `json:"object_key,optional"`
	OssBucket          string     `json:"oss_bucket,optional"`
	OssRegion          string     `json:"oss_region,optional"`
	OssEndpoint        string     `json:"oss_endpoint,optional"`
	Sts                *UploadSTS `json:"sts,optional"`
}

type UploadCompleteRequest struct {
	SessionIdentity string `json:"session_identity"`
}

type UploadCompleteResponse struct {
	FileIdentity       string `json:"file_identity"`
	RepositoryIdentity string `json:"repository_identity"`
}

type UploadStsRefreshRequest struct {
	SessionIdentity string `json:"session_identity"`
}

type UploadStsRefreshResponse struct {
	SessionIdentity string     `json:"session_identity"`
	ObjectKey       string     `json:"object_key"`
	OssBucket       string     `json:"oss_bucket"`
	OssRegion       string     `json:"oss_region"`
	OssEndpoint     string     `json:"oss_endpoint"`
	Sts             *UploadSTS `json:"sts"`
}

type FolderPathItem struct {
	Id       int64  `json:"id"`
	Identity string `json:"identity"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type MailCodeRequest struct {
	Email string `json:"email"`
}

type MailCodeResponse struct {
	Code string `json:"code"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type ShareBasicCreateRequest struct {
	UserRepositoryIdentity string `json:"user_repository_identity"`
	ExpiredTime            int    `json:"expired_time"`
}

type ShareBasicCreateResponse struct {
	Identity string `json:"identity"`
}

type ShareBasicDetailRequest struct {
	Identity string `json:"identity"`
}

type ShareBasicDetailResponse struct {
	RepositoryIdentity string `json:"repository_identity"`
	Name               string `json:"name"`
	Ext                string `json:"ext"`
	Size               int64  `json:"size"`
	Path               string `json:"path"`
}

type ShareFileDetailRequest struct {
	Identity string `path:"identity"`
}

type ShareFileDetailResponse struct {
	RepositoryIdentity string `json:"repository_identity"`
	Name               string `json:"name"`
	Ext                string `json:"ext"`
	Size               int64  `json:"size"`
	Path               string `json:"path"`
}

type ShareFileSaveRequest struct {
	RepositoryIdentity string `json:"repository_identity"`
	ParentId           int64  `json:"parent_id"`
}

type ShareFileSaveResponse struct {
	Identity string `json:"identity"`
}

type UserDetailRequest struct {
	Identity string `json:"identity, optional"`
}

type UserDetailResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserFile struct {
	Id                 int64  `json:"id"`
	Identity           string `json:"identity"`
	RepositoryIdentity string `json:"repository_identity"`
	Name               string `json:"name"`
	Ext                string `json:"ext"`
	Path               string `json:"path"`
	Size               int64  `json:"size"`
	IsDir              int    `json:"is_dir"`
	IsFavorite         int    `json:"is_favorite"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	DeletedAt          string `json:"deleted_at,optional"`
	LastAccessedAt     string `json:"last_accessed_at,optional"`
}

type UserFileDeleteRequest struct {
	Identity string `json:"identity"`
}

type UserFileDeleteResponse struct{}

type UserFileListRequest struct {
	Id           int64  `form:"id,optional"`
	Identity     string `form:"identity,optional"`
	Page         int    `form:"page,optional"`
	Size         int    `form:"size,optional"`
	Query        string `form:"query,optional"`
	FileType     string `form:"file_type,optional"`
	FavoriteOnly bool   `form:"favorite_only,optional"`
	OrderBy      string `form:"order_by,optional"`
	OrderDir     string `form:"order_dir,optional"`
}

type UserFileListResponse struct {
	List  []*UserFile `json:"list"`
	Count int64       `json:"count"`
}

type UserFileMoveRequest struct {
	ParentIdentity string `json:"parent_identity"`
	Identity       string `json:"identity"`
}

type UserFileMoveResponse struct{}

type UserFileNameUpdateRequest struct {
	Name     string `json:"name"`
	Identity string `json:"identity"`
}

type UserFileNameUpdateResponse struct{}

type UserFolderChildrenRequest struct {
	Id int64 `path:"id"`
}

type UserFolderChildrenResponse struct {
	List []*UserFolderNode `json:"list"`
}

type UserFolderCreateRequest struct {
	Name     string `json:"name"`
	ParentId int64  `json:"parentId"`
}

type UserFolderCreateResponse struct {
	Identity string `json:"identity"`
}

type UserFolderNode struct {
	Id          int64  `json:"id"`
	Identity    string `json:"identity"`
	ParentId    int64  `json:"parent_id"`
	Name        string `json:"name"`
	HasChildren int    `json:"has_children"`
}

type UserFolderPathRequest struct {
	Identity string `path:"identity"`
}

type UserFolderPathResponse struct {
	List []*FolderPathItem `json:"list"`
}

type UserRegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Code     string `json:"code"`
}

type UserRegisterResponse struct{}

type UserRepositySaveRequest struct {
	ParentId         int64  `json:"parentId"`
	ReposityIdentity string `json:"repositoryIdentity"`
	Ext              string `json:"ext"`
	Name             string `json:"name"`
}

type UserRepositySaveResponse struct{}

type UserFilePreviewRequest struct {
	Identity string `path:"identity"`
}

type UserFilePreviewResponse struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Ext       string `json:"ext"`
	Size      int64  `json:"size"`
	URL       string `json:"url,optional"`
	Text      string `json:"text,optional"`
	Truncated bool   `json:"truncated"`
}

type UserRecentFileListRequest struct {
	Limit int `form:"limit,optional"`
}

type UserRecentFileListResponse struct {
	List []*UserFile `json:"list"`
}

type UserFileFavoriteRequest struct {
	Identity   string `json:"identity"`
	IsFavorite int    `json:"is_favorite"`
}

type UserFileFavoriteResponse struct{}

type UserFileBatchDeleteRequest struct {
	Identities []string `json:"identities"`
}

type UserFileBatchDeleteResponse struct{}

type UserFileBatchMoveRequest struct {
	Identities     []string `json:"identities"`
	ParentIdentity string   `json:"parent_identity"`
}

type UserFileBatchMoveResponse struct{}

type UserFileBatchFavoriteRequest struct {
	Identities []string `json:"identities"`
	IsFavorite int      `json:"is_favorite"`
}

type UserFileBatchFavoriteResponse struct{}

type UserRecycleListRequest struct {
	Page     int    `form:"page,optional"`
	Size     int    `form:"size,optional"`
	Query    string `form:"query,optional"`
	OrderBy  string `form:"order_by,optional"`
	OrderDir string `form:"order_dir,optional"`
}

type UserRecycleListResponse struct {
	List  []*UserFile `json:"list"`
	Count int64       `json:"count"`
}

type UserRecycleRestoreRequest struct {
	Identities []string `json:"identities"`
}

type UserRecycleRestoreResponse struct{}

type UserRecycleDeleteRequest struct {
	Identities []string `json:"identities"`
}

type UserRecycleDeleteResponse struct{}

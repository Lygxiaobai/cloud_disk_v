package handler

import (
	"net/http"

	"cloud_disk/core/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.ErrorRecovery},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/mail/code/send/register", Handler: MailCodeSendRegisterHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/refresh/token", Handler: RefreshTokenHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/detail", Handler: UserDetailHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/user/login", Handler: UserLoginHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/user/register", Handler: UserRegisterHandler(serverCtx)},
			}...,
		),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.ErrorRecovery, serverCtx.Auth, serverCtx.Casbin},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/file/upload", Handler: FileUploadHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/file/upload/multipart", Handler: FileUploadMultipartHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/file/upload/init", Handler: UploadInitHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/file/upload/complete", Handler: UploadCompleteHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/file/upload/sts/refresh", Handler: UploadStsRefreshHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/share/basic/create", Handler: ShareBasicCreateHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/share/file/save", Handler: ShareFileSaveHandler(serverCtx)},
				{Method: http.MethodDelete, Path: "/user/file/delete", Handler: UserFileDeleteHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/file/list", Handler: UserFileListHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/file/move", Handler: UserFileMoveHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/file/name/update", Handler: UserFileNameUpdateHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/file/preview/:identity", Handler: UserFilePreviewHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/file/favorite", Handler: UserFileFavoriteHandler(serverCtx)},
				{Method: http.MethodDelete, Path: "/user/file/batch/delete", Handler: UserFileBatchDeleteHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/file/batch/move", Handler: UserFileBatchMoveHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/file/batch/favorite", Handler: UserFileBatchFavoriteHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/file/recent", Handler: UserRecentFileListHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/folder/children/:id", Handler: UserFolderChildrenHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/folder/create", Handler: UserFolderCreateHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/folder/path/:identity", Handler: UserFolderPathHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/user/repository/save", Handler: UserRepositorySaveHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/user/recycle/list", Handler: UserRecycleListHandler(serverCtx)},
				{Method: http.MethodPut, Path: "/user/recycle/restore", Handler: UserRecycleRestoreHandler(serverCtx)},
				{Method: http.MethodDelete, Path: "/user/recycle/delete", Handler: UserRecycleDeleteHandler(serverCtx)},
			}...,
		),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.ErrorRecovery},
			[]rest.Route{
				{Method: http.MethodGet, Path: "/share/file/detail/:identity", Handler: ShareFileDetailHandler(serverCtx)},
			}...,
		),
	)
}

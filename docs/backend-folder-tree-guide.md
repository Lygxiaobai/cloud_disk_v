# 目录树后端改造说明

这份文档是给你自己手敲后端时用的。

目标：

- 左侧目录树只展示文件夹
- 中间列表继续展示当前目录下的文件夹和文件
- 顶部面包屑可以准确展示完整路径
- 后面“移动到某个目录”时，也可以复用目录树接口

这份说明会明确告诉你：

- 要改哪个文件
- 为什么要改
- 建议代码怎么写
- 注释应该怎么写

## 1. 整体思路

你现在的数据结构已经够用了：

- `user_repository.id`
- `user_repository.parent_id`
- `user_repository.identity`
- `user_repository.user_identity`
- `user_repository.is_dir`

也就是说：

- `parent_id = 0` 表示根目录下内容
- `is_dir = 1` 表示文件夹
- `is_dir = 0` 表示文件

所以目录树不需要改表结构，只需要补两个接口：

1. 获取某个目录下的直接子文件夹
2. 获取某个目录的完整路径

## 2. 你要改哪些文件

核心会改这些文件：

- [core.api](d:\Go_Project\cloud_disk\core\core.api)
- [routes.go](d:\Go_Project\cloud_disk\core\internal\handler\routes.go)
- [types.go](d:\Go_Project\cloud_disk\core\internal\types\types.go)
- 新增 `core/internal/handler/user-folder-children-handler.go`
- 新增 `core/internal/handler/user-folder-path-handler.go`
- 新增 `core/internal/logic/user-folder-children-logic.go`
- 新增 `core/internal/logic/user-folder-path-logic.go`

如果你是手敲，不用现在去跑 `goctl`，可以先按下面方式手动补齐。

## 3. 先改 API 定义

文件：

- [core.api](d:\Go_Project\cloud_disk\core\core.api)

在加了 `middleware: Auth` 的 service 里补两个接口。

### 3.1 新增接口定义

```api
@server (
    middleware: Auth
)
service core-api {
    @handler FileUpload
    post /file/upload (FileUploadRequest) returns (FileUploadResponse)

    @handler UserRepositorySave
    post /user/repository/save (UserRepositySaveRequest) returns (UserRepositySaveResponse)

    @handler UserFileList
    get /user/file/list (UserFileListRequest) returns (UserFileListResponse)

//新增
    @handler UserFolderChildren
    get /user/folder/children (UserFolderChildrenRequest) returns (UserFolderChildrenResponse)
//新增
    @handler UserFolderPath
    get /user/folder/path/:identity (UserFolderPathRequest) returns (UserFolderPathResponse)

    @handler UserFileNameUpdate
    put /user/file/name/update (UserFileNameUpdateRequest) returns (UserFileNameUpdateResponse)

    @handler UserFolderCreate
    put /user/folder/create (UserFolderCreateRequest) returns (UserFolderCreateResponse)

    @handler UserFileDelete
    delete /user/file/delete (UserFileDeleteRequest) returns (UserFileDeleteResponse)

    @handler UserFileMove
    put /user/file/move (UserFileMoveRequest) returns (UserFileMoveResponse)

    @handler ShareBasicCreate
    post /share/basic/create (ShareBasicCreateRequest) returns (ShareBasicCreateResponse)

    @handler ShareFileSave
    post /share/file/save (ShareFileSaveRequest) returns (ShareFileSaveResponse)

    @handler FileUploadMultipart
    post /file/upload/multipart (FileUploadMultipartRequest) returns (FileUploadMultipartResponse)
}
```

### 3.2 新增类型定义

```api
type UserFolderChildrenRequest {
    Id int64 `json:"id,optional"`
}

type UserFolderChildrenResponse {
    List []*UserFolderNode `json:"list"`
}

type UserFolderNode {
    Id          int64  `json:"id"`
    Identity    string `json:"identity"`
    ParentId    int64  `json:"parent_id"`
    Name        string `json:"name"`
    HasChildren int    `json:"has_children"`
}

type UserFolderPathRequest {
    Identity string `path:"identity"`
}

type UserFolderPathResponse {
    List []*FolderPathItem `json:"list"`
}

type FolderPathItem {
    Id       int64  `json:"id"`
    Identity string `path:"identity"`
    Name     string `json:"name"`
}
```

说明：

- `UserFolderChildrenRequest.Id`
  表示当前要查询哪个目录的孩子
  `0` 表示根目录
- `HasChildren`
  表示这个文件夹下面是否还有子文件夹
  前端目录树会根据它决定是否显示展开箭头
- `UserFolderPathRequest.Identity`
  表示当前目录是谁
- `UserFolderPathResponse.List`
  返回从根到当前目录的完整路径

## 4. 改 routes.go

文件：

- [routes.go](d:\Go_Project\cloud_disk\core\internal\handler\routes.go)

在鉴权路由中补两个 handler：

```go
{
    Method:  http.MethodGet,
    Path:    "/user/folder/children",
    Handler: UserFolderChildrenHandler(serverCtx),
},
{
    Method:  http.MethodGet,
    Path:    "/user/folder/path/:identity",
    Handler: UserFolderPathHandler(serverCtx),
},
```

放在 `UserFileList` 附近最合适，因为语义上都属于目录浏览能力。

## 5. 新增 children handler

文件：

- `core/internal/handler/user-folder-children-handler.go`

建议代码：

```go
package handler

import (
    "net/http"

    "cloud_disk/core/internal/logic"
    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/rest/httpx"
)

func UserFolderChildrenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.UserFolderChildrenRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        l := logic.NewUserFolderChildrenLogic(r.Context(), svcCtx)
        resp, err := l.UserFolderChildren(&req, r.Header.Get("UserIdentity"))
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
```

## 6. 新增 path handler

文件：

- `core/internal/handler/user-folder-path-handler.go`

建议代码：

```go
package handler

import (
    "net/http"

    "cloud_disk/core/internal/logic"
    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/rest/httpx"
)

func UserFolderPathHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.UserFolderPathRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        l := logic.NewUserFolderPathLogic(r.Context(), svcCtx)
        resp, err := l.UserFolderPath(&req, r.Header.Get("UserIdentity"))
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
```

## 7. 新增 children logic

文件：

- `core/internal/logic/user-folder-children-logic.go`

建议代码如下，注释我已经写进去了：

```go
package logic

import (
    "context"

    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/core/logx"
)

type UserFolderChildrenLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewUserFolderChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderChildrenLogic {
    return &UserFolderChildrenLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *UserFolderChildrenLogic) UserFolderChildren(req *types.UserFolderChildrenRequest, userIdentity string) (resp *types.UserFolderChildrenResponse, err error) {
    list := make([]*types.UserFolderNode, 0)

    // 这里只查询“当前目录下的直接子文件夹”，不查文件。
    // 左侧目录树的职责是目录导航，不是文件内容展示。
    sql := `
SELECT
    ur.id,
    ur.identity,
    ur.parent_id,
    ur.name,
    CASE WHEN EXISTS (
        SELECT 1
        FROM user_repository child
        WHERE child.parent_id = ur.id
          AND child.user_identity = ur.user_identity
          AND child.is_dir = 1
          AND child.deleted_at IS NULL
    ) THEN 1 ELSE 0 END AS has_children
FROM user_repository ur
WHERE ur.user_identity = ?
  AND ur.parent_id = ?  
  AND ur.is_dir = 1
  AND ur.deleted_at IS NULL
ORDER BY ur.id ASC
`

    err = l.svcCtx.Engine.SQL(sql, userIdentity, req.Id).Find(&list)
    if err != nil {
        return nil, err
    }

    resp = &types.UserFolderChildrenResponse{
        List: list,
    }
    return
}
```

说明：

- 这里只返回直接子文件夹
- `HasChildren` 是通过 `EXISTS` 算出来的
- 左侧树展开时，只要 `HasChildren = 1`，前端就会继续请求下一层

## 8. 新增 path logic

文件：

- `core/internal/logic/user-folder-path-logic.go`

建议代码如下：

```go
package logic

import (
    "context"
    "errors"

    "cloud_disk/core/internal/models"
    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/core/logx"
)

type UserFolderPathLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewUserFolderPathLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderPathLogic {
    return &UserFolderPathLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *UserFolderPathLogic) UserFolderPath(req *types.UserFolderPathRequest, userIdentity string) (resp *types.UserFolderPathResponse, err error) {
    // 根目录路径固定返回“全部文件”。
    if req.Identity == "" {
        return &types.UserFolderPathResponse{
            List: []*types.FolderPathItem{
                {
                    Id:       0,
                    Identity: "",
                    Name:     "全部文件",
                },
            },
        }, nil
    }

    current := new(models.UserRepository)
    has, err := l.svcCtx.Engine.
        Where("identity = ? AND user_identity = ? AND is_dir = 1", req.Identity, userIdentity).
        Get(current)
    if err != nil {
        return nil, err
    }
    if !has {
        return nil, errors.New("folder not found")
    }

    path := make([]*types.FolderPathItem, 0)

    // 从当前目录一路向上找父目录，直到根目录。
    for {
        path = append(path, &types.FolderPathItem{
            Id:       int64(current.Id),
            Identity: current.Identity,
            Name:     current.Name,
        })

        if current.ParentId == 0 {
            break
        }

        parent := new(models.UserRepository)
        has, err = l.svcCtx.Engine.
            Where("id = ? AND user_identity = ? AND is_dir = 1", current.ParentId, userIdentity).
            Get(parent)
        if err != nil {
            return nil, err
        }
        if !has {
            return nil, errors.New("parent folder not found")
        }

        current = parent
    }

    // 当前 path 是从“当前目录 -> 父目录 -> 根目录”，要反转成“根目录 -> 当前目录”。
    for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
        path[left], path[right] = path[right], path[left]
    }

    // 在最前面补一个虚拟根节点，前端面包屑展示更统一。
    path = append([]*types.FolderPathItem{
        {
            Id:       0,
            Identity: "",
            Name:     "全部文件",
        },
    }, path...)

    resp = &types.UserFolderPathResponse{
        List: path,
    }
    return
}
```

说明：

- 这个接口是给面包屑用的
- 前端进入任何目录时，都可以调一次这个接口，拿到完整层级
- 这样就算目录名重复，比如 `jetbra / jetbra / config-jetbrains`，也能展示正确顺序

## 9. types.go 要补什么

文件：

- [types.go](d:\Go_Project\cloud_disk\core\internal\types\types.go)

如果你不用 `goctl` 自动生成，就要手动把这几个结构体补进去：

- `UserFolderChildrenRequest`
- `UserFolderChildrenResponse`
- `UserFolderNode`
- `UserFolderPathRequest`
- `UserFolderPathResponse`
- `FolderPathItem`

字段就按 `core.api` 里写的抄过去就行。

## 10. 这两个接口前端会怎么用

### 10.1 左侧目录树

前端会这样调：

- 初次加载左侧树：`GET /user/folder/children?id=0`
- 展开某个目录：`GET /user/folder/children?id=当前目录id`

### 10.2 顶部面包屑

前端会这样调：

- 点击进入某个目录时：`GET /user/folder/path/目录identity`

#
## 10.4 进入文件夹查看该目录下所有文件和文件夹

你现在其实已经有一个接近的接口：

- `GET /user/file/list?id=目录id&page=1&size=20`

它本质上就是：

- 根据 `parent_id` 查询当前目录下的所有内容
- 返回文件夹和文件

但是它有一个不太顺手的地方：

- 前端需要先知道目录的数据库 `id`
- 如果前端更自然地拿到的是目录 `identity`，还要先做一次转换

所以你可以再补一个“目录内容接口”，直接按目录 `identity` 进入目录查看内容。

推荐接口：

```api
@handler UserFolderContent
get /user/folder/content/:identity (UserFolderContentRequest) returns (UserFolderContentResponse)
```

### 10.4.1 这个接口的用途

这个接口的职责是：

- 进入某个文件夹
- 返回该文件夹下的所有直接子内容
- 包括文件夹
- 也包括文件

举例：

```text
全部文件
└─ jetbra
   └─ config-jetbrains
      ├─ dns.conf
      ├─ power.conf
      └─ url.conf
```

当你进入 `config-jetbrains` 时，这个接口应该返回：

- `dns.conf`
- `power.conf`
- `url.conf`

如果当前目录下还有子文件夹，也一并返回。

### 10.4.2 要改哪些文件

你需要再补这些文件或内容：

- [core.api](d:\Go_Project\cloud_disk\core\core.api)
- [routes.go](d:\Go_Project\cloud_disk\core\internal\handler\routes.go)
- [types.go](d:\Go_Project\cloud_disk\core\internal\types\types.go)
- 新增 `core/internal/handler/user-folder-content-handler.go`
- 新增 `core/internal/logic/user-folder-content-logic.go`

### 10.4.3 core.api 里新增定义

```api
type UserFolderContentRequest {
    Identity string `path:"identity"`
    Page     int    `json:"page,optional"`
    Size     int    `json:"size,optional"`
}

type UserFolderContentResponse {
    List  []*UserFile `json:"list"`
    Count int64       `json:"count"`
}
```

然后在鉴权 service 里新增：

```api
@handler UserFolderContent
get /user/folder/content/:identity (UserFolderContentRequest) returns (UserFolderContentResponse)
```

### 10.4.4 routes.go 新增路由

```go
{
    Method:  http.MethodGet,
    Path:    "/user/folder/content/:identity",
    Handler: UserFolderContentHandler(serverCtx),
},
```

### 10.4.5 handler 建议代码

文件：

- `core/internal/handler/user-folder-content-handler.go`

```go
package handler

import (
    "net/http"

    "cloud_disk/core/internal/logic"
    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/rest/httpx"
)

func UserFolderContentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.UserFolderContentRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        l := logic.NewUserFolderContentLogic(r.Context(), svcCtx)
        resp, err := l.UserFolderContent(&req, r.Header.Get("UserIdentity"))
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
```

### 10.4.6 logic 建议代码

文件：

- `core/internal/logic/user-folder-content-logic.go`

```go
package logic

import (
    "context"
    "errors"

    "cloud_disk/core/internal/define"
    "cloud_disk/core/internal/models"
    "cloud_disk/core/internal/svc"
    "cloud_disk/core/internal/types"
    "github.com/zeromicro/go-zero/core/logx"
)

type UserFolderContentLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewUserFolderContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderContentLogic {
    return &UserFolderContentLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *UserFolderContentLogic) UserFolderContent(req *types.UserFolderContentRequest, userIdentity string) (resp *types.UserFolderContentResponse, err error) {
    // 先根据目录 identity 找到它自己的数据库 id。
    // 因为子内容查询最终还是要落到 parent_id 上。
    folder := new(models.UserRepository)
    has, err := l.svcCtx.Engine.
        Where("identity = ? AND user_identity = ? AND is_dir = 1", req.Identity, userIdentity).
        Get(folder)
    if err != nil {
        return nil, err
    }
    if !has {
        return nil, errors.New("folder not found")
    }

    list := make([]*types.UserFile, 0)

    size := req.Size
    if size == 0 {
        size = define.PageSize
    }

    page := req.Page
    if page == 0 {
        page = define.Page
    }

    offset := (page - 1) * size

    // 查询当前目录下的所有直接子内容：包括文件夹和文件。
    err = l.svcCtx.Engine.Table("user_repository").
        Where("user_identity = ? AND parent_id = ?", userIdentity, folder.Id).
        Select("user_repository.id,user_repository.identity,user_repository.repository_identity,user_repository.name,user_repository.ext,user_repository.is_dir,repository_pool.size,repository_pool.path").
        Join("LEFT", "repository_pool", "user_repository.repository_identity = repository_pool.identity").
        Limit(size, offset).
        Where("user_repository.deleted_at IS NULL").
        Find(&list)
    if err != nil {
        return nil, err
    }

    count, err := l.svcCtx.Engine.
        Where("user_identity = ? AND parent_id = ?", userIdentity, folder.Id).
        Count(&models.UserRepository{})
    if err != nil {
        return nil, err
    }

    resp = &types.UserFolderContentResponse{
        List:  list,
        Count: count,
    }
    return
}
```

### 10.4.7 这个接口和 `/user/file/list` 的区别

- `/user/file/list`
  用 `id` 查
  更像底层通用接口

- `/user/folder/content/:identity`
  用目录 `identity` 查
  更像前端直接进入目录时用的接口

如果你只想保留一套接口，也可以不新增，继续用 `/user/file/list`。
但如果你想让“进入文件夹”这件事语义更清晰，我建议补这个接口。

### 10.4.8 前端怎么用

进入某个文件夹时：

```text
GET /user/folder/content/:identity
```

例如：

```text
GET /user/folder/content/401179e5-d6e2-41e7-bca6-742a60a44cea
```

返回后：

- 中间列表显示该目录下所有文件和文件夹
- 左侧目录树继续只显示文件夹
- 顶部面包屑继续用 `/user/folder/path/:identity`
```
## 10.3 中间文件列表

这个接口保持你现在这套：

- `GET /user/file/list?id=目录id&page=1&size=20`

所以三块区域会各司其职：

- 左侧目录树：只显示文件夹层级
- 顶部面包屑：显示完整路径
- 中间列表：显示当前目录下所有内容

## 11. 建议顺手补的后端校验

你现在做目录树时，建议把这些也一起补上：

### 11.1 移动文件夹时，禁止移动到自己里面

文件：

- [user-file-move-logic.go](d:\Go_Project\cloud_disk\core\internal\logic\user-file-move-logic.go)

至少要防：

- A 移到 A 自己下面
- A 移到 A 的子孙目录下面

不然目录树一旦做出来，就可能产生循环结构。

### 11.2 查询目录路径时校验归属

已经在上面的逻辑里写了：

- `user_identity = ?`

这个不能少，不然用户可能查到别人的目录路径。

### 11.3 同一父目录下不能重名

你创建文件夹时已经做了这层校验，这个思路继续保持即可。

## 12. 建议你的手敲顺序

建议按这个顺序来敲：

1. 先改 `core/core.api`
2. 手动补 `types.go`
3. 新建两个 handler 文件
4. 新建两个 logic 文件
5. 改 `routes.go`
6. 启动后端测试这两个接口

## 13. 你可以先这样测试

### 查根目录子文件夹

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/folder/children?id=0' -Headers @{ Authorization = '你的token' }
```

### 查某个目录完整路径

```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:8888/user/folder/path/某个目录identity' -Headers @{ Authorization = '你的token' }
```

## 14. 最后一句建议

如果你是第一次自己手敲这一块，我建议你先只把：

- `/user/folder/children`
- `/user/folder/path`

做通，不要一上来就一起改移动逻辑。

先让：

- 左侧目录树能展开
- 顶部面包屑能正确显示

等这两块跑稳了，再去加强 `move` 的防环路校验，会更顺。
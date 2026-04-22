# Cloud Disk Feature Flow (ASCII)

This version does not depend on Mermaid. Any Markdown viewer can display it.

## 1. Upload Flow

```text
+----------------------+
| Frontend picks file  |
+----------------------+
           |
           v
+----------------------+
| Frontend computes    |
| MD5 / hash           |
+----------------------+
           |
           v
+----------------------+
| POST /file/upload/   |
| init                 |
+----------------------+
           |
           v
+----------------------+
| Backend checks       |
| repository_pool by   |
| hash + size          |
+----------------------+
     | Yes                          | No
     v                              v
+----------------------+   +----------------------+
| Insert               |   | Create upload_      |
| user_repository      |   | session             |
| only                 |   +----------------------+
+----------------------+              |
     |                                v
     v                     +----------------------+
+----------------------+   | Generate objectKey   |
| Return instant_hit   |   +----------------------+
| = true               |              |
+----------------------+              v
     |                     +----------------------+
     v                     | Call STS AssumeRole  |
+----------------------+   +----------------------+
| Upload done          |              |
+----------------------+              v
                            +----------------------+
                            | Return STS +         |
                            | sessionIdentity +    |
                            | objectKey            |
                            +----------------------+
                                       |
                                       v
                            +----------------------+
                            | Frontend uses        |
                            | ali-oss multipart    |
                            | upload to OSS        |
                            +----------------------+
                                       |
                                       v
                            +----------------------+
                            | STS about to expire? |
                            +----------------------+
                               | Yes        | No
                               v            v
                    +------------------+   +----------------------+
                    | POST /file/      |   | Upload finished      |
                    | upload/sts/      |   +----------------------+
                    | refresh          |              |
                    +------------------+              v
                               |            +----------------------+
                               v            | POST /file/upload/   |
                    +------------------+    | complete             |
                    | Return new STS   |    +----------------------+
                    +------------------+              |
                               |                      v
                               +-----------> +----------------------+
                                            | HeadObject verify     |
                                            | uploaded size         |
                                            +----------------------+
                                                       |
                                                       v
                                            +----------------------+
                                            | Write repository_    |
                                            | pool                 |
                                            +----------------------+
                                                       |
                                                       v
                                            +----------------------+
                                            | Write user_          |
                                            | repository           |
                                            +----------------------+
                                                       |
                                                       v
                                            +----------------------+
                                            | Mark upload_session  |
                                            | completed            |
                                            +----------------------+
                                                       |
                                                       v
                                            +----------------------+
                                            | Frontend shows       |
                                            | success              |
                                            +----------------------+
```

## 2. Preview And Recent Files Flow

```text
+----------------------+
| Open disk page       |
+----------------------+
           |
           v
+----------------------+
| GET /user/file/list  |
+----------------------+
           |
           v
+----------------------+
| Backend applies      |
| search / filter /    |
| sort                 |
+----------------------+
           |
           v
+----------------------+
| Render file list     |
+----------------------+
           |
           v
+----------------------+
| Click preview        |
+----------------------+
           |
           v
+---------------------------+
| GET /user/file/preview/   |
| :identity                 |
+---------------------------+
           |
           v
+----------------------+
| Backend checks type  |
+----------------------+
   | image/video/audio/pdf
   |----------------------------> signed OSS URL
   |
   | text
   |----------------------------> read first 4KB from OSS
   |
   | other
   |----------------------------> download URL
           |
           v
+----------------------+
| Preview dialog       |
+----------------------+
           |
           v
+----------------------+
| Write recent file    |
| into Redis ZSet      |
+----------------------+
           |
           v
+----------------------+
| GET /user/file/      |
| recent               |
+----------------------+
           |
           v
+----------------------+
| Show recent files    |
+----------------------+
```

## 3. Search / Filter / Sort Flow

```text
+----------------------+
| Frontend changes     |
| query / file type /  |
| favorite / sort      |
+----------------------+
           |
           v
+----------------------+
| GET /user/file/list  |
+----------------------+
           |
           v
+----------------------+
| Backend builds SQL   |
| with:                |
| - query              |
| - file_type          |
| - favorite_only      |
| - order_by           |
| - order_dir          |
+----------------------+
           |
           v
+----------------------+
| Return filtered list |
+----------------------+
```

## 4. Batch Operation Flow

```text
+----------------------+
| User selects files   |
+----------------------+
           |
           v
+----------------------+
| Choose batch action  |
+----------------------+
   | batch favorite
   |----------------------------> update is_favorite
   |
   | batch move
   |----------------------------> validate target folder
   |                              validate not self/child
   |                              update parent_id
   |
   | batch delete
   |----------------------------> collect subtree
                                  set deleted_at
```

## 5. Recycle Bin Flow

```text
+----------------------+
| Delete file/folder   |
+----------------------+
           |
           v
+----------------------+
| Soft delete by       |
| writing deleted_at   |
+----------------------+
           |
           v
+----------------------+
| GET /user/recycle/   |
| list                 |
+----------------------+
           |
           v
+----------------------+
| Show recycle bin     |
+----------------------+
     | Restore                      | Permanent delete
     v                              v
+----------------------+   +----------------------+
| Check parent folder  |   | Physical DELETE      |
| already restored     |   | from database        |
+----------------------+   +----------------------+
     | Yes      | No
     v          v
+-----------+  +----------------------+
| Clear     |  | Reject restore       |
| deleted_at|  | child first          |
+-----------+  +----------------------+
```


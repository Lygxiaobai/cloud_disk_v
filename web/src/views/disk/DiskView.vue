<template>
  <div class="page-shell bd-page">
    <div class="bd-shell">
      <aside class="bd-sidebar">
        <div class="brand">
          <div class="brand-icon">盘</div>
          <div>
            <strong>我的网盘</strong>
            <p>仿百度网盘布局</p>
          </div>
        </div>

        <input ref="fileInputRef" class="hidden-input" type="file" multiple @change="handleFilePicked" />
        <el-button class="side-btn" type="primary" @click="triggerUpload">上传文件</el-button>
        <el-button class="side-btn" @click="handleCreateFolder">新建文件夹</el-button>

        <div class="side-nav">
          <button class="nav-btn" :class="{ active: activeNav === 'files' }" type="button" @click="openSidebarView('files')">全部文件</button>
          <button class="nav-btn" :class="{ active: activeNav === 'favorite' }" type="button" @click="openSidebarView('favorite')">我的收藏</button>
          <button class="nav-btn" :class="{ active: activeNav === 'recent' }" type="button" @click="openSidebarView('recent')">最近文件</button>
          <button class="nav-btn" :class="{ active: activeNav === 'recycle' }" type="button" @click="openSidebarView('recycle')">回收站</button>
        </div>

        <div class="panel side-card">
          <div class="card-head">
            <span>目录树</span>
            <button class="text-btn" type="button" @click="diskStore.loadRoot">根目录</button>
          </div>
          <el-tree
            v-if="folderTreeSupported"
            node-key="identity"
            lazy
            :load="loadTreeNode"
            :props="treeProps"
            :current-node-key="diskStore.currentFolderIdentity"
            highlight-current
            @node-click="handleTreeNodeClick"
          />
          <el-alert v-else :closable="false" :title="folderTreeMessage" show-icon type="info" />
        </div>

        <div class="panel side-card">
          <div class="card-head">
            <span>最近访问</span>
            <button class="text-btn" type="button" @click="openSidebarView('recent')">更多</button>
          </div>
          <el-empty v-if="diskStore.recentItems.length === 0" description="暂无记录" />
          <button v-for="item in quickRecentItems" :key="item.identity" class="recent-link" type="button" @click="handleOpen(item)">
            <strong>{{ item.name }}</strong>
            <span>{{ item.last_accessed_at || item.updated_at || "-" }}</span>
          </button>
        </div>
      </aside>

      <main class="bd-main">
        <header class="panel topbar">
          <div>
            <span class="kicker">个人云存储中心</span>
            <h1>{{ pageHeading }}</h1>
            <el-breadcrumb v-if="diskStore.viewMode === 'files'" separator="/">
              <el-breadcrumb-item v-for="(crumb, index) in diskStore.breadcrumbs" :key="`${crumb.id}-${index}`">
                <button class="crumb-btn" type="button" @click="diskStore.jumpToBreadcrumb(index)">{{ crumb.name }}</button>
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <div class="top-tools">
            <el-input
              v-model="diskStore.query"
              class="search-box"
              clearable
              placeholder="搜索网盘文件"
              @clear="diskStore.applyFilters"
              @keyup.enter="diskStore.applyFilters"
            />
            <div class="account-box">
              <span>当前账号</span>
              <strong>{{ authStore.profile?.email || authStore.identity || "未同步邮箱" }}</strong>
            </div>
            <el-button @click="handleLogout">退出</el-button>
          </div>
        </header>

        <section class="panel main-card">
          <div class="toolbar">
            <div class="toolbar-left">
              <el-button type="primary" @click="triggerUpload">上传</el-button>
              <el-button @click="handleCreateFolder">新建文件夹</el-button>
              <el-button @click="diskStore.refresh">刷新</el-button>
            </div>
            <div v-if="selectedFiles.length > 0" class="toolbar-right">
              <template v-if="diskStore.viewMode === 'recycle'">
                <el-button type="primary" plain @click="handleBatchRestore">批量恢复</el-button>
                <el-button type="danger" plain @click="handleBatchDestroy">彻底删除</el-button>
              </template>
              <template v-else>
                <el-button type="warning" plain @click="handleBatchFavorite(1)">批量收藏</el-button>
                <el-button plain @click="handleBatchFavorite(0)">取消收藏</el-button>
                <el-button type="primary" plain @click="openMoveDialog()">批量移动</el-button>
                <el-button type="danger" plain @click="handleBatchDelete">批量删除</el-button>
              </template>
            </div>
          </div>

          <div class="filters">
            <el-select v-if="diskStore.viewMode !== 'recycle'" v-model="diskStore.fileType" placeholder="文件类型">
              <el-option label="全部类型" value="all" />
              <el-option label="文件夹" value="dir" />
              <el-option label="图片" value="image" />
              <el-option label="视频" value="video" />
              <el-option label="音频" value="audio" />
              <el-option label="文档" value="document" />
              <el-option label="压缩包" value="archive" />
              <el-option label="代码" value="code" />
              <el-option label="其他" value="other" />
            </el-select>
            <el-select v-model="diskStore.orderBy" placeholder="排序字段">
              <el-option label="默认排序" value="" />
              <el-option label="修改时间" value="updated_at" />
              <el-option label="创建时间" value="created_at" />
              <el-option label="文件大小" value="size" />
              <el-option label="名称" value="name" />
              <el-option v-if="diskStore.viewMode === 'recycle'" label="删除时间" value="deleted_at" />
            </el-select>
            <el-select v-model="diskStore.orderDir" placeholder="排序方向">
              <el-option label="降序" value="desc" />
              <el-option label="升序" value="asc" />
            </el-select>
            <el-switch v-if="diskStore.viewMode !== 'recycle'" v-model="diskStore.favoriteOnly" active-text="仅收藏" />
            <el-button type="primary" plain @click="diskStore.applyFilters">应用筛选</el-button>
            <el-button @click="handleResetFilters">重置</el-button>
          </div>

          <div v-if="uploadStore.tasks.length > 0" class="task-strip">
            <article v-for="task in uploadStore.tasks" :key="task.id" class="task-chip">
              <div class="task-chip-top">
                <strong>{{ task.name }}</strong>
                <span>{{ taskStatusText(task.status) }}</span>
              </div>
              <el-progress
                :percentage="Math.max(1, Math.round((task.status === 'hashing' ? task.hashProgress : task.progress) * 100))"
                :status="task.status === 'error' ? 'exception' : task.status === 'success' ? 'success' : undefined"
                :stroke-width="6"
              />
              <div class="task-chip-bottom">
                <span>{{ formatFileSize(task.size) }}</span>
                <span v-if="task.status === 'uploading' && task.etaSeconds !== null">剩余 {{ formatDuration(task.etaSeconds) }}</span>
                <el-button v-if="task.status === 'uploading'" link type="warning" @click="uploadStore.pauseTask(task.id)">暂停</el-button>
                <el-button v-if="task.status === 'paused'" link type="primary" @click="uploadStore.resumeTask(task.id, refreshAfterMutation)">继续</el-button>
                <el-button v-if="task.status === 'error'" link type="primary" @click="uploadStore.retryTask(task.id, currentUploadContext, refreshAfterMutation)">重试</el-button>
              </div>
            </article>
          </div>

          <div class="table-head">
            <strong>{{ pageHeading }}</strong>
            <span>共 {{ diskStore.total }} 项，已选 {{ selectedFiles.length }} 项</span>
          </div>

          <FileTable
            :files="diskStore.visibleItems"
            :loading="diskStore.loading"
            :view-mode="diskStore.viewMode"
            @delete="handleDelete"
            @destroy="handleDestroy"
            @move="openMoveDialog"
            @open="handleOpen"
            @rename="handleRename"
            @restore="handleRestore"
            @selection-change="selectedFiles = $event"
            @share="openShareDialog"
            @toggle-favorite="handleToggleFavorite"
          />

          <div v-if="diskStore.viewMode !== 'recent'" class="pagination-row">
            <el-pagination
              :current-page="diskStore.page"
              :page-size="diskStore.size"
              :total="diskStore.total"
              background
              layout="prev, pager, next, total"
              @current-change="diskStore.changePage"
            />
          </div>
        </section>
      </main>
    </div>

    <el-dialog v-model="moveDialogVisible" title="移动到文件夹" width="520px">
      <p class="muted dialog-tip">可以移动到根目录，也可以选择一个具体文件夹作为目标位置。</p>
      <div class="root-option">
        <el-button :type="moveTargetIdentity === '' ? 'primary' : 'default'" plain @click="selectMoveRoot">移动到根目录</el-button>
      </div>
      <el-tree v-if="folderTreeSupported" node-key="identity" lazy :load="loadTreeNode" :props="treeProps" highlight-current @node-click="handleMoveTargetSelect" />
      <el-alert v-else :closable="false" :title="folderTreeMessage" show-icon type="info" />
      <div class="move-selection panel">
        <span class="muted">当前目标</span>
        <strong>{{ moveTargetName }}</strong>
        <span class="soft-code">{{ moveTargetIdentity || "root" }}</span>
      </div>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleMove">确认移动</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="shareDialogVisible" title="创建分享" width="520px">
      <div class="share-panel">
        <p class="muted">为 <strong>{{ shareTarget?.name }}</strong> 生成一个可访问的分享链接。</p>
        <el-form label-position="top">
          <el-form-item label="分享有效期">
            <el-select v-model="shareExpiredTime" style="width: 100%">
              <el-option :value="3600" label="1 小时" />
              <el-option :value="86400" label="1 天" />
              <el-option :value="604800" label="7 天" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="shareLink" label="分享链接">
            <el-input :model-value="shareLink" readonly />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="shareDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleCreateShare">{{ shareLink ? "重新生成" : "生成分享链接" }}</el-button>
      </template>
    </el-dialog>

    <FilePreviewDialog v-model="previewDialogVisible" :preview="previewData" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useRouter } from "vue-router";
import FilePreviewDialog from "@/components/FilePreviewDialog.vue";
import FileTable from "@/components/FileTable.vue";
import * as diskApi from "@/api/disk";
import type { FilePreviewResponse, FolderTreeNode, UserFile } from "@/types/api";
import { useAuthStore } from "@/stores/auth";
import { useDiskStore } from "@/stores/disk";
import { useUploadStore } from "@/stores/upload";
import { getErrorMessage } from "@/utils/error";

const router = useRouter();
const authStore = useAuthStore();
const diskStore = useDiskStore();
const uploadStore = useUploadStore();
const fileInputRef = ref<HTMLInputElement | null>(null);
const selectedFiles = ref<UserFile[]>([]);
const moveDialogVisible = ref(false);
const moveCandidate = ref<UserFile | null>(null);
const moveTargetIdentity = ref("");
const moveTargetName = ref("根目录");
const shareDialogVisible = ref(false);
const shareTarget = ref<UserFile | null>(null);
const shareExpiredTime = ref(86400);
const shareLink = ref("");
const folderTreeSupported = ref(true);
const folderTreeMessage = ref("目录树接口加载失败，请先检查 /user/folder/children 和 /user/folder/path。");
const previewDialogVisible = ref(false);
const previewData = ref<FilePreviewResponse | null>(null);
const treeProps = { label: "name", isLeaf: (data: FolderTreeNode) => data.has_children !== 1 };
const quickRecentItems = computed(() => diskStore.recentItems.slice(0, 4));
const activeNav = computed(() => (diskStore.viewMode === "recent" ? "recent" : diskStore.viewMode === "recycle" ? "recycle" : diskStore.favoriteOnly ? "favorite" : "files"));
const pageHeading = computed(() => (activeNav.value === "favorite" ? "收藏夹" : activeNav.value === "recent" ? "最近文件" : activeNav.value === "recycle" ? "回收站" : diskStore.currentFolderName));
const currentUploadContext = computed(() => ({ parentId: diskStore.currentFolderId, parentIdentity: diskStore.currentFolderIdentity || undefined }));

onMounted(async () => { await authStore.bootstrap(); await diskStore.loadRoot(); });

async function openSidebarView(target: "files" | "favorite" | "recent" | "recycle") {
  selectedFiles.value = [];
  if (target === "favorite") { if (diskStore.viewMode !== "files") await diskStore.setViewMode("files"); diskStore.favoriteOnly = true; await diskStore.applyFilters(); return; }
  if (target === "files") { if (diskStore.viewMode !== "files") await diskStore.setViewMode("files"); diskStore.favoriteOnly = false; await diskStore.applyFilters(); return; }
  diskStore.favoriteOnly = false; await diskStore.setViewMode(target);
}
function triggerUpload() { if (diskStore.viewMode !== "files") { ElMessage.info("请先切回文件视图后再上传"); return; } fileInputRef.value?.click(); }
function formatFileSize(size: number) { if (!size) return "0 B"; const units = ["B", "KB", "MB", "GB", "TB"]; let value = size; let index = 0; while (value >= 1024 && index < units.length - 1) { value /= 1024; index += 1; } return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`; }
function formatSpeed(bytes: number) { return `${formatFileSize(bytes)}/s`; }
function formatDuration(seconds: number) { if (seconds < 60) return `${Math.ceil(seconds)} 秒`; const minutes = Math.floor(seconds / 60); return `${minutes} 分 ${Math.ceil(seconds % 60)} 秒`; }
function taskStatusText(status: string) { return ({ hashing: "计算中", waiting: "等待上传", uploading: "上传中", paused: "已暂停", completing: "写入网盘", success: "已完成", error: "失败" } as Record<string, string>)[status] || status; }

async function loadTreeNode(node: any, resolve: (data: FolderTreeNode[]) => void) {
  const folderId = node.level === 0 ? 0 : Number(node.data.id);
  try { const data = await diskApi.listFolderChildren(folderId); folderTreeSupported.value = true; resolve(data.list); }
  catch (error) { if (node.level === 0) { folderTreeSupported.value = false; folderTreeMessage.value = getErrorMessage(error, "目录树接口未完成。"); } else { ElMessage.error(getErrorMessage(error, "加载子目录失败")); } resolve([]); }
}
async function handleTreeNodeClick(folder: FolderTreeNode) { if (diskStore.viewMode !== "files") diskStore.favoriteOnly = false; await diskStore.openFolder(folder); }
async function refreshAfterMutation() { await diskStore.refresh(); await diskStore.loadRecent(); }
async function handleFilePicked(event: Event) { const target = event.target as HTMLInputElement; const files = Array.from(target.files ?? []); if (files.length === 0) return; try { await uploadStore.addFiles(files, currentUploadContext.value, refreshAfterMutation); ElMessage.success(`已加入 ${files.length} 个上传任务`); } catch (error) { ElMessage.error(getErrorMessage(error, "创建上传任务失败")); } finally { target.value = ""; } }

async function handleCreateFolder() {
  try { const { value } = await ElMessageBox.prompt("输入新的文件夹名称", "新建文件夹", { confirmButtonText: "创建", inputErrorMessage: "请输入文件夹名称", inputPattern: /\S+/ }); await diskApi.createFolder({ name: value, parentId: diskStore.currentFolderId }); ElMessage.success("文件夹创建成功"); await refreshAfterMutation(); }
  catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "创建文件夹失败")); }
}
async function handleRename(file: UserFile) {
  try { const { value } = await ElMessageBox.prompt("输入新的名称", "重命名", { confirmButtonText: "保存", inputPattern: /\S+/, inputValue: file.name }); await diskApi.renameFile({ identity: file.identity, name: value }); ElMessage.success("重命名成功"); await refreshAfterMutation(); }
  catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "重命名失败")); }
}
async function handleDelete(file: UserFile) {
  try { await ElMessageBox.confirm(`确认删除 ${file.name} 吗？`, "删除确认", { confirmButtonText: "删除", type: "warning" }); await diskApi.deleteFile({ identity: file.identity }); ElMessage.success("已移入回收站"); await refreshAfterMutation(); }
  catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "删除失败")); }
}
async function handleBatchDelete() {
  try { await ElMessageBox.confirm(`确认删除选中的 ${selectedFiles.value.length} 项吗？`, "批量删除", { confirmButtonText: "删除", type: "warning" }); await diskApi.batchDeleteFiles({ identities: selectedFiles.value.map((item) => item.identity) }); selectedFiles.value = []; ElMessage.success("已移入回收站"); await refreshAfterMutation(); }
  catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "批量删除失败")); }
}
async function handleOpen(file: UserFile) {
  if (file.is_dir === 1) { diskStore.favoriteOnly = false; await diskStore.openFolder(file); return; }
  try { previewData.value = await diskApi.previewFile(file.identity); previewDialogVisible.value = true; void diskStore.loadRecent(); }
  catch (error) { ElMessage.error(getErrorMessage(error, "加载预览失败")); }
}
function openMoveDialog(file?: UserFile) { if (diskStore.viewMode === "recycle") return; moveCandidate.value = file ?? null; moveTargetIdentity.value = ""; moveTargetName.value = "根目录"; moveDialogVisible.value = true; }
function handleMoveTargetSelect(folder: FolderTreeNode) { if (moveCandidate.value?.identity === folder.identity) { ElMessage.warning("不能移动到自己里面"); return; } moveTargetIdentity.value = folder.identity; moveTargetName.value = folder.name; }
function selectMoveRoot() { moveTargetIdentity.value = ""; moveTargetName.value = "根目录"; }
async function handleMove() {
  try { if (moveCandidate.value) await diskApi.moveFile({ identity: moveCandidate.value.identity, parent_identity: moveTargetIdentity.value }); else { await diskApi.batchMoveFiles({ identities: selectedFiles.value.map((item) => item.identity), parent_identity: moveTargetIdentity.value }); selectedFiles.value = []; } ElMessage.success("移动成功"); moveDialogVisible.value = false; await refreshAfterMutation(); }
  catch (error) { ElMessage.error(getErrorMessage(error, "移动失败")); }
}
function openShareDialog(file: UserFile) { shareTarget.value = file; shareExpiredTime.value = 86400; shareLink.value = ""; shareDialogVisible.value = true; }
async function handleCreateShare() { if (!shareTarget.value) return; try { const data = await diskApi.createShare({ expired_time: shareExpiredTime.value, user_repository_identity: shareTarget.value.identity }); shareLink.value = `${window.location.origin}/share/${data.identity}`; ElMessage.success("分享链接已生成"); } catch (error) { ElMessage.error(getErrorMessage(error, "创建分享失败")); } }
async function handleToggleFavorite(file: UserFile) { try { await diskApi.favoriteFile({ identity: file.identity, is_favorite: file.is_favorite === 1 ? 0 : 1 }); await diskStore.refresh(); } catch (error) { ElMessage.error(getErrorMessage(error, "更新收藏失败")); } }
async function handleBatchFavorite(isFavorite: number) { try { await diskApi.batchFavoriteFiles({ identities: selectedFiles.value.map((item) => item.identity), is_favorite: isFavorite }); selectedFiles.value = []; ElMessage.success(isFavorite === 1 ? "已批量收藏" : "已取消收藏"); await diskStore.refresh(); } catch (error) { ElMessage.error(getErrorMessage(error, "批量更新收藏失败")); } }
async function handleRestore(file: UserFile) { try { await diskApi.restoreRecycleFiles({ identities: [file.identity] }); ElMessage.success("恢复成功"); await refreshAfterMutation(); } catch (error) { ElMessage.error(getErrorMessage(error, "恢复失败")); } }
async function handleBatchRestore() { try { await diskApi.restoreRecycleFiles({ identities: selectedFiles.value.map((item) => item.identity) }); selectedFiles.value = []; ElMessage.success("批量恢复成功"); await refreshAfterMutation(); } catch (error) { ElMessage.error(getErrorMessage(error, "批量恢复失败")); } }
async function handleDestroy(file: UserFile) { try { await ElMessageBox.confirm(`确认彻底删除 ${file.name} 吗？该操作不可恢复。`, "彻底删除", { confirmButtonText: "彻底删除", type: "warning" }); await diskApi.deleteRecycleFiles({ identities: [file.identity] }); ElMessage.success("已彻底删除"); await refreshAfterMutation(); } catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "彻底删除失败")); } }
async function handleBatchDestroy() { try { await ElMessageBox.confirm(`确认彻底删除选中的 ${selectedFiles.value.length} 项吗？该操作不可恢复。`, "彻底删除", { confirmButtonText: "彻底删除", type: "warning" }); await diskApi.deleteRecycleFiles({ identities: selectedFiles.value.map((item) => item.identity) }); selectedFiles.value = []; ElMessage.success("批量彻底删除成功"); await refreshAfterMutation(); } catch (error) { if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "批量彻底删除失败")); } }
function handleResetFilters() { diskStore.query = ""; diskStore.fileType = "all"; diskStore.favoriteOnly = false; diskStore.orderBy = ""; diskStore.orderDir = "desc"; void diskStore.applyFilters(); }
async function handleLogout() { authStore.logout(); await router.replace("/login"); }
</script>

<style scoped>
.bd-page{--cd-card:#fff;--cd-card-strong:#f7faff;--cd-text:#1f2937;--cd-text-soft:#7a8799;--cd-border:#e6edf7;--cd-primary:#1677ff;--cd-primary-strong:#0f5fd6;--cd-accent:#6ea8ff;--cd-danger:#ff4d4f;--cd-shadow:0 14px 34px rgba(22,119,255,.08);--cd-radius-lg:18px;--cd-radius-md:14px;padding:18px;min-height:100vh;background:linear-gradient(180deg,#f5f9ff 0%,#eef4ff 100%)}
.bd-shell{display:grid;gap:18px;grid-template-columns:240px minmax(0,1fr)}.bd-sidebar,.bd-main{display:grid;gap:16px;align-content:start}
.brand{display:flex;gap:14px;align-items:center;padding:18px 16px;border-radius:18px;background:linear-gradient(180deg,#1677ff 0%,#2e89ff 100%);color:#fff}.brand-icon{display:flex;align-items:center;justify-content:center;width:44px;height:44px;border-radius:14px;background:rgba(255,255,255,.16);font-size:20px;font-weight:700}.brand p{margin:4px 0 0;color:rgba(255,255,255,.82);font-size:13px}
.side-btn{width:100%}.side-nav{display:grid;gap:8px}.nav-btn{padding:12px 14px;border:0;border-radius:14px;background:transparent;text-align:left;cursor:pointer;color:var(--cd-text)}.nav-btn.active,.nav-btn:hover{background:rgba(22,119,255,.1);color:var(--cd-primary-strong);font-weight:700}
.side-card{padding:16px}.card-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:12px;font-weight:700}.text-btn{border:0;background:transparent;color:var(--cd-primary-strong);cursor:pointer}
.recent-link{display:grid;gap:4px;width:100%;margin-top:8px;padding:10px 12px;border:1px solid var(--cd-border);border-radius:12px;background:var(--cd-card-strong);text-align:left;cursor:pointer}.recent-link strong,.recent-link span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.recent-link span,.task-meta{color:var(--cd-text-soft);font-size:12px}
.topbar,.main-card{padding:22px 24px}.topbar{display:flex;justify-content:space-between;gap:18px;align-items:center}.kicker{display:inline-flex;padding:5px 10px;border-radius:999px;background:rgba(22,119,255,.1);color:var(--cd-primary-strong);font-size:12px;font-weight:700}.topbar h1{margin:10px 0 6px;font-size:28px}.top-tools{display:flex;gap:12px;align-items:center}.search-box{width:300px}.account-box{min-width:210px;padding:12px 14px;border:1px solid var(--cd-border);border-radius:14px;background:var(--cd-card-strong)}.account-box span{display:block;margin-bottom:6px;color:var(--cd-text-soft);font-size:12px}.account-box strong{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.crumb-btn{padding:0;border:0;background:transparent;color:var(--cd-primary-strong);cursor:pointer}.toolbar,.toolbar-left,.toolbar-right,.filters,.batch-actions,.table-head,.task-strip,.task-chip-bottom{display:flex;gap:10px;flex-wrap:wrap}.toolbar,.table-head{justify-content:space-between;align-items:center}.toolbar{margin-bottom:14px}.filters{align-items:center;margin-bottom:18px;padding:14px 16px;border:1px solid var(--cd-border);border-radius:14px;background:var(--cd-card-strong)}.table-head{margin:10px 0 12px}.table-head span{color:var(--cd-text-soft);font-size:13px}
.task-strip{margin-bottom:18px}.task-chip{flex:1 1 240px;padding:12px;border:1px solid var(--cd-border);border-radius:14px;background:var(--cd-card-strong)}.task-chip-top,.task-chip-bottom{display:flex;justify-content:space-between;gap:8px;align-items:center}.task-chip-top strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:13px}.task-chip-top span,.task-chip-bottom span{color:var(--cd-text-soft);font-size:12px}
.pagination-row{display:flex;justify-content:flex-end;margin-top:18px}.hidden-input{display:none}.dialog-tip{margin:0 0 14px}.root-option{display:flex;margin-bottom:12px}.move-selection{display:grid;gap:8px;margin-top:14px;padding:16px}
@media (max-width:1180px){.bd-shell{grid-template-columns:1fr}}@media (max-width:900px){.topbar,.toolbar,.table-head{flex-direction:column;align-items:stretch}.top-tools{flex-direction:column;align-items:stretch}.search-box,.account-box{width:100%;min-width:0}}
</style>

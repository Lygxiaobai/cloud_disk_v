<template>
  <div class="page-shell bd-page">
    <input ref="fileInputRef" class="hidden-input" type="file" multiple @change="handleFilePicked" />
    <input ref="versionFileInputRef" class="hidden-input" type="file" @change="handleVersionFilePicked" />

    <div v-if="isMobile" class="mobile-topbar panel">
      <button class="icon-btn" type="button" aria-label="打开菜单" @click="sidebarOpen = true">
        <span class="hamburger"></span>
        <span class="hamburger"></span>
        <span class="hamburger"></span>
      </button>
      <div class="mobile-brand">
        <strong>我的云盘</strong>
        <small>{{ authStore.profile?.email || authStore.identity || "未同步邮箱" }}</small>
      </div>
      <el-button size="small" @click="handleLogout">退出</el-button>
    </div>

    <div class="bd-shell">
      <aside v-if="!isMobile" class="bd-sidebar">
        <div class="brand">
          <div class="brand-icon">盘</div>
          <div>
            <strong>我的云盘</strong>
            <p>更适合重度用户的文件中心</p>
          </div>
        </div>

        <el-button class="side-btn" type="primary" @click="triggerUpload">上传文件</el-button>
        <el-button class="side-btn" @click="handleCreateFolder">新建文件夹</el-button>

        <div class="side-nav">
          <button
            v-for="item in navItems"
            :key="item.key"
            class="nav-btn"
            :class="{ active: activeNav === item.key }"
            type="button"
            @click="openSidebarView(item.key)"
          >
            {{ item.label }}
          </button>
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
          <button
            v-for="item in quickRecentItems"
            :key="item.identity"
            class="recent-link"
            type="button"
            @click="handleOpen(item)"
          >
            <strong>{{ item.name }}</strong>
            <span>{{ item.last_accessed_at || item.updated_at || "-" }}</span>
          </button>
        </div>
      </aside>

      <el-drawer
        v-if="isMobile"
        v-model="sidebarOpen"
        direction="ltr"
        size="82%"
        :with-header="false"
        class="bd-drawer"
      >
        <aside class="bd-sidebar bd-sidebar-drawer">
          <div class="brand">
            <div class="brand-icon">盘</div>
            <div>
              <strong>我的云盘</strong>
              <p>{{ authStore.profile?.email || authStore.identity || "未同步邮箱" }}</p>
            </div>
          </div>

          <el-button class="side-btn" type="primary" @click="triggerUpload">上传文件</el-button>
          <el-button class="side-btn" @click="handleCreateFolder">新建文件夹</el-button>

          <div class="side-nav">
            <button
              v-for="item in navItems"
              :key="item.key"
              class="nav-btn"
              :class="{ active: activeNav === item.key }"
              type="button"
              @click="openSidebarView(item.key)"
            >
              {{ item.label }}
            </button>
          </div>

          <div class="panel side-card">
            <div class="card-head">
              <span>目录树</span>
              <button class="text-btn" type="button" @click="handleDrawerRoot">根目录</button>
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
            <button
              v-for="item in quickRecentItems"
              :key="item.identity"
              class="recent-link"
              type="button"
              @click="handleOpen(item)"
            >
              <strong>{{ item.name }}</strong>
              <span>{{ item.last_accessed_at || item.updated_at || "-" }}</span>
            </button>
          </div>
        </aside>
      </el-drawer>

      <main class="bd-main">
        <header class="panel topbar">
          <div>
            <span class="kicker">个人云存储中心</span>
            <h1>{{ pageHeading }}</h1>
            <el-breadcrumb v-if="showBreadcrumb" separator="/">
              <el-breadcrumb-item
                v-for="(crumb, index) in diskStore.breadcrumbs"
                :key="`${crumb.id}-${index}`"
              >
                <button class="crumb-btn" type="button" @click="diskStore.jumpToBreadcrumb(index)">
                  {{ crumb.name }}
                </button>
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <div class="top-tools">
            <el-input
              v-model="diskStore.query"
              class="search-box"
              clearable
              :placeholder="searchPlaceholder"
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
              <template v-if="diskStore.viewMode === 'files'">
                <el-button type="primary" @click="triggerUpload">上传</el-button>
                <el-button @click="handleCreateFolder">新建文件夹</el-button>
              </template>
              <el-button @click="diskStore.refresh">刷新</el-button>
            </div>
            <div v-if="showSelectionActions" class="toolbar-right">
              <template v-if="diskStore.viewMode === 'recycle'">
                <el-button type="primary" plain @click="handleBatchRestore">批量恢复</el-button>
                <el-button type="danger" plain @click="handleBatchDestroy">彻底删除</el-button>
              </template>
              <template v-else>
                <el-button type="warning" plain @click="handleBatchFavorite(1)">批量收藏</el-button>
                <el-button plain @click="handleBatchFavorite(0)">取消收藏</el-button>
                <el-button type="primary" plain @click="openBatchRenameDialog">批量重命名</el-button>
                <el-button type="primary" plain @click="openMoveDialog()">批量移动</el-button>
                <el-button type="danger" plain @click="handleBatchDelete">批量删除</el-button>
              </template>
            </div>
          </div>

          <div class="filters">
            <el-select v-if="diskStore.viewMode === 'files'" v-model="diskStore.searchScope" style="width: 140px">
              <el-option label="当前目录" value="folder" />
              <el-option label="全盘搜索" value="all" />
            </el-select>
            <el-select v-if="diskStore.viewMode !== 'recycle' && diskStore.viewMode !== 'shares'" v-model="diskStore.fileType" style="width: 132px">
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
            <el-input-number v-if="diskStore.viewMode === 'large'" v-model="diskStore.minSizeMB" :min="1" :step="50" controls-position="right" />
            <span v-if="diskStore.viewMode === 'large'" class="filter-tip">MB 以上视为大文件</span>
            <el-select v-model="diskStore.orderBy" style="width: 140px">
              <el-option label="默认排序" value="" />
              <el-option label="修改时间" value="updated_at" />
              <el-option label="创建时间" value="created_at" />
              <el-option v-if="diskStore.viewMode !== 'shares'" label="文件大小" value="size" />
              <el-option label="名称" value="name" />
              <el-option v-if="diskStore.viewMode === 'recycle'" label="删除时间" value="deleted_at" />
            </el-select>
            <el-select v-model="diskStore.orderDir" style="width: 112px">
              <el-option label="降序" value="desc" />
              <el-option label="升序" value="asc" />
            </el-select>
            <el-switch v-if="diskStore.viewMode !== 'recycle' && diskStore.viewMode !== 'shares'" v-model="diskStore.favoriteOnly" active-text="仅收藏" />
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
                <span v-if="task.status === 'uploading' && task.etaSeconds !== null">
                  剩余 {{ formatDuration(task.etaSeconds) }}
                </span>
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

          <ShareTable
            v-if="diskStore.viewMode === 'shares'"
            :items="diskStore.shareItems"
            :loading="diskStore.loading"
            @copy="handleCopyShare"
            @delete="handleDeleteShare"
            @open="handleOpenSharePage"
          />
          <FileTable
            v-else
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
            @versions="openVersionDialog"
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

    <el-dialog v-model="moveDialogVisible" title="移动到文件夹" width="min(520px, 92vw)" top="5vh">
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

    <el-dialog v-model="shareDialogVisible" title="创建分享" width="min(560px, 92vw)" top="5vh">
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
          <el-form-item label="访问口令">
            <el-input v-model="shareAccessCode" clearable maxlength="8" placeholder="可选，建议 4-8 位" />
          </el-form-item>
          <el-form-item label="分享模式">
            <el-switch v-model="shareAllowDownload" inline-prompt active-text="可保存" inactive-text="仅预览" />
          </el-form-item>
          <el-form-item v-if="shareLink" label="分享链接">
            <el-input :model-value="shareLink" readonly />
          </el-form-item>
          <el-form-item v-if="shareSummary" label="复制信息">
            <el-input :model-value="shareSummary" readonly type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="shareDialogVisible = false">关闭</el-button>
        <el-button v-if="shareSummary" @click="copyText(shareSummary)">复制</el-button>
        <el-button type="primary" @click="handleCreateShare">{{ shareLink ? "重新生成" : "生成分享链接" }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchRenameDialogVisible" title="批量重命名" width="min(560px, 92vw)" top="5vh">
      <div class="share-panel">
        <p class="muted">对当前选中的 {{ selectedFiles.length }} 个文件应用同一套命名规则。</p>
        <el-form label-position="top">
          <el-form-item label="前缀">
            <el-input v-model="batchRenameForm.prefix" clearable placeholder="例如：项目-" />
          </el-form-item>
          <el-form-item label="后缀">
            <el-input v-model="batchRenameForm.suffix" clearable placeholder="例如：-终稿" />
          </el-form-item>
          <el-form-item label="文本替换">
            <div class="rename-grid">
              <el-input v-model="batchRenameForm.findText" clearable placeholder="查找文本" />
              <el-input v-model="batchRenameForm.replaceText" clearable placeholder="替换为" />
            </div>
          </el-form-item>
          <el-form-item label="顺序编号">
            <el-switch v-model="batchRenameForm.applySequence" inline-prompt active-text="开启" inactive-text="关闭" />
            <div v-if="batchRenameForm.applySequence" class="rename-grid rename-grid-narrow">
              <el-input-number v-model="batchRenameForm.startIndex" :min="1" controls-position="right" />
              <el-input-number v-model="batchRenameForm.step" :min="1" controls-position="right" />
              <el-input-number v-model="batchRenameForm.padding" :min="0" :max="8" controls-position="right" />
            </div>
          </el-form-item>
          <el-form-item label="保留扩展名">
            <el-switch v-model="batchRenameForm.keepExt" inline-prompt active-text="保留" inactive-text="不保留" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="batchRenameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchRename">应用规则</el-button>
      </template>
    </el-dialog>

    <FileVersionDialog
      v-model="versionDialogVisible"
      :file="versionTarget"
      :loading="versionLoading"
      :restoring-identity="versionRestoringIdentity"
      :versions="versionItems"
      @restore-version="handleRestoreVersion"
      @upload-version="triggerVersionUpload"
    />
    <FilePreviewDialog v-model="previewDialogVisible" :preview="previewData" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { TreeOptionProps } from "element-plus";
import { useRouter } from "vue-router";

import FilePreviewDialog from "@/components/FilePreviewDialog.vue";
import FileTable from "@/components/FileTable.vue";
import FileVersionDialog from "@/components/FileVersionDialog.vue";
import ShareTable from "@/components/ShareTable.vue";
import * as diskApi from "@/api/disk";
import * as shareApi from "@/api/share";
import type { FilePreviewResponse, FileVersionItem, FolderTreeNode, ShareListItem, UserFile } from "@/types/api";
import { useAuthStore } from "@/stores/auth";
import { useDiskStore } from "@/stores/disk";
import { useUploadStore } from "@/stores/upload";
import { getErrorMessage } from "@/utils/error";

type SidebarTarget = "files" | "favorite" | "recent" | "duplicates" | "large" | "shares" | "recycle";

interface BatchRenameFormState {
  prefix: string;
  suffix: string;
  findText: string;
  replaceText: string;
  applySequence: boolean;
  startIndex: number;
  step: number;
  padding: number;
  keepExt: boolean;
}

// 侧边栏入口统一收口在这里，便于后面继续扩展“版本历史”“同步中心”等模块。
const navItems: Array<{ key: SidebarTarget; label: string }> = [
  { key: "files", label: "全部文件" },
  { key: "favorite", label: "我的收藏" },
  { key: "recent", label: "最近文件" },
  { key: "duplicates", label: "重复文件" },
  { key: "large", label: "大文件" },
  { key: "shares", label: "我的分享" },
  { key: "recycle", label: "回收站" },
];

const router = useRouter();
const authStore = useAuthStore();
const diskStore = useDiskStore();
const uploadStore = useUploadStore();

const fileInputRef = ref<HTMLInputElement | null>(null);
const versionFileInputRef = ref<HTMLInputElement | null>(null);
const selectedFiles = ref<UserFile[]>([]);
const moveDialogVisible = ref(false);
const moveCandidate = ref<UserFile | null>(null);
const moveTargetIdentity = ref("");
const moveTargetName = ref("根目录");
const shareDialogVisible = ref(false);
const shareTarget = ref<UserFile | null>(null);
const shareExpiredTime = ref(86400);
const shareAccessCode = ref("");
const shareAllowDownload = ref(true);
const shareLink = ref("");
const shareSummary = ref("");
const folderTreeSupported = ref(true);
const folderTreeMessage = ref("目录树接口加载失败，请先检查 /user/folder/children 和 /user/folder/path。");
const previewDialogVisible = ref(false);
const previewData = ref<FilePreviewResponse | null>(null);
const batchRenameDialogVisible = ref(false);
const versionDialogVisible = ref(false);
const versionTarget = ref<UserFile | null>(null);
const versionItems = ref<FileVersionItem[]>([]);
const versionLoading = ref(false);
const versionRestoringIdentity = ref("");

const batchRenameForm = ref<BatchRenameFormState>({
  prefix: "",
  suffix: "",
  findText: "",
  replaceText: "",
  applySequence: false,
  startIndex: 1,
  step: 1,
  padding: 0,
  keepExt: true,
});

const treeProps: TreeOptionProps = {
  label: "name",
  isLeaf: (data) => (data as FolderTreeNode).has_children !== 1,
};

const quickRecentItems = computed(() => diskStore.recentItems.slice(0, 4));
const activeNav = computed<SidebarTarget>(() => {
  if (diskStore.viewMode === "recent") return "recent";
  if (diskStore.viewMode === "recycle") return "recycle";
  if (diskStore.viewMode === "duplicates") return "duplicates";
  if (diskStore.viewMode === "large") return "large";
  if (diskStore.viewMode === "shares") return "shares";
  return diskStore.favoriteOnly ? "favorite" : "files";
});
const pageHeading = computed(() => {
  if (activeNav.value === "favorite") return "我的收藏";
  if (activeNav.value === "recent") return "最近文件";
  if (activeNav.value === "duplicates") return "重复文件治理";
  if (activeNav.value === "large") return "大文件治理";
  if (activeNav.value === "shares") return "我的分享";
  if (activeNav.value === "recycle") return "回收站";
  return diskStore.currentFolderName;
});
const searchPlaceholder = computed(() => {
  if (diskStore.viewMode === "shares") return "搜索分享记录";
  if (diskStore.viewMode === "duplicates") return "搜索重复文件";
  if (diskStore.viewMode === "large") return "搜索大文件";
  return diskStore.searchScope === "all" ? "搜索全盘文件" : "搜索当前目录文件";
});
const showBreadcrumb = computed(() => diskStore.viewMode === "files" && diskStore.searchScope === "folder");
const showSelectionActions = computed(() => diskStore.viewMode !== "shares" && selectedFiles.value.length > 0);
const currentUploadContext = computed(() => ({
  parentId: diskStore.currentFolderId,
  parentIdentity: diskStore.currentFolderIdentity || undefined,
}));

const isMobile = ref(false);
const sidebarOpen = ref(false);
const mql = typeof window !== "undefined" ? window.matchMedia("(max-width: 600px)") : null;

function handleMqlChange(event: MediaQueryListEvent): void {
  isMobile.value = event.matches;
  if (!event.matches) {
    sidebarOpen.value = false;
  }
}

function closeSidebarIfMobile(): void {
  if (isMobile.value) {
    sidebarOpen.value = false;
  }
}

async function handleDrawerRoot(): Promise<void> {
  await diskStore.loadRoot();
  closeSidebarIfMobile();
}

onMounted(async () => {
  isMobile.value = mql?.matches ?? false;
  mql?.addEventListener("change", handleMqlChange);
  await authStore.bootstrap();
  await diskStore.loadRoot();
});

onBeforeUnmount(() => {
  mql?.removeEventListener("change", handleMqlChange);
});

function formatFileSize(size: number): string {
  if (!size) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.ceil(seconds)} 秒`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes} 分 ${Math.ceil(seconds % 60)} 秒`;
}

function taskStatusText(status: string): string {
  return ({
    hashing: "计算中",
    waiting: "等待上传",
    uploading: "上传中",
    paused: "已暂停",
    completing: "写入网盘",
    success: "已完成",
    error: "失败",
  } as Record<string, string>)[status] || status;
}

async function openSidebarView(target: SidebarTarget): Promise<void> {
  selectedFiles.value = [];

  // 收藏仍然复用文件列表接口，只是切成 favoriteOnly 筛选状态。
  if (target === "favorite") {
    if (diskStore.viewMode !== "files") await diskStore.setViewMode("files");
    diskStore.favoriteOnly = true;
    diskStore.searchScope = "folder";
    await diskStore.applyFilters();
    closeSidebarIfMobile();
    return;
  }

  if (target === "files") {
    if (diskStore.viewMode !== "files") await diskStore.setViewMode("files");
    diskStore.favoriteOnly = false;
    diskStore.searchScope = "folder";
    await diskStore.applyFilters();
    closeSidebarIfMobile();
    return;
  }

  // 其他入口都是真正的独立视图模式，例如重复文件 / 大文件 / 我的分享。
  diskStore.favoriteOnly = false;
  await diskStore.setViewMode(target);
  closeSidebarIfMobile();
}

function triggerUpload(): void {
  if (diskStore.viewMode !== "files") {
    ElMessage.info("请先切回文件视图后再上传");
    return;
  }
  fileInputRef.value?.click();
  closeSidebarIfMobile();
}

function triggerVersionUpload(): void {
  if (versionTarget.value) {
    versionFileInputRef.value?.click();
  }
}

async function loadTreeNode(node: unknown, resolve: (data: FolderTreeNode[]) => void): Promise<void> {
  const rawNode = node as { level: number; data?: { id: number } };
  const folderId = rawNode.level === 0 ? 0 : Number(rawNode.data?.id || 0);
  try {
    const data = await diskApi.listFolderChildren(folderId);
    folderTreeSupported.value = true;
    resolve(data.list);
  } catch (error) {
    if (rawNode.level === 0) {
      folderTreeSupported.value = false;
      folderTreeMessage.value = getErrorMessage(error, "目录树接口未完成。");
    } else {
      ElMessage.error(getErrorMessage(error, "加载子目录失败"));
    }
    resolve([]);
  }
}

async function handleTreeNodeClick(folder: FolderTreeNode): Promise<void> {
  // 从左侧目录树进入文件夹时，总是回到标准目录浏览体验。
  if (diskStore.viewMode !== "files") {
    diskStore.favoriteOnly = false;
  }
  await diskStore.openFolder(folder);
  closeSidebarIfMobile();
}

// 所有增删改动作都复用这一个刷新入口，保证列表和最近访问保持同步。
async function refreshAfterMutation(): Promise<void> {
  await diskStore.refresh();
  await diskStore.loadRecent();
}

async function handleFilePicked(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement;
  const files = Array.from(target.files ?? []);
  if (files.length === 0) return;
  try {
    await uploadStore.addFiles(files, currentUploadContext.value, refreshAfterMutation);
    ElMessage.success(`已加入 ${files.length} 个上传任务`);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "创建上传任务失败"));
  } finally {
    target.value = "";
  }
}

async function handleVersionFilePicked(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file || !versionTarget.value) return;

  try {
    await uploadStore.addFiles(
      [file],
      { parentId: versionTarget.value.id, targetFileIdentity: versionTarget.value.identity },
      async () => {
        await refreshAfterMutation();
        await loadVersionHistory(versionTarget.value!.identity);
      },
    );
    ElMessage.success("已加入新版本上传任务");
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "创建版本上传任务失败"));
  } finally {
    target.value = "";
  }
}

async function handleCreateFolder(): Promise<void> {
  if (diskStore.viewMode !== "files") {
    ElMessage.info("请在文件视图中创建文件夹");
    return;
  }
  closeSidebarIfMobile();
  try {
    const { value } = await ElMessageBox.prompt("输入新的文件夹名称", "新建文件夹", {
      confirmButtonText: "创建",
      inputErrorMessage: "请输入文件夹名称",
      inputPattern: /\S+/,
    });
    await diskApi.createFolder({ name: value, parentId: diskStore.currentFolderId });
    ElMessage.success("文件夹创建成功");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "创建文件夹失败"));
  }
}

async function handleRename(file: UserFile): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt("输入新的名称", "重命名", {
      confirmButtonText: "保存",
      inputPattern: /\S+/,
      inputValue: file.name,
    });
    await diskApi.renameFile({ identity: file.identity, name: value });
    ElMessage.success("重命名成功");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "重命名失败"));
  }
}

function openBatchRenameDialog(): void {
  batchRenameForm.value = {
    prefix: "",
    suffix: "",
    findText: "",
    replaceText: "",
    applySequence: false,
    startIndex: 1,
    step: 1,
    padding: 0,
    keepExt: true,
  };
  batchRenameDialogVisible.value = true;
}

async function handleBatchRename(): Promise<void> {
  try {
    const result = await diskApi.batchRenameFiles({
      identities: selectedFiles.value.map((item) => item.identity),
      prefix: batchRenameForm.value.prefix || undefined,
      suffix: batchRenameForm.value.suffix || undefined,
      find_text: batchRenameForm.value.findText || undefined,
      replace_text: batchRenameForm.value.findText ? batchRenameForm.value.replaceText : undefined,
      apply_sequence: batchRenameForm.value.applySequence,
      start_index: batchRenameForm.value.startIndex,
      step: batchRenameForm.value.step,
      padding: batchRenameForm.value.padding,
      keep_ext: batchRenameForm.value.keepExt,
    });
    batchRenameDialogVisible.value = false;
    selectedFiles.value = [];
    ElMessage.success(`已完成 ${result.list.length} 个文件的重命名`);
    await refreshAfterMutation();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "批量重命名失败"));
  }
}

async function handleDelete(file: UserFile): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除 ${file.name} 吗？`, "删除确认", { confirmButtonText: "删除", type: "warning" });
    await diskApi.deleteFile({ identity: file.identity });
    ElMessage.success("已移入回收站");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "删除失败"));
  }
}

async function handleBatchDelete(): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selectedFiles.value.length} 项吗？`, "批量删除", { confirmButtonText: "删除", type: "warning" });
    await diskApi.batchDeleteFiles({ identities: selectedFiles.value.map((item) => item.identity) });
    selectedFiles.value = [];
    ElMessage.success("已移入回收站");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "批量删除失败"));
  }
}

async function handleOpen(file: UserFile): Promise<void> {
  if (file.is_dir === 1) {
    diskStore.favoriteOnly = false;
    diskStore.searchScope = "folder";
    await diskStore.openFolder(file);
    closeSidebarIfMobile();
    return;
  }
  try {
    previewData.value = await diskApi.previewFile(file.identity);
    previewDialogVisible.value = true;
    closeSidebarIfMobile();
    void diskStore.loadRecent();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "加载预览失败"));
  }
}

function openMoveDialog(file?: UserFile): void {
  // 回收站和分享管理不是目录上下文，不允许在这两个视图里发起移动。
  if (diskStore.viewMode === "recycle" || diskStore.viewMode === "shares") return;
  moveCandidate.value = file ?? null;
  moveTargetIdentity.value = "";
  moveTargetName.value = "根目录";
  moveDialogVisible.value = true;
}

function handleMoveTargetSelect(folder: FolderTreeNode): void {
  if (moveCandidate.value?.identity === folder.identity) {
    ElMessage.warning("不能移动到自己里面");
    return;
  }
  moveTargetIdentity.value = folder.identity;
  moveTargetName.value = folder.name;
}

function selectMoveRoot(): void {
  moveTargetIdentity.value = "";
  moveTargetName.value = "根目录";
}

async function handleMove(): Promise<void> {
  try {
    if (moveCandidate.value) {
      await diskApi.moveFile({ identity: moveCandidate.value.identity, parent_identity: moveTargetIdentity.value });
    } else {
      await diskApi.batchMoveFiles({ identities: selectedFiles.value.map((item) => item.identity), parent_identity: moveTargetIdentity.value });
      selectedFiles.value = [];
    }
    ElMessage.success("移动成功");
    moveDialogVisible.value = false;
    await refreshAfterMutation();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "移动失败"));
  }
}

async function openVersionDialog(file: UserFile): Promise<void> {
  versionTarget.value = file;
  versionDialogVisible.value = true;
  await loadVersionHistory(file.identity);
}

async function loadVersionHistory(identity: string): Promise<void> {
  versionLoading.value = true;
  try {
    const data = await diskApi.listFileVersions(identity);
    versionItems.value = data.list;
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "加载版本历史失败"));
  } finally {
    versionLoading.value = false;
  }
}

async function handleRestoreVersion(version: FileVersionItem): Promise<void> {
  if (!versionTarget.value || version.is_current === 1) return;

  try {
    await ElMessageBox.confirm(`确认将 ${version.created_at} 的版本恢复为当前版本吗？`, "恢复历史版本", {
      confirmButtonText: "恢复版本",
      type: "warning",
    });
    versionRestoringIdentity.value = version.identity;
    await diskApi.restoreFileVersion({
      file_identity: versionTarget.value.identity,
      version_identity: version.identity,
    });
    await refreshAfterMutation();
    versionTarget.value =
      diskStore.visibleItems.find((item) => item.identity === versionTarget.value?.identity) || versionTarget.value;
    await loadVersionHistory(versionTarget.value.identity);
    ElMessage.success("版本恢复成功");
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(getErrorMessage(error, "恢复历史版本失败"));
    }
  } finally {
    versionRestoringIdentity.value = "";
  }
}

function openShareDialog(file: UserFile): void {
  // 每次打开分享弹窗都重置配置，避免沿用上一次的口令和权限状态。
  shareTarget.value = file;
  shareExpiredTime.value = 86400;
  shareAccessCode.value = "";
  shareAllowDownload.value = true;
  shareLink.value = "";
  shareSummary.value = "";
  shareDialogVisible.value = true;
}

async function handleCreateShare(): Promise<void> {
  if (!shareTarget.value) return;
  try {
    // 口令在前端统一转成大写，便于用户复制和肉眼核对。
    const accessCode = shareAccessCode.value.trim().toUpperCase();
    const data = await diskApi.createShare({
      expired_time: shareExpiredTime.value,
      user_repository_identity: shareTarget.value.identity,
      access_code: accessCode || undefined,
      allow_download: shareAllowDownload.value ? 1 : 0,
    });
    shareLink.value = `${window.location.origin}/share/${data.identity}`;
    shareSummary.value = accessCode ? `链接：${shareLink.value}\n提取码：${accessCode}` : shareLink.value;
    ElMessage.success("分享链接已生成");
    if (diskStore.viewMode === "shares") await diskStore.refresh();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "创建分享失败"));
  }
}

async function handleToggleFavorite(file: UserFile): Promise<void> {
  try {
    await diskApi.favoriteFile({ identity: file.identity, is_favorite: file.is_favorite === 1 ? 0 : 1 });
    await diskStore.refresh();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "更新收藏失败"));
  }
}

async function handleBatchFavorite(isFavorite: number): Promise<void> {
  try {
    await diskApi.batchFavoriteFiles({ identities: selectedFiles.value.map((item) => item.identity), is_favorite: isFavorite });
    selectedFiles.value = [];
    ElMessage.success(isFavorite === 1 ? "已批量收藏" : "已取消收藏");
    await diskStore.refresh();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "批量更新收藏失败"));
  }
}

async function handleRestore(file: UserFile): Promise<void> {
  try {
    await diskApi.restoreRecycleFiles({ identities: [file.identity] });
    ElMessage.success("恢复成功");
    await refreshAfterMutation();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "恢复失败"));
  }
}

async function handleBatchRestore(): Promise<void> {
  try {
    await diskApi.restoreRecycleFiles({ identities: selectedFiles.value.map((item) => item.identity) });
    selectedFiles.value = [];
    ElMessage.success("批量恢复成功");
    await refreshAfterMutation();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "批量恢复失败"));
  }
}

async function handleDestroy(file: UserFile): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认彻底删除 ${file.name} 吗？该操作不可恢复。`, "彻底删除", { confirmButtonText: "彻底删除", type: "warning" });
    await diskApi.deleteRecycleFiles({ identities: [file.identity] });
    ElMessage.success("已彻底删除");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "彻底删除失败"));
  }
}

async function handleBatchDestroy(): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认彻底删除选中的 ${selectedFiles.value.length} 项吗？该操作不可恢复。`, "彻底删除", { confirmButtonText: "彻底删除", type: "warning" });
    await diskApi.deleteRecycleFiles({ identities: selectedFiles.value.map((item) => item.identity) });
    selectedFiles.value = [];
    ElMessage.success("批量彻底删除成功");
    await refreshAfterMutation();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "批量彻底删除失败"));
  }
}

function handleResetFilters(): void {
  // 重置时把全盘搜索和大文件阈值也一并恢复，保证用户回到最基础的浏览状态。
  diskStore.query = "";
  diskStore.fileType = "all";
  diskStore.favoriteOnly = false;
  diskStore.orderBy = "";
  diskStore.orderDir = "desc";
  diskStore.searchScope = "folder";
  diskStore.minSizeMB = 100;
  void diskStore.applyFilters();
}

// 分享复制统一走这个入口，便于后面扩展为“复制带口令的完整文案”。
async function copyText(text: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success("已复制到剪贴板");
  } catch {
    ElMessage.warning("复制失败，请手动复制");
  }
}

async function handleCopyShare(item: ShareListItem): Promise<void> {
  await copyText(`${window.location.origin}/share/${item.identity}`);
}

function handleOpenSharePage(item: ShareListItem): void {
  window.open(`/share/${item.identity}`, "_blank", "noopener,noreferrer");
}

async function handleDeleteShare(item: ShareListItem): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认取消 ${item.name} 的分享吗？`, "取消分享", { confirmButtonText: "取消分享", type: "warning" });
    await shareApi.deleteShares({ identities: [item.identity] });
    ElMessage.success("分享已取消");
    await diskStore.refresh();
  } catch (error) {
    if (error !== "cancel" && error !== "close") ElMessage.error(getErrorMessage(error, "取消分享失败"));
  }
}

async function handleLogout(): Promise<void> {
  authStore.logout();
  await router.replace("/login");
}
</script>

<style scoped>
.bd-page {
  --cd-card: #fff;
  --cd-card-strong: #f7faff;
  --cd-text: #1f2937;
  --cd-text-soft: #7a8799;
  --cd-border: #e6edf7;
  --cd-primary-strong: #0f5fd6;
  padding: 18px;
  min-height: 100vh;
  background: linear-gradient(180deg, #f5f9ff 0%, #eef4ff 100%);
}

.bd-shell {
  display: grid;
  gap: 18px;
  grid-template-columns: 240px minmax(0, 1fr);
}

.bd-sidebar,
.bd-main {
  display: grid;
  gap: 16px;
  align-content: start;
}

.brand {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 18px 16px;
  border-radius: 18px;
  background: linear-gradient(180deg, #1677ff 0%, #2e89ff 100%);
  color: #fff;
}

.brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.16);
  font-size: 20px;
  font-weight: 700;
}

.brand p {
  margin: 4px 0 0;
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
}

.side-btn,
.search-box {
  width: 100%;
}

.side-nav {
  display: grid;
  gap: 8px;
}

.nav-btn,
.text-btn,
.crumb-btn {
  border: 0;
  background: transparent;
  cursor: pointer;
}

.nav-btn {
  padding: 12px 14px;
  border-radius: 14px;
  text-align: left;
  color: var(--cd-text);
}

.nav-btn.active,
.nav-btn:hover {
  background: rgba(22, 119, 255, 0.1);
  color: var(--cd-primary-strong);
  font-weight: 700;
}

.side-card {
  padding: 16px;
}

.card-head,
.toolbar,
.toolbar-left,
.toolbar-right,
.filters,
.table-head,
.task-strip,
.task-chip-bottom,
.top-tools {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.card-head,
.toolbar,
.table-head {
  justify-content: space-between;
  align-items: center;
}

.recent-link {
  display: grid;
  gap: 4px;
  width: 100%;
  margin-top: 8px;
  padding: 10px 12px;
  border: 1px solid var(--cd-border);
  border-radius: 12px;
  background: var(--cd-card-strong);
  text-align: left;
  cursor: pointer;
}

.recent-link strong,
.recent-link span,
.account-box strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recent-link span,
.filter-tip,
.muted,
.table-head span,
.task-chip-top span,
.task-chip-bottom span,
.account-box span {
  color: var(--cd-text-soft);
  font-size: 12px;
}

.topbar,
.main-card {
  padding: 22px 24px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: center;
}

.kicker {
  display: inline-flex;
  padding: 5px 10px;
  border-radius: 999px;
  background: rgba(22, 119, 255, 0.1);
  color: var(--cd-primary-strong);
  font-size: 12px;
  font-weight: 700;
}

.topbar h1 {
  margin: 10px 0 6px;
  font-size: 28px;
}

.account-box {
  min-width: 210px;
  padding: 12px 14px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: var(--cd-card-strong);
}

.account-box span {
  display: block;
  margin-bottom: 6px;
}

.filters {
  align-items: center;
  margin-bottom: 18px;
  padding: 14px 16px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: var(--cd-card-strong);
}

.task-chip {
  flex: 1 1 240px;
  padding: 12px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: var(--cd-card-strong);
}

.task-chip-top,
.task-chip-bottom {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
}

.task-chip-top strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}

.hidden-input {
  display: none;
}

.dialog-tip {
  margin: 0 0 14px;
}

.root-option {
  display: flex;
  margin-bottom: 12px;
}

.move-selection {
  display: grid;
  gap: 8px;
  margin-top: 14px;
  padding: 16px;
}

.soft-code {
  color: var(--cd-primary-strong);
  font-family: Consolas, Monaco, monospace;
}

.share-panel {
  display: grid;
  gap: 10px;
}

.rename-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  width: 100%;
}

.rename-grid-narrow {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-top: 10px;
}

@media (max-width: 1180px) {
  .bd-shell {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .topbar,
  .toolbar,
  .table-head,
  .top-tools {
    flex-direction: column;
    align-items: stretch;
  }

  .account-box {
    min-width: 0;
  }

  .rename-grid,
  .rename-grid-narrow {
    grid-template-columns: 1fr;
  }
}

.mobile-topbar {
  display: none;
}

@media (max-width: 600px) {
  .bd-page {
    padding: 10px;
  }

  .bd-shell {
    gap: 12px;
  }

  .mobile-topbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    margin-bottom: 12px;
  }

  .mobile-brand {
    flex: 1 1 0;
    min-width: 0;
    display: grid;
    gap: 2px;
  }

  .mobile-brand strong {
    font-size: 15px;
  }

  .mobile-brand small {
    color: var(--cd-text-soft);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .icon-btn {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    width: 38px;
    height: 38px;
    padding: 0;
    border: 1px solid var(--cd-border);
    border-radius: 10px;
    background: var(--cd-card-strong);
    cursor: pointer;
  }

  .hamburger {
    display: block;
    width: 16px;
    height: 2px;
    background: var(--cd-text);
    border-radius: 999px;
  }

  .topbar {
    padding: 14px;
  }

  .topbar h1 {
    font-size: 20px;
    margin: 6px 0 4px;
  }

  .main-card {
    padding: 14px;
  }

  .account-box {
    display: none;
  }

  .filters {
    padding: 10px;
    gap: 8px;
    flex-direction: column;
    align-items: stretch;
  }

  .filters .el-select {
    width: 100% !important;
  }

  .filters :deep(.el-select__wrapper) {
    width: 100%;
  }

  .filters .el-button,
  .filters .el-switch,
  .filters .el-input-number {
    width: 100%;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .toolbar-left .el-button,
  .toolbar-right .el-button {
    flex: 1 1 calc(50% - 5px);
  }

  .task-chip {
    flex: 1 1 100%;
  }

  .rename-grid,
  .rename-grid-narrow {
    grid-template-columns: 1fr !important;
  }

  .pagination-row {
    justify-content: center;
    margin-top: 14px;
  }

  .pagination-row :deep(.el-pagination__total),
  .pagination-row :deep(.el-pagination__jump) {
    display: none;
  }

  .pagination-row :deep(.el-pager li) {
    min-width: 28px;
  }

  .share-panel :deep(.el-form-item__label),
  .dialog-tip {
    font-size: 13px;
  }
}

.bd-sidebar-drawer {
  display: grid;
  gap: 16px;
  align-content: start;
}
</style>

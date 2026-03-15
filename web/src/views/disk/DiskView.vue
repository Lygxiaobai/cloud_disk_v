<template>
  <div class="page-shell disk-page">
    <header class="panel hero-panel">
      <div class="hero-copy">
        <span class="pill">Cloud Disk Workspace</span>
        <h1 class="page-title">你好，{{ authStore.displayName }}</h1>
        <p class="page-subtitle">
          先把“登录、刷新、目录、上传、分享”这些主链路走顺。复杂体验可以后面继续加，但基础工作台先稳住。
        </p>
      </div>

      <div class="hero-actions">
        <div class="profile-card">
          <span class="muted">当前账号</span>
          <strong>{{ authStore.profile?.email || authStore.identity || "未同步邮箱" }}</strong>
        </div>
        <el-button round @click="handleLogout">退出登录</el-button>
      </div>
    </header>

    <section class="workspace-grid">
      <aside class="panel rail-panel">
        <div class="rail-section tree-section">
          <span class="pill">目录树</span>
          <p class="muted tree-note">左侧只展示文件夹，用于快速看清目录层级，也方便后面扩展文件移动。</p>

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

          <el-alert
            v-else
            :closable="false"
            :title="folderTreeMessage"
            show-icon
            type="info"
          />
        </div>

        <div class="rail-section">
          <span class="pill">Storage Snapshot</span>
          <div class="stat-grid stat-grid-rail">
            <article class="stat-card">
              <span class="muted">当前目录文件夹</span>
              <strong>{{ diskStore.folderItems.length }}</strong>
            </article>
            <article class="stat-card">
              <span class="muted">当前目录文件</span>
              <strong>{{ diskStore.fileItems.length }}</strong>
            </article>
            <article class="stat-card">
              <span class="muted">列表总数</span>
              <strong>{{ diskStore.total }}</strong>
            </article>
          </div>
        </div>

        <div class="rail-section tip-box">
          <strong>当前约束</strong>
          <p class="muted">
            目录树依赖后端新增 `/user/folder/children` 和 `/user/folder/path`。后端未完成前，页面会保留现有文件列表与面包屑能力。
          </p>
        </div>
      </aside>

      <main class="panel main-panel">
        <div class="workspace-topbar">
          <div>
            <h2>{{ diskStore.currentFolderName }}</h2>
            <el-breadcrumb separator="/">
              <el-breadcrumb-item
                v-for="(crumb, index) in diskStore.breadcrumbs"
                :key="`${crumb.id}-${index}`"
              >
                <button class="crumb-button" type="button" @click="diskStore.jumpToBreadcrumb(index)">
                  {{ crumb.name }}
                </button>
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>

          <div class="toolbar">
            <input ref="fileInputRef" class="hidden-input" type="file" @change="handleFilePicked" />
            <el-button type="primary" @click="triggerUpload">上传文件</el-button>
            <el-button @click="handleCreateFolder">新建文件夹</el-button>
            <el-button @click="diskStore.refresh">刷新列表</el-button>
          </div>
        </div>

        <FileTable
          :files="diskStore.items"
          :loading="diskStore.loading"
          @delete="handleDelete"
          @move="openMoveDialog"
          @open="handleOpen"
          @rename="handleRename"
          @share="openShareDialog"
        />

        <div class="pagination-row">
          <el-pagination
            :current-page="diskStore.page"
            :page-size="diskStore.size"
            :total="diskStore.total"
            background
            layout="prev, pager, next, total"
            @current-change="diskStore.changePage"
          />
        </div>
      </main>
    </section>

    <el-dialog v-model="moveDialogVisible" title="移动到文件夹" width="520px">
      <template v-if="folderTreeSupported">
        <p class="muted dialog-tip">选择一个目标文件夹。这里使用的也是左侧目录树接口。</p>
        <el-tree
          node-key="identity"
          lazy
          :load="loadTreeNode"
          :props="treeProps"
          highlight-current
          @node-click="handleMoveTargetSelect"
        />
        <div v-if="moveTargetIdentity" class="move-selection panel">
          <span class="muted">当前目标</span>
          <strong>{{ moveTargetName }}</strong>
          <span class="soft-code">{{ moveTargetIdentity }}</span>
        </div>
      </template>

      <template v-else-if="moveTargets.length > 0">
        <el-radio-group v-model="moveTargetIdentity" class="folder-choice-group">
          <el-radio
            v-for="folder in moveTargets"
            :key="folder.identity"
            :label="folder.identity"
            size="large"
          >
            <div class="folder-choice">
              <strong>{{ folder.name }}</strong>
              <span class="soft-code">{{ folder.identity }}</span>
            </div>
          </el-radio>
        </el-radio-group>
      </template>

      <el-empty v-else description="当前页没有可以移动到的目标文件夹" />

      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button :disabled="!moveTargetIdentity" type="primary" @click="handleMove">
          确认移动
        </el-button>
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
        <el-button type="primary" @click="handleCreateShare">
          {{ shareLink ? "重新生成" : "生成分享链接" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useRouter } from "vue-router";

import FileTable from "@/components/FileTable.vue";
import * as diskApi from "@/api/disk";
import type { FolderTreeNode, UserFile } from "@/types/api";
import { useAuthStore } from "@/stores/auth";
import { useDiskStore } from "@/stores/disk";
import { getErrorMessage } from "@/utils/error";

const router = useRouter();
const authStore = useAuthStore();
const diskStore = useDiskStore();

const fileInputRef = ref<HTMLInputElement | null>(null);
const moveDialogVisible = ref(false);
const moveCandidate = ref<UserFile | null>(null);
const moveTargetIdentity = ref("");
const moveTargetName = ref("");
const shareDialogVisible = ref(false);
const shareTarget = ref<UserFile | null>(null);
const shareExpiredTime = ref(86400);
const shareLink = ref("");
const folderTreeSupported = ref(true);
const folderTreeMessage = ref("目录树接口加载失败，请先实现 /user/folder/children 和 /user/folder/path。");

const moveTargets = computed(() =>
  diskStore.folderItems.filter((folder) => folder.identity !== moveCandidate.value?.identity),
);

const treeProps = {
  label: "name",
  isLeaf: (data: FolderTreeNode) => data.has_children !== 1,
};

onMounted(async () => {
  await authStore.bootstrap();
  await diskStore.loadRoot();
});

function triggerUpload(): void {
  fileInputRef.value?.click();
}

function getFileExt(fileName: string): string {
  const index = fileName.lastIndexOf(".");
  return index === -1 ? "" : fileName.slice(index);
}

async function loadTreeNode(node: any, resolve: (data: FolderTreeNode[]) => void): Promise<void> {
  const folderId = node.level === 0 ? 0 : Number(node.data.id);

  try {
    const data = await diskApi.listFolderChildren(folderId);
    folderTreeSupported.value = true;
    folderTreeMessage.value = "目录树接口已可用。";
    resolve(data.list);
  } catch (error) {
    if (node.level === 0) {
      folderTreeSupported.value = false;
      folderTreeMessage.value = getErrorMessage(
        error,
        "目录树接口未完成。请先实现 /user/folder/children 和 /user/folder/path。",
      );
    } else {
      ElMessage.error(getErrorMessage(error, "加载子目录失败"));
    }
    resolve([]);
  }
}

async function handleTreeNodeClick(folder: FolderTreeNode): Promise<void> {
  await diskStore.openFolder(folder);
}

async function handleFilePicked(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) {
    return;
  }

  try {
    const uploaded = await diskApi.uploadFile(file);
    await diskApi.saveUploadedRepository({
      ext: getFileExt(file.name),
      name: file.name,
      parentId: diskStore.currentFolderId,
      repositoryIdentity: uploaded.identity,
    });
    ElMessage.success("文件已上传并保存到当前目录");
    await diskStore.refresh();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "上传文件失败"));
  } finally {
    target.value = "";
  }
}

async function handleCreateFolder(): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt("输入新的文件夹名称", "新建文件夹", {
      confirmButtonText: "创建",
      inputErrorMessage: "请输入文件夹名称",
      inputPattern: /\S+/,
    });

    await diskApi.createFolder({
      name: value,
      parentId: diskStore.currentFolderId,
    });
    ElMessage.success("文件夹创建成功");
    await diskStore.refresh();
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(getErrorMessage(error, "创建文件夹失败"));
    }
  }
}

async function handleRename(file: UserFile): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt("输入新的名称", "重命名", {
      confirmButtonText: "保存",
      inputPattern: /\S+/,
      inputValue: file.name,
    });

    await diskApi.renameFile({
      identity: file.identity,
      name: value,
    });
    ElMessage.success("重命名成功");
    await diskStore.refresh();
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(getErrorMessage(error, "重命名失败"));
    }
  }
}

async function handleDelete(file: UserFile): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除 ${file.name} 吗？`, "删除确认", {
      confirmButtonText: "删除",
      type: "warning",
    });

    await diskApi.deleteFile({ identity: file.identity });
    ElMessage.success("删除成功");
    await diskStore.refresh();
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(getErrorMessage(error, "删除失败"));
    }
  }
}

async function handleOpen(file: UserFile): Promise<void> {
  if (file.is_dir === 1) {
    await diskStore.openFolder(file);
    return;
  }

  if (file.path) {
    window.open(file.path, "_blank", "noopener,noreferrer");
  } else {
    ElMessage.warning("当前文件没有可访问的路径");
  }
}

function openMoveDialog(file: UserFile): void {
  moveCandidate.value = file;
  moveTargetIdentity.value = "";
  moveTargetName.value = "";

  if (!folderTreeSupported.value && moveTargets.value.length === 0) {
    ElMessage.info("当前页没有可移动到的目标文件夹");
    return;
  }

  moveDialogVisible.value = true;
}

function handleMoveTargetSelect(folder: FolderTreeNode): void {
  if (moveCandidate.value?.identity === folder.identity) {
    ElMessage.warning("不能移动到自己");
    return;
  }

  moveTargetIdentity.value = folder.identity;
  moveTargetName.value = folder.name;
}

async function handleMove(): Promise<void> {
  if (!moveCandidate.value || !moveTargetIdentity.value) {
    return;
  }

  try {
    await diskApi.moveFile({
      identity: moveCandidate.value.identity,
      parent_identity: moveTargetIdentity.value,
    });
    ElMessage.success("文件移动成功");
    moveDialogVisible.value = false;
    await diskStore.refresh();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "移动文件失败"));
  }
}

function openShareDialog(file: UserFile): void {
  shareTarget.value = file;
  shareExpiredTime.value = 86400;
  shareLink.value = "";
  shareDialogVisible.value = true;
}

async function handleCreateShare(): Promise<void> {
  if (!shareTarget.value) {
    return;
  }

  try {
    const data = await diskApi.createShare({
      expired_time: shareExpiredTime.value,
      user_repository_identity: shareTarget.value.identity,
    });
    shareLink.value = `${window.location.origin}/share/${data.identity}`;
    ElMessage.success("分享链接已生成");
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "创建分享失败"));
  }
}

async function handleLogout(): Promise<void> {
  authStore.logout();
  await router.replace("/login");
}
</script>

<style scoped>
.disk-page {
  display: grid;
  gap: 20px;
}

.hero-panel {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  padding: 28px 30px;
}

.hero-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
  min-width: 260px;
}

.profile-card {
  width: 100%;
  padding: 16px;
  border-radius: var(--cd-radius-md);
  background: rgba(255, 252, 247, 0.72);
  border: 1px solid rgba(31, 107, 79, 0.08);
}

.profile-card strong {
  display: block;
  margin-top: 8px;
  font-size: 16px;
  word-break: break-all;
}

.workspace-grid {
  display: grid;
  gap: 20px;
  grid-template-columns: 320px minmax(0, 1fr);
}

.rail-panel,
.main-panel {
  padding: 24px;
}

.rail-section + .rail-section {
  margin-top: 20px;
}

.tree-section {
  min-height: 280px;
}

.tree-note {
  margin: 14px 0 12px;
  line-height: 1.6;
}

.stat-grid-rail {
  grid-template-columns: 1fr;
}

.tip-box {
  padding: 18px;
  border-radius: var(--cd-radius-md);
  background: linear-gradient(160deg, rgba(203, 124, 50, 0.1), rgba(31, 107, 79, 0.08));
}

.workspace-topbar {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: flex-start;
  margin-bottom: 18px;
}

.workspace-topbar h2 {
  margin: 0 0 10px;
  font-size: 28px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}

.crumb-button {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--cd-primary-strong);
  cursor: pointer;
}

.hidden-input {
  display: none;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}

.folder-choice-group {
  display: grid;
  gap: 12px;
}

.folder-choice {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 0;
}

.dialog-tip {
  margin: 0 0 14px;
}

.move-selection {
  display: grid;
  gap: 8px;
  margin-top: 14px;
  padding: 16px;
}

.share-panel {
  padding-top: 6px;
}

@media (max-width: 1080px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 820px) {
  .hero-panel,
  .workspace-topbar {
    flex-direction: column;
  }

  .hero-actions {
    align-items: stretch;
  }

  .toolbar {
    justify-content: flex-start;
  }
}
</style>
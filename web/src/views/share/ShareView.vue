<template>
  <div class="page-shell share-page">
    <section class="panel share-shell">
      <div class="share-copy">
        <span class="pill">Shared Resource</span>
        <h1 class="page-title">有人把这个文件分享给你了。</h1>
        <p class="page-subtitle">
          你可以直接打开，也可以在登录后把它保存到自己指定的目录。
        </p>

        <div class="share-meta panel">
          <template v-if="shareStore.detail">
            <div class="meta-row">
              <span class="muted">文件名</span>
              <strong>{{ shareStore.detail.name }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">类型</span>
              <strong>{{ shareStore.detail.ext || "文件" }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">大小</span>
              <strong>{{ formatFileSize(shareStore.detail.size) }}</strong>
            </div>
            <div class="meta-row">
              <span class="muted">资源地址</span>
              <strong class="path-text">{{ shareStore.detail.path || "-" }}</strong>
            </div>
          </template>
          <el-skeleton v-else :rows="4" animated />
        </div>
      </div>

      <div class="panel share-action-card">
        <h2>操作</h2>
        <p class="muted">
          点击保存时，如果还没登录会先跳登录；登录后会回到当前分享页，并展开目录树让你选择保存位置。
        </p>

        <div class="action-stack">
          <el-button :disabled="!shareStore.detail?.path" size="large" type="primary" @click="openSharedFile">
            直接打开文件
          </el-button>

          <el-button plain size="large" @click="handlePrepareSave">
            {{ authStore.isLoggedIn ? "选择目录并保存" : "登录后保存到我的网盘" }}
          </el-button>

          <el-button size="large" @click="goToDisk">
            {{ authStore.isLoggedIn ? "进入我的网盘" : "前往登录" }}
          </el-button>
        </div>

        <div v-if="authStore.isLoggedIn && showSavePanel" class="save-panel">
          <div class="save-head">
            <strong>选择保存目录</strong>
            <span class="muted">目录树和网盘页左侧一样，根目录也可以直接选择。</span>
          </div>

          <div class="root-option">
            <el-button :type="selectedParentId === 0 ? 'primary' : 'default'" plain @click="selectRoot">
              保存到根目录
            </el-button>
          </div>

          <el-tree
            v-if="folderTreeSupported"
            node-key="identity"
            lazy
            :load="loadTreeNode"
            :props="treeProps"
            class="share-tree"
            @node-click="handleTreeNodeClick"
          />

          <el-alert
            v-else
            :closable="false"
            :title="folderTreeMessage"
            show-icon
            type="info"
          />

          <div class="selected-box panel">
            <span class="muted">当前目标目录</span>
            <strong>{{ selectedFolderName }}</strong>
          </div>

          <el-button :loading="saving" size="large" type="primary" @click="handleSaveToDrive">
            确认保存到该目录
          </el-button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";

import * as diskApi from "@/api/disk";
import * as shareApi from "@/api/share";
import { useAuthStore } from "@/stores/auth";
import { useShareStore } from "@/stores/share";
import type { FolderTreeNode } from "@/types/api";
import { getErrorMessage } from "@/utils/error";

const props = defineProps<{
  identity: string;
}>();

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const shareStore = useShareStore();

const saving = ref(false);
const showSavePanel = ref(false);
const folderTreeSupported = ref(true);
const folderTreeMessage = ref("目录树接口加载失败，请先实现 /user/folder/children/:id。");
const selectedParentId = ref(0);
const selectedFolderName = ref("全部文件");

const treeProps = {
  label: "name",
  isLeaf: (data: FolderTreeNode) => data.has_children !== 1,
};
const shareRedirectPath = computed(() => `/share/${props.identity}?action=save`);

async function loadShare(): Promise<void> {
  if (!props.identity?.trim()) {
    shareStore.clear();
    ElMessage.warning("分享标识缺失");
    return;
  }

  try {
    await shareStore.load(props.identity);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "获取分享详情失败"));
  }
}

onMounted(async () => {
  await authStore.bootstrap();
  await loadShare();
  if (route.query.action === "save" && authStore.isLoggedIn) {
    showSavePanel.value = true;
  }
});

watch(
  () => props.identity,
  async () => {
    await loadShare();
  },
);

watch(
  () => route.query.action,
  (action) => {
    if (action === "save" && authStore.isLoggedIn) {
      showSavePanel.value = true;
    }
  },
);

function formatFileSize(size: number): string {
  if (!size) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB"];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}

function openSharedFile(): void {
  if (shareStore.detail?.path) {
    window.open(shareStore.detail.path, "_blank", "noopener,noreferrer");
  }
}

async function handlePrepareSave(): Promise<void> {
  if (!authStore.isLoggedIn) {
    await router.push({
      name: "login",
      query: {
        redirect: shareRedirectPath.value,
      },
    });
    return;
  }

  showSavePanel.value = true;
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
      folderTreeMessage.value = getErrorMessage(error, "目录树接口未完成。请先实现 /user/folder/children/:id。");
    } else {
      ElMessage.error(getErrorMessage(error, "加载目录失败"));
    }
    resolve([]);
  }
}

function handleTreeNodeClick(folder: FolderTreeNode): void {
  selectedParentId.value = folder.id;
  selectedFolderName.value = folder.name;
}

function selectRoot(): void {
  selectedParentId.value = 0;
  selectedFolderName.value = "全部文件";
}

async function handleSaveToDrive(): Promise<void> {
  if (!shareStore.detail) {
    return;
  }

  if (!authStore.isLoggedIn) {
    await router.push({
      name: "login",
      query: {
        redirect: shareRedirectPath.value,
      },
    });
    return;
  }

  if (!showSavePanel.value) {
    showSavePanel.value = true;
    return;
  }

  saving.value = true;
  try {
    await shareApi.saveSharedFile({
      parent_id: selectedParentId.value,
      repository_identity: shareStore.detail.repository_identity,
    });
    ElMessage.success(`已保存到${selectedFolderName.value}`);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "保存到网盘失败"));
  } finally {
    saving.value = false;
  }
}

async function goToDisk(): Promise<void> {
  if (authStore.isLoggedIn) {
    await router.push("/disk");
    return;
  }
  await router.push({
    name: "login",
    query: { redirect: shareRedirectPath.value },
  });
}
</script>

<style scoped>
.share-page {
  display: grid;
  place-items: center;
}

.share-shell {
  display: grid;
  gap: 22px;
  width: min(1180px, 100%);
  padding: 26px;
  grid-template-columns: minmax(0, 1.1fr) 360px;
}

.share-meta,
.share-action-card {
  padding: 20px;
  border-radius: var(--cd-radius-md);
  background: rgba(255, 252, 247, 0.72);
}

.share-meta {
  margin-top: 22px;
}

.share-action-card h2 {
  margin-top: 0;
}

.meta-row + .meta-row {
  margin-top: 18px;
}

.meta-row strong {
  display: block;
  margin-top: 6px;
  font-size: 16px;
}

.path-text {
  word-break: break-all;
}

.action-stack {
  display: grid;
  gap: 12px;
  margin-top: 24px;
}

.save-panel {
  display: grid;
  gap: 14px;
  margin-top: 22px;
}

.save-head {
  display: grid;
  gap: 6px;
}

.root-option {
  display: flex;
  justify-content: flex-start;
}

.share-tree {
  padding: 10px 12px;
  border: 1px solid rgba(31, 107, 79, 0.08);
  border-radius: 18px;
  background: rgba(255, 252, 247, 0.6);
}

.selected-box {
  display: grid;
  gap: 6px;
  padding: 16px;
  border-radius: 18px;
  background: rgba(255, 252, 247, 0.72);
}

@media (max-width: 960px) {
  .share-shell {
    grid-template-columns: 1fr;
  }
}
</style>
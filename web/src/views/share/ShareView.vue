<template>
  <div class="page-shell share-page">
    <section class="share-shell">
      <article class="panel share-copy">
        <span class="pill">共享文件</span>
        <h1 class="page-title">有人把这个文件分享给你了</h1>
        <p class="page-subtitle">你可以在线查看，若分享允许保存，也可以登录后保存到自己的网盘。</p>

        <div class="meta-grid panel">
          <template v-if="shareStore.detail">
            <div class="meta-row"><span>文件名</span><strong>{{ shareStore.detail.name }}</strong></div>
            <div class="meta-row"><span>类型</span><strong>{{ shareStore.detail.ext || "文件" }}</strong></div>
            <div class="meta-row"><span>大小</span><strong>{{ formatFileSize(shareStore.detail.size) }}</strong></div>
            <div class="meta-row"><span>分享模式</span><strong>{{ shareStore.detail.allow_download === 1 ? "可保存" : "仅预览" }}</strong></div>
          </template>
          <el-skeleton v-else :rows="4" animated />
        </div>
      </article>

      <article class="panel share-actions">
        <div class="action-head">
          <span class="panel-tag">分享操作</span>
          <h2>如何处理这个文件</h2>
          <p class="muted">支持口令访问、在线预览，以及登录后保存到你的网盘目录。</p>
        </div>

        <el-alert v-if="shareStore.detail?.need_code" :closable="false" title="该分享已启用访问口令，请先输入提取码。" type="warning" show-icon />

        <div class="access-panel panel">
          <span class="muted">访问口令</span>
          <el-input v-model="accessCode" clearable maxlength="8" placeholder="输入提取码后解锁分享" @keyup.enter="handleVerifyAccess" />
          <el-button type="primary" @click="handleVerifyAccess">验证口令</el-button>
        </div>

        <div class="action-stack">
          <el-button :disabled="!canOpenSharedFile" size="large" type="primary" @click="openSharedFile">
            {{ shareStore.detail?.allow_download === 1 ? "打开分享文件" : "在线预览" }}
          </el-button>
          <el-button :disabled="!canSaveSharedFile" plain size="large" @click="handlePrepareSave">
            {{ authStore.isLoggedIn ? "选择目录并保存" : "登录后保存到我的网盘" }}
          </el-button>
          <el-button size="large" @click="goToDisk">
            {{ authStore.isLoggedIn ? "进入我的网盘" : "前往登录" }}
          </el-button>
        </div>

        <div v-if="authStore.isLoggedIn && showSavePanel && canSaveSharedFile" class="save-panel">
          <div class="save-head">
            <strong>选择保存目录</strong>
            <span class="muted">这里复用了网盘页同一套目录树接口，根目录也支持直接选择。</span>
          </div>

          <div class="root-option">
            <el-button :type="selectedParentId === 0 ? 'primary' : 'default'" plain @click="selectRoot">保存到根目录</el-button>
          </div>

          <el-tree v-if="folderTreeSupported" node-key="identity" lazy :load="loadTreeNode" :props="treeProps" class="share-tree" @node-click="handleTreeNodeClick" />
          <el-alert v-else :closable="false" :title="folderTreeMessage" show-icon type="info" />

          <div class="selected-box panel">
            <span class="muted">当前目标目录</span>
            <strong>{{ selectedFolderName }}</strong>
          </div>

          <el-button :loading="saving" size="large" type="primary" @click="handleSaveToDrive">确认保存到该目录</el-button>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import type { TreeOptionProps } from "element-plus";
import { useRoute, useRouter } from "vue-router";

import * as diskApi from "@/api/disk";
import * as shareApi from "@/api/share";
import { useAuthStore } from "@/stores/auth";
import { useShareStore } from "@/stores/share";
import type { FolderTreeNode } from "@/types/api";
import { getErrorMessage } from "@/utils/error";

const props = defineProps<{ identity: string }>();

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
const accessCode = ref("");

const treeProps: TreeOptionProps = {
  label: "name",
  isLeaf: (data) => (data as FolderTreeNode).has_children !== 1,
};
const shareRedirectPath = computed(() => `/share/${props.identity}?action=save`);
// 这两个计算属性把分享页的权限判断集中管理，
// 避免模板里散落大量 need_code / allow_download 条件。
const canOpenSharedFile = computed(() => Boolean(shareStore.detail?.path && !shareStore.detail?.need_code));
const canSaveSharedFile = computed(() => Boolean(shareStore.detail && !shareStore.detail.need_code && shareStore.detail.allow_download === 1));

async function loadShare(): Promise<void> {
  if (!props.identity?.trim()) {
    shareStore.clear();
    ElMessage.warning("分享标识缺失");
    return;
  }

  try {
    // 分享详情和口令校验共用同一个接口，输入口令后重新请求即可解锁。
    await shareStore.load(props.identity, accessCode.value.trim() || undefined);
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

watch(() => props.identity, async () => { await loadShare(); });
// 登录后从 redirect 回来时，如果原本目标是“保存到网盘”，自动展开保存面板。
watch(() => route.query.action, (action) => { if (action === "save" && authStore.isLoggedIn) showSavePanel.value = true; });

function formatFileSize(size: number): string {
  if (!size) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = size;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}

async function handleVerifyAccess(): Promise<void> {
  // 口令验证不做本地缓存，直接以后端校验结果为准。
  await loadShare();
}

function openSharedFile(): void {
  if (shareStore.detail?.path && !shareStore.detail.need_code) {
    window.open(shareStore.detail.path, "_blank", "noopener,noreferrer");
  }
}

async function handlePrepareSave(): Promise<void> {
  if (!canSaveSharedFile.value) {
    ElMessage.info("当前分享仅支持预览，不能保存到你的网盘。");
    return;
  }
  // 未登录时先走登录重定向，回来后继续停留在当前分享页。
  if (!authStore.isLoggedIn) {
    await router.push({ name: "login", query: { redirect: shareRedirectPath.value } });
    return;
  }
  showSavePanel.value = true;
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

// 保存分享文件时带上 share_identity + access_code，后端会再次校验权限，
// 这样即便用户绕过页面，也不能直接跳过口令把文件保存进自己的网盘。
async function handleSaveToDrive(): Promise<void> {
  if (!shareStore.detail) return;
  if (!authStore.isLoggedIn) {
    await router.push({ name: "login", query: { redirect: shareRedirectPath.value } });
    return;
  }
  if (!canSaveSharedFile.value) {
    ElMessage.info("当前分享不允许保存。");
    return;
  }

  saving.value = true;
  try {
    await shareApi.saveSharedFile({
      access_code: accessCode.value.trim() || undefined,
      parent_id: selectedParentId.value,
      share_identity: props.identity,
    });
    ElMessage.success(`已保存到 ${selectedFolderName.value}`);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "保存到网盘失败"));
  } finally {
    saving.value = false;
  }
}

async function goToDisk(): Promise<void> {
  if (authStore.isLoggedIn) {
    await router.push({ name: "disk" });
    return;
  }
  await router.push({ name: "login", query: { redirect: shareRedirectPath.value } });
}
</script>

<style scoped>
.share-page { min-height: 100vh; padding: 24px; background: linear-gradient(180deg, #f5f9ff 0%, #eef4ff 100%); }
.share-shell { display: grid; gap: 18px; grid-template-columns: minmax(0, 1.1fr) minmax(320px, .9fr); }
.share-copy,.share-actions,.access-panel,.selected-box { padding: 24px; }
.pill,.panel-tag { display: inline-flex; padding: 5px 10px; border-radius: 999px; background: rgba(22,119,255,.1); color: #0f5fd6; font-size: 12px; font-weight: 700; }
.page-title { margin: 12px 0 10px; font-size: 30px; }
.page-subtitle,.muted { color: #6b7280; }
.meta-grid,.action-stack,.save-panel { display: grid; gap: 14px; }
.meta-row { display: flex; justify-content: space-between; gap: 12px; padding: 14px 16px; border-radius: 14px; background: rgba(255,255,255,.7); }
.action-head { margin-bottom: 18px; }
.access-panel { display: grid; gap: 12px; margin-bottom: 18px; }
.root-option { display: flex; margin: 14px 0; }
.selected-box { margin: 16px 0; }
@media (max-width: 960px) { .share-shell { grid-template-columns: 1fr; } }
@media (max-width: 600px) {
  .share-page { padding: 12px; }
  .share-copy, .share-actions, .access-panel, .selected-box { padding: 16px; }
  .page-title { font-size: 22px; margin: 10px 0 8px; }
  .page-subtitle { font-size: 14px; }
  .meta-row { flex-direction: column; align-items: flex-start; gap: 4px; padding: 12px; }
  .action-stack .el-button { width: 100%; }
  .share-tree { max-height: 240px; overflow: auto; }
  .access-panel { margin-bottom: 14px; }
}
</style>

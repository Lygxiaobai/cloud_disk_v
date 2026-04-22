<template>
  <el-table v-if="!isMobile" :data="items" :loading="loading" empty-text="当前没有分享记录" row-key="identity">
    <el-table-column label="文件" min-width="300">
      <template #default="{ row }">
        <div class="share-main">
          <strong>{{ row.name }}</strong>
          <small>{{ row.ext || "文件" }}</small>
        </div>
      </template>
    </el-table-column>

    <el-table-column label="大小" width="120" align="right">
      <template #default="{ row }">
        <span class="cell-text">{{ formatFileSize(row.size) }}</span>
      </template>
    </el-table-column>

    <el-table-column label="访问口令" width="110" align="center">
      <template #default="{ row }">
        <el-tag :type="row.access_code_set ? 'warning' : 'info'" effect="plain">
          {{ row.access_code_set ? "已启用" : "无" }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="下载权限" width="120" align="center">
      <template #default="{ row }">
        <el-tag :type="row.allow_download === 1 ? 'success' : 'danger'" effect="plain">
          {{ row.allow_download === 1 ? "允许保存" : "仅预览" }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="访问量" width="92" align="center">
      <template #default="{ row }">
        <span class="cell-text">{{ row.click_num }}</span>
      </template>
    </el-table-column>

    <el-table-column label="过期时间" min-width="170">
      <template #default="{ row }">
        <span class="cell-text">{{ row.expires_at || "永久有效" }}</span>
      </template>
    </el-table-column>

    <el-table-column label="状态" width="100" align="center">
      <template #default="{ row }">
        <el-tag :type="row.expired ? 'danger' : 'success'" effect="plain">
          {{ row.expired ? "已过期" : "生效中" }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="220" fixed="right">
      <template #default="{ row }">
        <div class="action-row">
          <el-button link type="primary" @click="$emit('copy', row)">复制链接</el-button>
          <el-button link type="primary" @click="$emit('open', row)">打开分享页</el-button>
          <el-button link type="danger" @click="$emit('delete', row)">取消分享</el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div v-else class="share-card-list" v-loading="loading">
    <el-empty v-if="!loading && items.length === 0" description="当前没有分享记录" />
    <article v-for="row in items" :key="row.identity" class="share-card">
      <div class="share-card-head">
        <div class="share-main">
          <strong>{{ row.name }}</strong>
          <small>{{ row.ext || "文件" }} · {{ formatFileSize(row.size) }}</small>
        </div>
        <el-tag :type="row.expired ? 'danger' : 'success'" effect="plain" size="small">
          {{ row.expired ? "已过期" : "生效中" }}
        </el-tag>
      </div>

      <div class="share-card-tags">
        <el-tag :type="row.access_code_set ? 'warning' : 'info'" effect="plain" size="small">
          {{ row.access_code_set ? "口令已启用" : "无口令" }}
        </el-tag>
        <el-tag :type="row.allow_download === 1 ? 'success' : 'danger'" effect="plain" size="small">
          {{ row.allow_download === 1 ? "允许保存" : "仅预览" }}
        </el-tag>
        <span class="share-card-meta">访问 {{ row.click_num }} 次</span>
      </div>

      <div class="share-card-meta">过期：{{ row.expires_at || "永久有效" }}</div>

      <div class="share-card-actions">
        <el-button size="small" type="primary" @click="$emit('copy', row)">复制链接</el-button>
        <el-button size="small" @click="$emit('open', row)">打开</el-button>
        <el-button size="small" type="danger" @click="$emit('delete', row)">取消分享</el-button>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import type { ShareListItem } from "@/types/api";

// "我的分享"需要同时展示权限信息和治理动作，
// 所以单独拆一个轻量表格组件，避免把分享管理逻辑揉进文件表格。
defineProps<{
  items: ShareListItem[];
  loading: boolean;
}>();

defineEmits<{
  copy: [item: ShareListItem];
  delete: [item: ShareListItem];
  open: [item: ShareListItem];
}>();

const isMobile = ref(false);
const mql = typeof window !== "undefined" ? window.matchMedia("(max-width: 600px)") : null;

function handleMqlChange(event: MediaQueryListEvent): void {
  isMobile.value = event.matches;
}

onMounted(() => {
  isMobile.value = mql?.matches ?? false;
  mql?.addEventListener("change", handleMqlChange);
});

onBeforeUnmount(() => {
  mql?.removeEventListener("change", handleMqlChange);
});

function formatFileSize(size: number): string {
  if (!size) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = size;
  let index = 0;

  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }

  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}
</script>

<style scoped>
.share-main {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.share-main strong,
.share-main small,
.cell-text {
  color: var(--cd-text-soft);
}

.share-main strong {
  color: var(--cd-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.share-card-list {
  display: grid;
  gap: 10px;
}

.share-card {
  padding: 12px 14px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: #fff;
}

.share-card-head {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  justify-content: space-between;
}

.share-card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  margin: 8px 0 6px;
}

.share-card-meta {
  color: var(--cd-text-soft);
  font-size: 12px;
}

.share-card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.share-card-actions .el-button {
  flex: 1 1 auto;
  min-width: 72px;
}
</style>

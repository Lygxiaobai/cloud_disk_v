<template>
  <el-dialog
    :model-value="modelValue"
    width="min(760px, 92vw)"
    top="5vh"
    title="文件版本历史"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="version-head">
      <div>
        <strong>{{ file?.name || "未选择文件" }}</strong>
        <p class="muted">当前版本排在最前面，历史版本按时间倒序展示。</p>
      </div>
      <el-button :disabled="!file" type="primary" @click="$emit('uploadVersion')">上传新版本</el-button>
    </div>

    <el-table v-if="!isMobile" :data="versions" :loading="loading" empty-text="当前文件还没有历史版本">
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.is_current === 1 ? 'success' : 'info'" effect="plain">
            {{ row.is_current === 1 ? "当前版本" : "历史版本" }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="名称" min-width="220">
        <template #default="{ row }">
          <div class="version-name">
            <strong>{{ row.name }}</strong>
            <small>{{ row.ext || "文件" }}</small>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="大小" width="120" align="right">
        <template #default="{ row }">
          <span class="muted">{{ formatFileSize(row.size) }}</span>
        </template>
      </el-table-column>

      <el-table-column label="来源" width="120" align="center">
        <template #default="{ row }">
          <span class="muted">{{ versionActionText(row.action) }}</span>
        </template>
      </el-table-column>

      <el-table-column label="内容指纹" min-width="210">
        <template #default="{ row }">
          <code class="hash-code">{{ row.hash || "-" }}</code>
        </template>
      </el-table-column>

      <el-table-column label="时间" min-width="168">
        <template #default="{ row }">
          <span class="muted">{{ row.created_at }}</span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <!-- Current version has already been applied, so only history rows expose restore. -->
          <el-button
            v-if="row.is_current !== 1"
            link
            type="primary"
            :loading="restoringIdentity === row.identity"
            @click="$emit('restoreVersion', row)"
          >
            恢复此版本
          </el-button>
          <span v-else class="muted">当前使用中</span>
        </template>
      </el-table-column>
    </el-table>

    <div v-else class="version-card-list" v-loading="loading">
      <el-empty v-if="!loading && versions.length === 0" description="当前文件还没有历史版本" />
      <article v-for="row in versions" :key="row.identity" class="version-card">
        <div class="version-card-head">
          <el-tag :type="row.is_current === 1 ? 'success' : 'info'" effect="plain" size="small">
            {{ row.is_current === 1 ? "当前版本" : "历史版本" }}
          </el-tag>
          <span class="muted">{{ versionActionText(row.action) }} · {{ formatFileSize(row.size) }}</span>
        </div>
        <div class="version-name">
          <strong>{{ row.name }}</strong>
          <small>{{ row.ext || "文件" }}</small>
        </div>
        <div class="muted version-time">{{ row.created_at }}</div>
        <code class="hash-code">{{ row.hash || "-" }}</code>
        <div class="version-card-actions">
          <el-button
            v-if="row.is_current !== 1"
            size="small"
            type="primary"
            :loading="restoringIdentity === row.identity"
            @click="$emit('restoreVersion', row)"
          >
            恢复此版本
          </el-button>
          <span v-else class="muted">当前使用中</span>
        </div>
      </article>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import type { FileVersionItem, UserFile } from "@/types/api";

// 版本历史单独拆组件，避免主页面把版本表格、上传入口和普通文件列表耦在一起。
defineProps<{
  modelValue: boolean;
  file: UserFile | null;
  loading: boolean;
  restoringIdentity: string;
  versions: FileVersionItem[];
}>();

defineEmits<{
  restoreVersion: [version: FileVersionItem];
  uploadVersion: [];
  "update:modelValue": [value: boolean];
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

function versionActionText(action: string): string {
  if (action === "current") {
    return "当前";
  }
  if (action === "replace") {
    return "替换";
  }
  if (action === "restore") {
    return "恢复";
  }
  return action || "-";
}
</script>

<style scoped>
.version-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.version-head strong {
  display: block;
  margin-bottom: 4px;
  font-size: 16px;
}

.version-name {
  display: grid;
  gap: 4px;
}

.version-name small {
  color: var(--cd-text-soft);
}

.hash-code {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(22, 119, 255, 0.08);
  color: var(--cd-primary-strong);
}

.version-card-list {
  display: grid;
  gap: 10px;
}

.version-card {
  padding: 12px 14px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: #fff;
  display: grid;
  gap: 8px;
}

.version-card-head {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.version-time {
  font-size: 12px;
}

.version-card-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}
</style>

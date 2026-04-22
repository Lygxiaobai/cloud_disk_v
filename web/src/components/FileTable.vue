<template>
  <el-table
    v-if="!isMobile"
    :data="files"
    :loading="loading"
    empty-text="当前列表没有内容"
    row-key="identity"
    @selection-change="handleSelectionChange"
  >
    <el-table-column type="selection" width="52" />

    <el-table-column label="文件名" min-width="360">
      <template #default="{ row }">
        <button class="file-link" type="button" @click="$emit('open', row)">
          <span class="file-icon" :class="iconClass(row)">
            {{ iconText(row) }}
          </span>
          <span class="file-main">
            <strong>{{ row.name }}</strong>
            <small>
              {{ row.is_dir === 1 ? "文件夹" : row.ext || "文件" }}
              <template v-if="row.path"> · {{ row.path }}</template>
            </small>
          </span>
        </button>
      </template>
    </el-table-column>

    <el-table-column label="收藏" width="74" align="center">
      <template #default="{ row }">
        <button
          class="favorite-button"
          type="button"
          :disabled="viewMode === 'recycle'"
          @click.stop="$emit('toggleFavorite', row)"
        >
          {{ row.is_favorite === 1 ? "★" : "☆" }}
        </button>
      </template>
    </el-table-column>

    <el-table-column label="大小" width="128" align="right">
      <template #default="{ row }">
        <span class="cell-text">{{ row.is_dir === 1 ? "-" : formatFileSize(row.size) }}</span>
      </template>
    </el-table-column>

    <el-table-column label="修改时间" min-width="176">
      <template #default="{ row }">
        <span class="cell-text">{{ row.updated_at || "-" }}</span>
      </template>
    </el-table-column>

    <el-table-column v-if="viewMode === 'recent'" label="最近访问" min-width="176">
      <template #default="{ row }">
        <span class="cell-text">{{ row.last_accessed_at || "-" }}</span>
      </template>
    </el-table-column>

    <el-table-column v-if="viewMode === 'duplicates'" label="重复数" width="92" align="center">
      <template #default="{ row }">
        <span class="cell-text">{{ row.duplicate_count || 0 }}</span>
      </template>
    </el-table-column>

    <el-table-column v-if="viewMode === 'duplicates'" label="重复占用" width="128" align="right">
      <template #default="{ row }">
        <span class="cell-text">{{ formatFileSize(row.duplicate_group_size || 0) }}</span>
      </template>
    </el-table-column>

    <el-table-column v-if="viewMode === 'recycle'" label="删除时间" min-width="176">
      <template #default="{ row }">
        <span class="cell-text">{{ row.deleted_at || "-" }}</span>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="360" fixed="right">
      <template #default="{ row }">
        <div class="action-row">
          <template v-if="viewMode === 'recycle'">
            <el-button link type="primary" @click="$emit('restore', row)">恢复</el-button>
            <el-button link type="danger" @click="$emit('destroy', row)">彻底删除</el-button>
          </template>

          <template v-else>
            <el-button link type="primary" @click="$emit('open', row)">
              {{ row.is_dir === 1 ? "打开" : "预览" }}
            </el-button>
            <el-button link type="primary" @click="$emit('rename', row)">重命名</el-button>
            <el-button link type="primary" @click="$emit('move', row)">移动</el-button>
            <el-button v-if="row.is_dir !== 1" link type="primary" @click="$emit('versions', row)">
              版本
            </el-button>
            <el-button v-if="row.is_dir !== 1" link type="warning" @click="$emit('share', row)">
              分享
            </el-button>
            <el-button link type="danger" @click="$emit('delete', row)">删除</el-button>
          </template>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div v-else class="file-card-list" v-loading="loading">
    <el-empty v-if="!loading && files.length === 0" description="当前列表没有内容" />
    <article
      v-for="row in files"
      :key="row.identity"
      class="file-card"
      :class="{ 'is-selected': selectedSet.has(row.identity) }"
    >
      <div class="file-card-head">
        <el-checkbox
          :model-value="selectedSet.has(row.identity)"
          @update:model-value="toggleRowSelection(row, $event)"
          @click.stop
        />
        <button class="file-link" type="button" @click="$emit('open', row)">
          <span class="file-icon" :class="iconClass(row)">{{ iconText(row) }}</span>
          <span class="file-main">
            <strong>{{ row.name }}</strong>
            <small>{{ row.is_dir === 1 ? "文件夹" : row.ext || "文件" }}</small>
          </span>
        </button>
        <button
          class="favorite-button"
          type="button"
          :disabled="viewMode === 'recycle'"
          @click.stop="$emit('toggleFavorite', row)"
        >
          {{ row.is_favorite === 1 ? "★" : "☆" }}
        </button>
      </div>

      <div class="file-card-meta">
        <span>{{ row.is_dir === 1 ? "文件夹" : formatFileSize(row.size) }}</span>
        <span v-if="viewMode === 'recycle'">删除于 {{ row.deleted_at || "-" }}</span>
        <span v-else-if="viewMode === 'recent'">访问于 {{ row.last_accessed_at || row.updated_at || "-" }}</span>
        <span v-else-if="viewMode === 'duplicates'">重复 {{ row.duplicate_count || 0 }} 份</span>
        <span v-else>{{ row.updated_at || "-" }}</span>
      </div>

      <div class="file-card-actions">
        <template v-if="viewMode === 'recycle'">
          <el-button size="small" type="primary" @click="$emit('restore', row)">恢复</el-button>
          <el-button size="small" type="danger" @click="$emit('destroy', row)">彻底删除</el-button>
        </template>
        <template v-else>
          <el-button size="small" type="primary" @click="$emit('open', row)">
            {{ row.is_dir === 1 ? "打开" : "预览" }}
          </el-button>
          <el-button size="small" @click="$emit('rename', row)">重命名</el-button>
          <el-dropdown trigger="click" @command="(cmd: string) => handleMoreCommand(cmd, row)">
            <el-button size="small">更多</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="move">移动到</el-dropdown-item>
                <el-dropdown-item v-if="row.is_dir !== 1" command="versions">版本历史</el-dropdown-item>
                <el-dropdown-item v-if="row.is_dir !== 1" command="share">创建分享</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { UserFile } from "@/types/api";

// 这个表格同时服务于普通文件、最近访问、重复文件、大文件和回收站，
// 所以把差异列统一收敛到 viewMode 分支里，避免多个表格组件分叉维护。
const props = defineProps<{
  files: UserFile[];
  loading: boolean;
  viewMode: "files" | "recent" | "recycle" | "duplicates" | "large";
}>();

const emit = defineEmits<{
  delete: [file: UserFile];
  destroy: [file: UserFile];
  move: [file: UserFile];
  open: [file: UserFile];
  rename: [file: UserFile];
  restore: [file: UserFile];
  selectionChange: [files: UserFile[]];
  share: [file: UserFile];
  toggleFavorite: [file: UserFile];
  versions: [file: UserFile];
}>();

const isMobile = ref(false);
const mql = typeof window !== "undefined" ? window.matchMedia("(max-width: 600px)") : null;
const selectedSet = ref<Set<string>>(new Set());

function syncMobile(): void {
  isMobile.value = mql?.matches ?? false;
}

function handleMqlChange(event: MediaQueryListEvent): void {
  isMobile.value = event.matches;
}

onMounted(() => {
  syncMobile();
  mql?.addEventListener("change", handleMqlChange);
});

onBeforeUnmount(() => {
  mql?.removeEventListener("change", handleMqlChange);
});

// files 翻页/刷新时清空选中，避免上一页选中的 identity 残留。
watch(
  () => props.files,
  () => {
    if (selectedSet.value.size > 0) {
      selectedSet.value = new Set();
      emit("selectionChange", []);
    }
  },
);

const selectedFiles = computed<UserFile[]>(() =>
  props.files.filter((file) => selectedSet.value.has(file.identity)),
);

function handleSelectionChange(selection: UserFile[]): void {
  selectedSet.value = new Set(selection.map((item) => item.identity));
  emit("selectionChange", selection);
}

function toggleRowSelection(row: UserFile, checked: boolean | string | number): void {
  const next = new Set(selectedSet.value);
  if (checked) {
    next.add(row.identity);
  } else {
    next.delete(row.identity);
  }
  selectedSet.value = next;
  emit("selectionChange", selectedFiles.value);
}

function handleMoreCommand(command: string, row: UserFile): void {
  if (command === "move") emit("move", row);
  else if (command === "versions") emit("versions", row);
  else if (command === "share") emit("share", row);
  else if (command === "delete") emit("delete", row);
}

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

function iconText(row: UserFile): string {
  // 图标文字只做快速识别，不追求覆盖所有扩展名。
  if (row.is_dir === 1) {
    return "夹";
  }

  const ext = row.ext.toLowerCase();
  if ([".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"].includes(ext)) {
    return "图";
  }
  if ([".mp4", ".mov", ".avi", ".mkv", ".webm"].includes(ext)) {
    return "视";
  }
  if ([".mp3", ".wav", ".aac", ".ogg", ".flac"].includes(ext)) {
    return "音";
  }
  if ([".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"].includes(ext)) {
    return "文";
  }
  return "件";
}

function iconClass(row: UserFile): string {
  // 颜色分组和 iconText 保持一致，让列表在高密度场景下也能快速扫读。
  if (row.is_dir === 1) {
    return "folder";
  }

  const ext = row.ext.toLowerCase();
  if ([".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"].includes(ext)) {
    return "image";
  }
  if ([".mp4", ".mov", ".avi", ".mkv", ".webm"].includes(ext)) {
    return "video";
  }
  if ([".mp3", ".wav", ".aac", ".ogg", ".flac"].includes(ext)) {
    return "audio";
  }
  if ([".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"].includes(ext)) {
    return "document";
  }
  return "default";
}
</script>

<style scoped>
.file-link {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.file-main {
  min-width: 0;
}

.file-link strong {
  display: block;
  margin-bottom: 3px;
  overflow: hidden;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-link small {
  display: block;
  overflow: hidden;
  color: var(--cd-text-soft);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 700;
}

.file-icon.folder {
  background: rgba(15, 118, 208, 0.12);
  color: #0f76d0;
}

.file-icon.image {
  background: rgba(37, 99, 235, 0.12);
  color: #2563eb;
}

.file-icon.video {
  background: rgba(59, 130, 246, 0.14);
  color: #1d4ed8;
}

.file-icon.audio {
  background: rgba(14, 165, 233, 0.14);
  color: #0284c7;
}

.file-icon.document {
  background: rgba(99, 102, 241, 0.14);
  color: #4f46e5;
}

.file-icon.default {
  background: rgba(100, 116, 139, 0.14);
  color: #475569;
}

.favorite-button {
  border: 0;
  background: transparent;
  color: #f59e0b;
  font-size: 18px;
  cursor: pointer;
}

.favorite-button:disabled {
  color: #cbd5e1;
  cursor: not-allowed;
}

.cell-text {
  color: var(--cd-text-soft);
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.file-card-list {
  display: grid;
  gap: 10px;
}

.file-card {
  padding: 12px 14px;
  border: 1px solid var(--cd-border);
  border-radius: 14px;
  background: #fff;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.file-card.is-selected {
  border-color: rgba(22, 119, 255, 0.4);
  box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.08);
}

.file-card-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-card-head .file-link {
  flex: 1 1 0;
  min-width: 0;
}

.file-card-head .favorite-button {
  flex-shrink: 0;
  font-size: 20px;
}

.file-card-meta {
  display: flex;
  gap: 10px;
  margin: 8px 0 10px 34px;
  color: var(--cd-text-soft);
  font-size: 12px;
  flex-wrap: wrap;
}

.file-card-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-left: 34px;
}

.file-card-actions .el-button {
  flex: 1 1 auto;
  min-width: 64px;
}
</style>

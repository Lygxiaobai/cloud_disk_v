<template>
  <el-table
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

    <el-table-column v-if="viewMode === 'recycle'" label="删除时间" min-width="176">
      <template #default="{ row }">
        <span class="cell-text">{{ row.deleted_at || "-" }}</span>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="300" fixed="right">
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
            <el-button v-if="row.is_dir !== 1" link type="warning" @click="$emit('share', row)">
              分享
            </el-button>
            <el-button link type="danger" @click="$emit('delete', row)">删除</el-button>
          </template>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { UserFile } from "@/types/api";

defineProps<{
  files: UserFile[];
  loading: boolean;
  viewMode: "files" | "recent" | "recycle";
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
}>();

function handleSelectionChange(selection: UserFile[]): void {
  emit("selectionChange", selection);
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
    return "档";
  }
  return "文";
}

function iconClass(row: UserFile): string {
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
</style>

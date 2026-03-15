<template>
  <el-table :data="files" :loading="loading" empty-text="这个目录还没有内容">
    <el-table-column label="名称" min-width="240">
      <template #default="{ row }">
        <button class="file-link" type="button" @click="$emit('open', row)">
          <span class="file-icon" :class="{ folder: row.is_dir === 1 }">
            {{ row.is_dir === 1 ? "DIR" : "FILE" }}
          </span>
          <span>
            <strong>{{ row.name }}</strong>
            <small>{{ row.is_dir === 1 ? "文件夹" : row.ext || "文件" }}</small>
          </span>
        </button>
      </template>
    </el-table-column>

    <el-table-column label="类型" width="120">
      <template #default="{ row }">
        <el-tag :type="row.is_dir === 1 ? 'success' : 'warning'" effect="plain" round>
          {{ row.is_dir === 1 ? "文件夹" : "文件" }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="大小" width="160">
      <template #default="{ row }">
        {{ row.is_dir === 1 ? "-" : formatFileSize(row.size) }}
      </template>
    </el-table-column>

    <el-table-column label="资源地址" min-width="220">
      <template #default="{ row }">
        <span class="muted path-cell">
          {{ row.path || "-" }}
        </span>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="310" fixed="right">
      <template #default="{ row }">
        <div class="action-row">
          <el-button link type="primary" @click="$emit('open', row)">
            {{ row.is_dir === 1 ? "进入" : "打开" }}
          </el-button>
          <el-button link type="primary" @click="$emit('rename', row)">重命名</el-button>
          <el-button link type="primary" @click="$emit('move', row)">移动</el-button>
          <el-button v-if="row.is_dir !== 1" link type="warning" @click="$emit('share', row)">
            分享
          </el-button>
          <el-button link type="danger" @click="$emit('delete', row)">删除</el-button>
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
}>();

defineEmits<{
  delete: [file: UserFile];
  move: [file: UserFile];
  open: [file: UserFile];
  rename: [file: UserFile];
  share: [file: UserFile];
}>();

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

.file-link strong {
  display: block;
  margin-bottom: 4px;
  font-size: 15px;
}

.file-link small {
  color: var(--cd-text-soft);
}

.file-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  background: rgba(203, 124, 50, 0.14);
  color: var(--cd-accent);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.file-icon.folder {
  background: rgba(31, 107, 79, 0.14);
  color: var(--cd-primary-strong);
}

.path-cell {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.5;
  word-break: break-all;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="preview?.name || '文件预览'"
    width="min(960px, 92vw)"
    top="5vh"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template v-if="preview">
      <div class="preview-shell">
        <template v-if="preview.kind === 'image' && preview.url">
          <img :src="preview.url" alt="preview" class="preview-image" />
        </template>

        <template v-else-if="preview.kind === 'video' && preview.url">
          <video :src="preview.url" class="preview-media" controls />
        </template>

        <template v-else-if="preview.kind === 'audio' && preview.url">
          <audio :src="preview.url" class="preview-audio" controls />
        </template>

        <template v-else-if="preview.kind === 'pdf' && preview.url">
          <iframe :src="preview.url" class="preview-frame" title="pdf-preview" />
        </template>

        <template v-else-if="preview.kind === 'text'">
          <div class="text-header muted" v-if="preview.truncated">当前仅展示前 4 KB 文本内容</div>
          <pre class="preview-text">{{ preview.text || "(空文件)" }}</pre>
        </template>

        <template v-else>
          <div class="download-box panel">
            <p class="muted">这个文件类型暂时走下载/新窗口预览。</p>
            <el-button v-if="preview.url" type="primary" @click="openInNewTab">打开文件</el-button>
          </div>
        </template>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { FilePreviewResponse } from "@/types/api";

const props = defineProps<{
  modelValue: boolean;
  preview: FilePreviewResponse | null;
}>();

defineEmits<{
  "update:modelValue": [value: boolean];
}>();

function openInNewTab(): void {
  if (props.preview?.url) {
    window.open(props.preview.url, "_blank", "noopener,noreferrer");
  }
}
</script>

<style scoped>
.preview-shell {
  min-height: 200px;
}

.preview-image,
.preview-media,
.preview-frame {
  width: 100%;
  max-height: 72vh;
  border: 0;
  border-radius: 18px;
  background: rgba(31, 42, 36, 0.05);
}

.preview-audio {
  width: 100%;
  margin-top: 18px;
}

.preview-text {
  margin: 0;
  max-height: 72vh;
  overflow: auto;
  padding: 18px;
  border-radius: 18px;
  background: rgba(31, 42, 36, 0.05);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

.text-header {
  margin-bottom: 10px;
}

.download-box {
  display: grid;
  gap: 14px;
  padding: 24px;
}

@media (max-width: 600px) {
  .preview-image,
  .preview-media,
  .preview-frame {
    max-height: 60vh;
    border-radius: 12px;
  }

  .preview-text {
    padding: 12px;
    font-size: 13px;
    max-height: 60vh;
    border-radius: 12px;
  }

  .download-box {
    padding: 16px;
  }
}
</style>

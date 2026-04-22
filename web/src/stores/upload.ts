import OSS from "ali-oss";
import { computed, ref } from "vue";
import { defineStore } from "pinia";

import * as diskApi from "@/api/disk";
import type { UploadInitResponse, UploadStsRefreshResponse } from "@/types/api";
import { computeFileMD5 } from "@/utils/hash";
import { getErrorMessage } from "@/utils/error";

type UploadStatus = "hashing" | "waiting" | "uploading" | "paused" | "completing" | "success" | "error";

interface UploadContext {
  parentId: number;
  parentIdentity?: string;
  targetFileIdentity?: string;
}

export interface UploadTask {
  checkpoint?: Record<string, unknown>;
  client?: any;
  errorMessage: string;
  etaSeconds: number | null;
  file: File;
  fileIdentity?: string;
  hash?: string;
  hashProgress: number;
  id: string;
  name: string;
  needsFinalize: boolean;
  objectKey?: string;
  progress: number;
  repositoryIdentity?: string;
  sessionIdentity?: string;
  size: number;
  speedBytes: number;
  status: UploadStatus;
}

interface PersistedResume {
	checkpoint?: Record<string, unknown>;
	fileName: string;
  fileSize: number;
  hash: string;
  objectKey: string;
  sessionIdentity: string;
}

const RESUME_PREFIX = "cloud-disk:upload:resume:";

function createTask(file: File): UploadTask {
  return {
    errorMessage: "",
    etaSeconds: null,
    file,
    hashProgress: 0,
    id: crypto.randomUUID(),
    name: file.name,
    needsFinalize: false,
    progress: 0,
    size: file.size,
    speedBytes: 0,
    status: "hashing",
  };
}

function fileExt(name: string): string {
  const index = name.lastIndexOf(".");
  return index === -1 ? "" : name.slice(index);
}

function getResumeKey(hash: string, file: File): string {
  return `${RESUME_PREFIX}${hash}:${file.size}:${file.name}`;
}

function stripCheckpoint(checkpoint: any): Record<string, unknown> | undefined {
  if (!checkpoint || typeof checkpoint !== "object") {
    return undefined;
  }

  const { file, ...rest } = checkpoint as Record<string, unknown>;
  return rest;
}

function readResume(hash: string, file: File): PersistedResume | null {
  const raw = localStorage.getItem(getResumeKey(hash, file));
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as PersistedResume;
  } catch {
    localStorage.removeItem(getResumeKey(hash, file));
    return null;
  }
}

function persistResume(task: UploadTask): void {
	if (!task.hash || !task.sessionIdentity || !task.objectKey) {
		return;
	}

  // Only persist resumable upload metadata; the browser file object itself
  // stays in memory and is reselected by the user when needed.
  const payload: PersistedResume = {
    checkpoint: task.checkpoint,
    fileName: task.name,
    fileSize: task.size,
    hash: task.hash,
    objectKey: task.objectKey,
    sessionIdentity: task.sessionIdentity,
  };
  localStorage.setItem(getResumeKey(task.hash, task.file), JSON.stringify(payload));
}

function clearResume(task: UploadTask): void {
  if (!task.hash) {
    return;
  }
  localStorage.removeItem(getResumeKey(task.hash, task.file));
}

function getRefreshInterval(expiration?: string): number {
  if (!expiration) {
    return 5 * 60 * 1000;
  }

  const expiresAt = Date.parse(expiration);
  if (Number.isNaN(expiresAt)) {
    return 5 * 60 * 1000;
  }

  // Refresh a bit ahead of expiry, but avoid refreshing too frequently.
  const safeWindow = expiresAt - Date.now() - 60 * 1000;
  return Math.max(60 * 1000, Math.min(5 * 60 * 1000, safeWindow));
}

async function refreshTaskSTS(task: UploadTask): Promise<{
  accessKeyId: string;
  accessKeySecret: string;
  stsToken: string;
}> {
  if (!task.sessionIdentity) {
    throw new Error("上传会话缺失");
  }

  const refreshed = await diskApi.refreshUploadSTS({ session_identity: task.sessionIdentity });
  task.objectKey = refreshed.object_key;

  return {
    accessKeyId: refreshed.sts.access_key_id,
    accessKeySecret: refreshed.sts.access_key_secret,
    stsToken: refreshed.sts.security_token,
  };
}

function createOSSClient(task: UploadTask, config: UploadInitResponse | UploadStsRefreshResponse): any {
  return new OSS({
    accessKeyId: config.sts?.access_key_id,
    accessKeySecret: config.sts?.access_key_secret,
    bucket: config.oss_bucket,
    endpoint: `https://${config.oss_endpoint}`,
    refreshSTSToken: async () => refreshTaskSTS(task),
    refreshSTSTokenInterval: getRefreshInterval(config.sts?.expiration),
    region: config.oss_region,
    stsToken: config.sts?.security_token,
  });
}

function getPartSize(fileSize: number): number {
  if (fileSize <= 20 * 1024 * 1024) {
    return 1024 * 1024;
  }
  if (fileSize <= 200 * 1024 * 1024) {
    return 2 * 1024 * 1024;
  }
  return 5 * 1024 * 1024;
}

export const useUploadStore = defineStore("upload", () => {
  const tasks = ref<UploadTask[]>([]);

  const activeCount = computed(() => tasks.value.filter((task) => task.status === "hashing" || task.status === "waiting" || task.status === "uploading" || task.status === "completing").length);

  async function addFiles(
    files: File[],
    context: UploadContext,
    onSuccess?: () => Promise<void>,
  ): Promise<void> {
    for (const file of files) {
      const task = createTask(file);
      tasks.value.unshift(task);
      void prepareTask(task, context, onSuccess);
    }
  }

async function prepareTask(task: UploadTask, context: UploadContext, onSuccess?: () => Promise<void>): Promise<void> {
    task.status = "hashing";
    task.errorMessage = "";
    task.hashProgress = 0;
    task.speedBytes = 0;
    task.etaSeconds = null;

    try {
      task.hash = await computeFileMD5(task.file, (progress) => {
        task.hashProgress = progress;
      });

      // If the same local file already has a saved checkpoint, resume that
      // OSS multipart session instead of creating a new one.
      const resume = readResume(task.hash, task.file);
      if (resume?.sessionIdentity && resume.objectKey) {
        try {
          const refreshed = await diskApi.refreshUploadSTS({ session_identity: resume.sessionIdentity });
          task.sessionIdentity = refreshed.session_identity;
          task.objectKey = refreshed.object_key;
          task.checkpoint = resume.checkpoint;
          await uploadWithOSS(task, refreshed, onSuccess);
          return;
        } catch {
          clearResume(task);
        }
      }

      const initResp = await diskApi.uploadInit({
        ext: fileExt(task.name),
        hash: task.hash,
        name: task.name,
        parent_id: context.parentId,
        parent_identity: context.parentIdentity,
        size: task.size,
        target_file_identity: context.targetFileIdentity,
      });

      if (initResp.instant_hit) {
        // 秒传命中: no physical upload is needed, the server only creates
        // the logical user file entry and returns immediately.
        task.progress = 1;
        task.status = "success";
        task.fileIdentity = initResp.file_identity;
        task.repositoryIdentity = initResp.repository_identity;
        clearResume(task);
        if (onSuccess) {
          await onSuccess();
        }
        return;
      }

      task.sessionIdentity = initResp.session_identity;
      task.objectKey = initResp.object_key;
      task.checkpoint = undefined;
      persistResume(task);
      await uploadWithOSS(task, initResp, onSuccess);
    } catch (error) {
      task.status = "error";
      task.errorMessage = getErrorMessage(error, "准备上传失败");
    }
  }

async function uploadWithOSS(
  task: UploadTask,
  config: UploadInitResponse | UploadStsRefreshResponse,
  onSuccess?: () => Promise<void>,
): Promise<void> {
    if (!task.objectKey || !task.sessionIdentity) {
      task.status = "error";
      task.errorMessage = "上传会话缺失";
      return;
    }

    task.client = createOSSClient(task, config);
    task.status = "uploading";
    task.errorMessage = "";

    let previousTransferred = task.progress * task.size;
    let previousTime = performance.now();

    try {
      await task.client.multipartUpload(task.objectKey, task.file, {
        checkpoint: task.checkpoint ? { ...task.checkpoint } : undefined,
        parallel: 3,
        partSize: getPartSize(task.size),
        progress: async (percentage: number, checkpoint: any) => {
          const safePercentage = Number.isFinite(percentage) ? Math.max(0, Math.min(1, percentage)) : task.progress;
          const transferred = safePercentage * task.size;
          const now = performance.now();
          const deltaBytes = Math.max(0, transferred - previousTransferred);
          const deltaSeconds = Math.max((now - previousTime) / 1000, 0.001);
          const speed = deltaBytes / deltaSeconds;

          task.progress = safePercentage;
          task.checkpoint = stripCheckpoint(checkpoint);
          task.speedBytes = speed;
          task.etaSeconds = speed > 0 ? Math.max(0, (task.size - transferred) / speed) : null;
          task.status = "uploading";

          previousTransferred = transferred;
          previousTime = now;
          persistResume(task);
        },
      });
    } catch (error: any) {
      if (error?.name === "cancel") {
        task.status = "paused";
        task.errorMessage = "";
        persistResume(task);
        return;
      }
      task.status = "error";
      task.errorMessage = getErrorMessage(error, "上传失败");
      persistResume(task);
      return;
    }

    task.needsFinalize = true;
    task.status = "completing";
    task.progress = 1;
    task.speedBytes = 0;
    task.etaSeconds = 0;

    await finalizeUpload(task, onSuccess);
  }

  async function finalizeUpload(task: UploadTask, onSuccess?: () => Promise<void>): Promise<void> {
    if (!task.sessionIdentity) {
      task.status = "error";
      task.errorMessage = "上传会话缺失";
      return;
    }

    try {
      const resp = await diskApi.uploadComplete({ session_identity: task.sessionIdentity });
      task.fileIdentity = resp.file_identity;
      task.repositoryIdentity = resp.repository_identity;
      task.needsFinalize = false;
      task.status = "success";
      clearResume(task);
      if (onSuccess) {
        await onSuccess();
      }
    } catch (error) {
      task.status = "error";
      task.errorMessage = getErrorMessage(error, "完成上传失败");
      persistResume(task);
    }
  }

  function pauseTask(taskId: string): void {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task) {
      return;
    }

    if (task.status === "uploading" && task.client) {
      task.client.cancel();
      return;
    }

    if (task.status === "waiting") {
      task.status = "paused";
    }
  }

  async function resumeTask(taskId: string, onSuccess?: () => Promise<void>): Promise<void> {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task || !task.sessionIdentity) {
      return;
    }

    if (task.needsFinalize) {
      task.status = "completing";
      await finalizeUpload(task, onSuccess);
      return;
    }

    try {
      const refreshed = await diskApi.refreshUploadSTS({ session_identity: task.sessionIdentity });
      task.objectKey = refreshed.object_key;
      task.status = "waiting";
      await uploadWithOSS(task, refreshed, onSuccess);
    } catch (error) {
      task.status = "error";
      task.errorMessage = getErrorMessage(error, "继续上传失败");
    }
  }

  async function retryTask(taskId: string, context: UploadContext, onSuccess?: () => Promise<void>): Promise<void> {
    const task = tasks.value.find((item) => item.id === taskId);
    if (!task) {
      return;
    }

    if (task.sessionIdentity) {
      await resumeTask(taskId, onSuccess);
      return;
    }

    await prepareTask(task, context, onSuccess);
  }

  function removeTask(taskId: string): void {
    const index = tasks.value.findIndex((task) => task.id === taskId);
    if (index === -1) {
      return;
    }

    const [task] = tasks.value.splice(index, 1);
    if (task) {
      clearResume(task);
    }
  }

  function clearFinished(): void {
    const finished = tasks.value.filter((task) => task.status === "success");
    finished.forEach(clearResume);
    tasks.value = tasks.value.filter((task) => task.status !== "success");
  }

  return {
    activeCount,
    addFiles,
    clearFinished,
    pauseTask,
    removeTask,
    resumeTask,
    retryTask,
    tasks,
  };
});

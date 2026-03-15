import { computed, ref } from "vue";
import { defineStore } from "pinia";

import * as diskApi from "@/api/disk";
import type { FolderPathItem, FolderTreeNode, UserFile } from "@/types/api";

interface Breadcrumb {
  id: number;
  identity?: string;
  name: string;
}

interface FolderTarget {
  id: number;
  identity: string;
  name: string;
}

const ROOT_BREADCRUMB: Breadcrumb = {
  id: 0,
  name: "全部文件",
};

function normalizePathItems(items: FolderPathItem[]): Breadcrumb[] {
  const mapped = items.map((item) => ({
    id: item.id,
    identity: item.identity,
    name: item.name,
  }));

  if (mapped.length === 0) {
    return [ROOT_BREADCRUMB];
  }

  if (mapped[0].id === 0) {
    return mapped;
  }

  return [ROOT_BREADCRUMB, ...mapped];
}

export const useDiskStore = defineStore("disk", () => {
  const items = ref<UserFile[]>([]);
  const breadcrumbs = ref<Breadcrumb[]>([ROOT_BREADCRUMB]);
  const currentFolderId = ref(0);
  const page = ref(1);
  const size = ref(20);
  const total = ref(0);
  const loading = ref(false);

  const folderItems = computed(() => items.value.filter((item) => item.is_dir === 1));
  const fileItems = computed(() => items.value.filter((item) => item.is_dir !== 1));
  const currentFolderName = computed(() => breadcrumbs.value[breadcrumbs.value.length - 1]?.name ?? ROOT_BREADCRUMB.name);
  const currentFolderIdentity = computed(() => breadcrumbs.value[breadcrumbs.value.length - 1]?.identity ?? "");

  async function loadFolder(
    folderId = currentFolderId.value,
    nextPage = page.value,
    folderIdentity = currentFolderIdentity.value,
  ): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      currentFolderId.value = folderId;
      const data = await diskApi.listFiles({
        identity: folderIdentity || undefined,
        id: folderId,
        page: nextPage,
        size: size.value,
      });
      items.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadRoot(): Promise<void> {
    breadcrumbs.value = [ROOT_BREADCRUMB];
    currentFolderId.value = 0;
    await loadFolder(0, 1, "");
  }

  async function refresh(): Promise<void> {
    await loadFolder(currentFolderId.value, page.value, currentFolderIdentity.value);
  }

  async function buildBreadcrumbs(folder: FolderTarget): Promise<Breadcrumb[]> {
    try {
      const data = await diskApi.fetchFolderPath(folder.identity);
      if (data.list.length > 0) {
        return normalizePathItems(data.list);
      }
    } catch {
      // 如果后端路径接口还没实现，前端退回到旧的 breadcrumb 追加逻辑。
    }

    return [
      ...breadcrumbs.value,
      {
        id: folder.id,
        identity: folder.identity,
        name: folder.name,
      },
    ];
  }

  async function openFolder(folder: FolderTarget | UserFile | FolderTreeNode): Promise<void> {
    breadcrumbs.value = await buildBreadcrumbs({
      id: folder.id,
      identity: folder.identity,
      name: folder.name,
    });
    await loadFolder(folder.id, 1, folder.identity);
  }

  async function jumpToBreadcrumb(index: number): Promise<void> {
    const target = breadcrumbs.value[index];
    if (!target) {
      return;
    }

    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1);
    await loadFolder(target.id, 1, target.identity ?? "");
  }

  async function changePage(nextPage: number): Promise<void> {
    await loadFolder(currentFolderId.value, nextPage, currentFolderIdentity.value);
  }

  return {
    breadcrumbs,
    changePage,
    currentFolderId,
    currentFolderIdentity,
    currentFolderName,
    fileItems,
    folderItems,
    items,
    jumpToBreadcrumb,
    loadFolder,
    loadRoot,
    loading,
    openFolder,
    page,
    refresh,
    size,
    total,
  };
});
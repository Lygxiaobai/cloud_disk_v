import { computed, ref } from "vue";
import { defineStore } from "pinia";

import * as diskApi from "@/api/disk";
import * as shareApi from "@/api/share";
import type { FolderPathItem, FolderTreeNode, ShareListItem, UserFile } from "@/types/api";

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

type DiskViewMode = "files" | "recent" | "recycle" | "duplicates" | "large" | "shares";

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
  const viewMode = ref<DiskViewMode>("files");
  const items = ref<UserFile[]>([]);
  const recentItems = ref<UserFile[]>([]);
  const recycleItems = ref<UserFile[]>([]);
  const shareItems = ref<ShareListItem[]>([]);
  const breadcrumbs = ref<Breadcrumb[]>([ROOT_BREADCRUMB]);
  const currentFolderId = ref(0);
  const page = ref(1);
  const size = ref(20);
  const total = ref(0);
  const loading = ref(false);
  const query = ref("");
  const fileType = ref("all");
  const favoriteOnly = ref(false);
  const orderBy = ref("");
  const orderDir = ref<"asc" | "desc">("desc");
  const searchScope = ref<"folder" | "all">("folder");
  const minSizeMB = ref(100);

  const folderItems = computed(() => items.value.filter((item) => item.is_dir === 1));
  const fileItems = computed(() => items.value.filter((item) => item.is_dir !== 1));
  const currentFolderName = computed(() => breadcrumbs.value[breadcrumbs.value.length - 1]?.name ?? ROOT_BREADCRUMB.name);
  const currentFolderIdentity = computed(() => breadcrumbs.value[breadcrumbs.value.length - 1]?.identity ?? "");
  const visibleItems = computed(() => {
    // 统一给主页面消费一个“当前视图数据源”，
    // 这样页面层不需要关心最近文件和回收站的底层数组差异。
    if (viewMode.value === "recent") {
      return recentItems.value;
    }
    if (viewMode.value === "recycle") {
      return recycleItems.value;
    }
    return items.value;
  });

  async function loadFolder(
    folderId = currentFolderId.value,
    nextPage = page.value,
    folderIdentity = currentFolderIdentity.value,
  ): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      currentFolderId.value = folderId;
      // 普通文件视图默认保留目录上下文，只有 searchScope=all 时才切全盘搜索。
      const data = await diskApi.listFiles({
        favorite_only: favoriteOnly.value || undefined,
        file_type: fileType.value !== "all" ? fileType.value : undefined,
        id: folderId,
        identity: folderIdentity || undefined,
        order_by: orderBy.value || undefined,
        order_dir: orderDir.value,
        page: nextPage,
        query: query.value || undefined,
        scope: searchScope.value,
        size: size.value,
      });
      items.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadDuplicateFiles(nextPage = page.value): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      // 重复文件治理固定从全盘维度返回，不依赖当前目录。
      const data = await diskApi.listFiles({
        favorite_only: favoriteOnly.value || undefined,
        file_type: fileType.value !== "all" ? fileType.value : undefined,
        order_by: orderBy.value || undefined,
        order_dir: orderDir.value,
        page: nextPage,
        query: query.value || undefined,
        size: size.value,
        view: "duplicates",
      });
      items.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadLargeFiles(nextPage = page.value): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      // 大文件治理也走全盘维度，并把阈值交给后端统一过滤。
      const data = await diskApi.listFiles({
        favorite_only: favoriteOnly.value || undefined,
        file_type: fileType.value !== "all" ? fileType.value : undefined,
        min_size_mb: minSizeMB.value,
        order_by: orderBy.value || undefined,
        order_dir: orderDir.value,
        page: nextPage,
        query: query.value || undefined,
        size: size.value,
        view: "large",
      });
      items.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadRecent(limit = 10): Promise<void> {
    loading.value = true;
    try {
      const data = await diskApi.listRecentFiles(limit);
      recentItems.value = data.list;
      total.value = data.list.length;
    } finally {
      loading.value = false;
    }
  }

  async function loadRecycle(nextPage = page.value): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      const data = await diskApi.listRecycleFiles({
        order_by: orderBy.value || undefined,
        order_dir: orderDir.value,
        page: nextPage,
        query: query.value || undefined,
        size: size.value,
      });
      recycleItems.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadShares(nextPage = page.value): Promise<void> {
    loading.value = true;
    try {
      page.value = nextPage;
      // “我的分享”走独立数据源，因为它本质上不是文件目录，而是分享记录列表。
      const data = await shareApi.listShares({
        page: nextPage,
        query: query.value || undefined,
        size: size.value,
      });
      shareItems.value = data.list;
      total.value = data.count;
    } finally {
      loading.value = false;
    }
  }

  async function loadRoot(): Promise<void> {
    viewMode.value = "files";
    breadcrumbs.value = [ROOT_BREADCRUMB];
    currentFolderId.value = 0;
    searchScope.value = "folder";
    await loadFolder(0, 1, "");
    void loadRecent();
  }

  async function refresh(): Promise<void> {
    // 不同视图分别走自己的数据源，避免把治理视图误刷新成目录列表。
    if (viewMode.value === "recent") {
      await loadRecent();
      return;
    }
    if (viewMode.value === "recycle") {
      await loadRecycle(page.value);
      return;
    }
    if (viewMode.value === "duplicates") {
      await loadDuplicateFiles(page.value);
      return;
    }
    if (viewMode.value === "large") {
      await loadLargeFiles(page.value);
      return;
    }
    if (viewMode.value === "shares") {
      await loadShares(page.value);
      return;
    }
    await loadFolder(currentFolderId.value, page.value, currentFolderIdentity.value);
  }

  async function buildBreadcrumbs(folder: FolderTarget): Promise<Breadcrumb[]> {
    try {
      const data = await diskApi.fetchFolderPath(folder.identity);
      if (data.list.length > 0) {
        return normalizePathItems(data.list);
      }
    } catch {
      // keep fallback path building for partially implemented backends
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
    viewMode.value = "files";
    searchScope.value = "folder";
    // 用户从治理视图或最近访问跳回目录时，自动恢复成目录浏览模式。
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

    viewMode.value = "files";
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1);
    await loadFolder(target.id, 1, target.identity ?? "");
  }

  async function changePage(nextPage: number): Promise<void> {
    if (viewMode.value === "recent") {
      return;
    }
    if (viewMode.value === "recycle") {
      await loadRecycle(nextPage);
      return;
    }
    if (viewMode.value === "duplicates") {
      await loadDuplicateFiles(nextPage);
      return;
    }
    if (viewMode.value === "large") {
      await loadLargeFiles(nextPage);
      return;
    }
    if (viewMode.value === "shares") {
      await loadShares(nextPage);
      return;
    }
    await loadFolder(currentFolderId.value, nextPage, currentFolderIdentity.value);
  }

  async function setViewMode(nextMode: DiskViewMode): Promise<void> {
    viewMode.value = nextMode;
    page.value = 1;
    // 切视图时统一重走 refresh，让分页和筛选状态都按新视图重算。
    await refresh();
  }

  async function applyFilters(): Promise<void> {
    page.value = 1;
    await refresh();
  }

  return {
    applyFilters,
    breadcrumbs,
    changePage,
    currentFolderId,
    currentFolderIdentity,
    currentFolderName,
    favoriteOnly,
    fileItems,
    fileType,
    folderItems,
    items,
    jumpToBreadcrumb,
    loadDuplicateFiles,
    loadFolder,
    loadLargeFiles,
    loadRecent,
    loadRecycle,
    loadRoot,
    loadShares,
    loading,
    minSizeMB,
    openFolder,
    orderBy,
    orderDir,
    page,
    query,
    recentItems,
    recycleItems,
    refresh,
    searchScope,
    setViewMode,
    shareItems,
    size,
    total,
    viewMode,
    visibleItems,
  };
});

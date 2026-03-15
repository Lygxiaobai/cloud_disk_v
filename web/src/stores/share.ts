import { ref } from "vue";
import { defineStore } from "pinia";

import * as shareApi from "@/api/share";
import type { ShareFileDetailResponse } from "@/types/api";

export const useShareStore = defineStore("share", () => {
  const detail = ref<ShareFileDetailResponse | null>(null);
  const loading = ref(false);

  async function load(identity: string): Promise<void> {
    loading.value = true;
    try {
      detail.value = await shareApi.fetchShareDetail(identity);
    } finally {
      loading.value = false;
    }
  }

  function clear(): void {
    detail.value = null;
  }

  return {
    clear,
    detail,
    load,
    loading,
  };
});

<template>
  <div class="auth-page page-shell">
    <section class="auth-shell">
      <article class="auth-copy">
        <span class="pill">Cloud Disk</span>
        <h1 class="page-title">一个更像桌面网盘的在线文件工作区</h1>
        <p class="page-subtitle">
          登录后可以直接进入文件管理中心，使用上传、预览、最近文件、收藏、批量操作和回收站能力。
        </p>

        <div class="feature-grid">
          <div class="feature-card">
            <strong>直传 OSS</strong>
            <span>支持秒传、分片上传、断点续传和自动 STS 续签。</span>
          </div>
          <div class="feature-card">
            <strong>文件预览</strong>
            <span>图片、视频、音频、PDF 和文本都支持快速预览。</span>
          </div>
          <div class="feature-card">
            <strong>管理体验</strong>
            <span>搜索、筛选、收藏、回收站和批量操作都已经接入。</span>
          </div>
        </div>
      </article>

      <article class="panel auth-card">
        <div class="auth-card-head">
          <span class="panel-tag">欢迎回来</span>
          <h2>登录网盘</h2>
          <p class="muted">输入账号密码，直接进入你的文件空间。</p>
        </div>

        <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
          <el-form-item label="用户名">
            <el-input v-model.trim="form.name" placeholder="请输入用户名" size="large" />
          </el-form-item>

          <el-form-item label="密码">
            <el-input
              v-model.trim="form.password"
              show-password
              placeholder="请输入密码"
              size="large"
              type="password"
            />
          </el-form-item>

          <el-button :loading="submitting" class="auth-submit" native-type="submit" size="large" type="primary">
            进入网盘
          </el-button>
        </el-form>

        <div class="auth-footer">
          <span class="muted">还没有账号？</span>
          <RouterLink to="/register">立即注册</RouterLink>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { RouterLink, useRoute, useRouter } from "vue-router";

import { useAuthStore } from "@/stores/auth";
import { getErrorMessage } from "@/utils/error";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const submitting = ref(false);
const form = reactive({
  name: "",
  password: "",
});

async function handleSubmit(): Promise<void> {
  if (!form.name || !form.password) {
    ElMessage.warning("请输入用户名和密码。");
    return;
  }

  submitting.value = true;
  try {
    await authStore.login({
      name: form.name,
      password: form.password,
    });
    ElMessage.success("登录成功。");

    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/disk";
    await router.replace(redirect);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "登录失败。"));
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.auth-page {
  display: grid;
  place-items: center;
}

.auth-shell {
  display: grid;
  gap: 26px;
  width: min(1200px, 100%);
  grid-template-columns: minmax(0, 1.15fr) minmax(380px, 440px);
}

.auth-copy {
  padding: 40px 18px 40px 8px;
}

.feature-grid {
  display: grid;
  gap: 16px;
  margin-top: 30px;
}

.feature-card {
  padding: 18px 20px;
  border: 1px solid rgba(22, 119, 255, 0.08);
  border-radius: var(--cd-radius-md);
  background: rgba(255, 255, 255, 0.72);
}

.feature-card strong {
  display: block;
  margin-bottom: 6px;
  font-size: 16px;
}

.feature-card span {
  color: var(--cd-text-soft);
}

.auth-card {
  align-self: center;
  padding: 30px;
}

.auth-card-head h2 {
  margin: 14px 0 8px;
  font-size: 30px;
}

.panel-tag {
  display: inline-flex;
  padding: 7px 12px;
  border-radius: 999px;
  background: rgba(22, 119, 255, 0.12);
  color: var(--cd-primary-strong);
  font-size: 12px;
  font-weight: 700;
}

.auth-submit {
  width: 100%;
  margin-top: 8px;
  border: 0;
  background: linear-gradient(135deg, var(--cd-primary) 0%, #3d8dff 100%);
}

.auth-footer {
  display: flex;
  gap: 8px;
  margin-top: 18px;
  font-size: 14px;
}

@media (max-width: 960px) {
  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-copy {
    padding: 8px 4px;
  }
}
</style>

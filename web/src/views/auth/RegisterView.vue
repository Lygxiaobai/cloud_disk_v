<template>
  <div class="auth-page page-shell">
    <section class="auth-shell register-shell">
      <article class="panel auth-card">
        <div class="auth-card-head">
          <span class="panel-tag">创建账号</span>
          <h2>注册网盘账号</h2>
          <p class="muted">邮箱验证码时效较短，建议收到后尽快完成注册。</p>
        </div>

        <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
          <el-form-item label="用户名">
            <el-input v-model.trim="form.name" placeholder="请输入用户名" size="large" />
          </el-form-item>

          <el-form-item label="邮箱">
            <el-input v-model.trim="form.email" placeholder="name@example.com" size="large" />
          </el-form-item>

          <el-form-item label="密码">
            <el-input
              v-model.trim="form.password"
              placeholder="请输入密码"
              show-password
              size="large"
              type="password"
            />
          </el-form-item>

          <el-form-item label="邮箱验证码">
            <div class="code-row">
              <el-input v-model.trim="form.code" placeholder="请输入验证码" size="large" />
              <el-button :disabled="countdown > 0" :loading="sendingCode" size="large" @click="handleSendCode">
                {{ countdown > 0 ? `${countdown}s` : "发送验证码" }}
              </el-button>
            </div>
          </el-form-item>

          <el-button :loading="submitting" class="auth-submit" native-type="submit" size="large" type="primary">
            注册账号
          </el-button>
        </el-form>

        <div class="auth-footer">
          <span class="muted">已经有账号？</span>
          <RouterLink to="/login">返回登录</RouterLink>
        </div>
      </article>

      <article class="auth-copy">
        <span class="pill">Register Flow</span>
        <h1 class="page-title">验证邮箱后，把你的上传、目录和分享都归入同一个个人工作区</h1>
        <p class="page-subtitle">
          当前前端已经串好了邮箱验证码、注册、登录和刷新令牌流程，注册完成后就能直接使用文件管理能力。
        </p>

        <div class="feature-grid">
          <div class="feature-card">
            <strong>邮箱验证</strong>
            <span>前端表单直接对接当前后端的验证码注册流程。</span>
          </div>
          <div class="feature-card">
            <strong>状态恢复</strong>
            <span>Pinia 会持久化令牌，刷新页面后仍可恢复登录状态。</span>
          </div>
          <div class="feature-card">
            <strong>后续扩展</strong>
            <span>后面继续加设备管理、登录审计或多端同步也比较顺。</span>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { RouterLink, useRouter } from "vue-router";

import { useAuthStore } from "@/stores/auth";
import { getErrorMessage } from "@/utils/error";

const router = useRouter();
const authStore = useAuthStore();

const form = reactive({
  code: "",
  email: "",
  name: "",
  password: "",
});

const sendingCode = ref(false);
const submitting = ref(false);
const countdown = ref(0);
let timer: number | undefined;

function startCountdown(): void {
  countdown.value = 60;
  timer = window.setInterval(() => {
    countdown.value -= 1;
    if (countdown.value <= 0 && timer) {
      window.clearInterval(timer);
      timer = undefined;
    }
  }, 1000);
}

async function handleSendCode(): Promise<void> {
  if (!form.email) {
    ElMessage.warning("请先输入邮箱。");
    return;
  }

  sendingCode.value = true;
  try {
    await authStore.sendRegisterCode(form.email);
    ElMessage.success("验证码已发送。");
    startCountdown();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "发送验证码失败。"));
  } finally {
    sendingCode.value = false;
  }
}

async function handleSubmit(): Promise<void> {
  if (!form.name || !form.email || !form.password || !form.code) {
    ElMessage.warning("请完整填写所有字段。");
    return;
  }

  submitting.value = true;
  try {
    await authStore.registerAccount({
      code: form.code,
      email: form.email,
      name: form.name,
      password: form.password,
    });
    ElMessage.success("注册成功，请登录。");
    await router.replace("/login");
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "注册失败。"));
  } finally {
    submitting.value = false;
  }
}

onBeforeUnmount(() => {
  if (timer) {
    window.clearInterval(timer);
  }
});
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
  grid-template-columns: minmax(380px, 450px) minmax(0, 1.15fr);
}

.auth-copy {
  padding: 40px 8px 40px 18px;
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

.code-row {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) 132px;
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
    order: -1;
    padding: 8px 4px;
  }

  .code-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .auth-card {
    padding: 20px 18px;
  }

  .auth-card-head h2 {
    font-size: 22px;
  }

  .feature-grid {
    margin-top: 18px;
    gap: 10px;
  }

  .feature-card {
    padding: 14px 16px;
  }

  .feature-card strong {
    font-size: 15px;
  }

  .page-title {
    font-size: 22px;
  }

  .page-subtitle {
    font-size: 14px;
  }

  .code-row {
    grid-template-columns: 1fr;
    gap: 8px;
  }
}
</style>

<template>
  <div class="auth-page page-shell">
    <section class="auth-grid">
      <article class="auth-copy">
        <span class="pill">Cloud Disk</span>
        <h1 class="page-title">Pull files, sharing, and access control into one focused workspace.</h1>
        <p class="page-subtitle">
          This first frontend cut focuses on the main path: sign in, refresh tokens, file listing,
          uploads, and sharing.
        </p>

        <div class="auth-highlights">
          <div class="auth-feature">
            <strong>Short-lived access token</strong>
            <span>Protected APIs use a single `Authorization` header.</span>
          </div>
          <div class="auth-feature">
            <strong>Refresh on demand</strong>
            <span>`401` triggers a refresh attempt, then replays the original request.</span>
          </div>
          <div class="auth-feature">
            <strong>One gateway</strong>
            <span>The frontend talks to `/api/*` and lets Nginx bridge the backend.</span>
          </div>
        </div>
      </article>

      <article class="panel auth-card">
        <div class="auth-card-head">
          <span class="pill auth-pill">Welcome back</span>
          <h2>Sign in</h2>
          <p class="muted">Enter your account and jump straight into the drive.</p>
        </div>

        <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
          <el-form-item label="Username">
            <el-input v-model.trim="form.name" placeholder="admin" size="large" />
          </el-form-item>

          <el-form-item label="Password">
            <el-input
              v-model.trim="form.password"
              show-password
              placeholder="Enter your password"
              size="large"
              type="password"
            />
          </el-form-item>

          <el-button :loading="submitting" class="auth-submit" native-type="submit" size="large" type="primary">
            Open my drive
          </el-button>
        </el-form>

        <div class="auth-footer">
          <span class="muted">Need an account?</span>
          <RouterLink to="/register">Create one</RouterLink>
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
    ElMessage.warning("Please enter both username and password.");
    return;
  }

  submitting.value = true;
  try {
    await authStore.login({
      name: form.name,
      password: form.password,
    });
    ElMessage.success("Signed in successfully.");

    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/disk";
    await router.replace(redirect);
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "Sign in failed."));
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

.auth-grid {
  display: grid;
  gap: 24px;
  width: min(1180px, 100%);
  grid-template-columns: minmax(0, 1.1fr) minmax(380px, 460px);
}

.auth-copy {
  padding: 38px;
}

.auth-highlights {
  display: grid;
  gap: 16px;
  margin-top: 28px;
}

.auth-feature {
  padding: 18px 20px;
  border: 1px solid rgba(31, 107, 79, 0.1);
  border-radius: var(--cd-radius-md);
  background: rgba(255, 251, 245, 0.7);
}

.auth-feature strong {
  display: block;
  margin-bottom: 6px;
  font-size: 16px;
}

.auth-card {
  align-self: center;
  padding: 28px;
}

.auth-card-head h2 {
  margin: 16px 0 8px;
  font-size: 30px;
}

.auth-pill {
  background: rgba(203, 124, 50, 0.12);
  color: #9a5c21;
}

.auth-submit {
  width: 100%;
  margin-top: 8px;
  border: 0;
  background: linear-gradient(135deg, var(--cd-primary) 0%, #2a8a67 100%);
}

.auth-footer {
  display: flex;
  gap: 8px;
  margin-top: 18px;
  font-size: 14px;
}

@media (max-width: 960px) {
  .auth-grid {
    grid-template-columns: 1fr;
  }

  .auth-copy {
    padding: 8px 4px;
  }
}
</style>
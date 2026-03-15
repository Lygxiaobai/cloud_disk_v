<template>
  <div class="auth-page page-shell">
    <section class="auth-grid">
      <article class="panel auth-card">
        <div class="auth-card-head">
          <span class="pill auth-pill">Create account</span>
          <h2>Register</h2>
          <p class="muted">Email codes are short-lived, so it is best to complete registration right away.</p>
        </div>

        <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
          <el-form-item label="Username">
            <el-input v-model.trim="form.name" placeholder="Choose a username" size="large" />
          </el-form-item>

          <el-form-item label="Email">
            <el-input v-model.trim="form.email" placeholder="name@example.com" size="large" />
          </el-form-item>

          <el-form-item label="Password">
            <el-input
              v-model.trim="form.password"
              placeholder="Set a password"
              show-password
              size="large"
              type="password"
            />
          </el-form-item>

          <el-form-item label="Email code">
            <div class="code-row">
              <el-input v-model.trim="form.code" placeholder="Enter the code from email" size="large" />
              <el-button :disabled="countdown > 0" :loading="sendingCode" size="large" @click="handleSendCode">
                {{ countdown > 0 ? `${countdown}s` : "Send code" }}
              </el-button>
            </div>
          </el-form-item>

          <el-button :loading="submitting" class="auth-submit" native-type="submit" size="large" type="primary">
            Register account
          </el-button>
        </el-form>

        <div class="auth-footer">
          <span class="muted">Already have an account?</span>
          <RouterLink to="/login">Back to sign in</RouterLink>
        </div>
      </article>

      <article class="auth-copy">
        <span class="pill">Register Flow</span>
        <h1 class="page-title">Verify the mailbox first, then attach uploads, folders, and shares to a personal workspace.</h1>
        <p class="page-subtitle">
          The backend already supports send-code, register, sign in, and refresh-token flows. The frontend just needs to
          connect the state and interaction cleanly.
        </p>

        <div class="auth-highlights">
          <div class="auth-feature">
            <strong>Email verification</strong>
            <span>The form maps directly to the existing backend registration flow.</span>
          </div>
          <div class="auth-feature">
            <strong>State recovery</strong>
            <span>Pinia keeps access and refresh tokens stable across page reloads.</span>
          </div>
          <div class="auth-feature">
            <strong>Room to grow</strong>
            <span>If you later add device sessions or remember-me logic, the structure is already in place.</span>
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
    ElMessage.warning("Please enter an email first.");
    return;
  }

  sendingCode.value = true;
  try {
    await authStore.sendRegisterCode(form.email);
    ElMessage.success("Verification code sent.");
    startCountdown();
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "Unable to send the verification code."));
  } finally {
    sendingCode.value = false;
  }
}

async function handleSubmit(): Promise<void> {
  if (!form.name || !form.email || !form.password || !form.code) {
    ElMessage.warning("Please complete every field.");
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
    ElMessage.success("Account created. You can sign in now.");
    await router.replace("/login");
  } catch (error) {
    ElMessage.error(getErrorMessage(error, "Registration failed."));
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

.auth-grid {
  display: grid;
  gap: 24px;
  width: min(1180px, 100%);
  grid-template-columns: minmax(380px, 460px) minmax(0, 1.1fr);
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
  background: rgba(31, 107, 79, 0.1);
  color: var(--cd-primary-strong);
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
  background: linear-gradient(135deg, var(--cd-accent) 0%, #e39b55 100%);
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
    order: -1;
    padding: 8px 4px;
  }

  .code-row {
    grid-template-columns: 1fr;
  }
}
</style>
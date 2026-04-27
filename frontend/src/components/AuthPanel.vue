<script setup lang="ts">
import { reactive, ref, watch } from 'vue'

const props = defineProps<{
  loading?: boolean
  mode?: 'login' | 'register'
}>()

const emit = defineEmits<{
  back: []
  submit: [payload: { mode: 'login' | 'register'; username: string; password: string }]
}>()

const mode = ref<'login' | 'register'>(props.mode ?? 'login')
const form = reactive({
  username: '',
  password: '',
})

watch(
  () => props.mode,
  (nextMode) => {
    if (nextMode) {
      mode.value = nextMode
    }
  },
)

watch(mode, () => {
  form.password = ''
})

function handleSubmit() {
  emit('submit', {
    mode: mode.value,
    username: form.username.trim(),
    password: form.password,
  })
}
</script>

<template>
  <div class="auth-shell">
    <el-card class="auth-card" shadow="never">
      <div class="auth-content">
        <div class="auth-header">
          <div class="auth-brand">
            <div class="auth-logo">⚖️</div>
            <button class="auth-title-button" type="button" @click="emit('back')">
              每日体重
            </button>
            <p class="auth-subtitle">记录每一天的健康变化</p>
          </div>
          <el-segmented
            v-model="mode"
            class="auth-segmented"
            :options="[
              { label: '登录', value: 'login' },
              { label: '注册', value: 'register' },
            ]"
          />
        </div>

        <el-form label-position="top" class="auth-form" @submit.prevent="handleSubmit">
          <el-form-item label="用户名">
            <el-input
              v-model="form.username"
              maxlength="32"
              placeholder="请输入用户名"
              size="large"
            />
          </el-form-item>

          <el-form-item label="密码">
            <el-input
              v-model="form.password"
              show-password
              type="password"
              size="large"
              placeholder="请输入密码"
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-button
            :loading="props.loading"
            class="auth-submit"
            type="primary"
            size="large"
            @click="handleSubmit"
          >
            {{ mode === 'login' ? '登录' : '注册账号' }}
          </el-button>

          <div class="auth-footer">
            <template v-if="mode === 'login'">
              <span class="footer-text">还没有账号？</span>
              <button type="button" class="footer-link" @click="mode = 'register'">
                立即注册
              </button>
            </template>
            <template v-else>
              <span class="footer-text">已有账号？</span>
              <button type="button" class="footer-link" @click="mode = 'login'">
                立即登录
              </button>
            </template>
          </div>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.auth-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 50%, #f0f9ff 100%);
  background-size: 200% 200%;
  animation: gradientBg 15s ease infinite;
}

@keyframes gradientBg {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.auth-card {
  width: 100%;
  max-width: 420px;
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15);
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  overflow: hidden;
  transition: all 0.3s ease;
}

.auth-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 30px 60px -12px rgba(0, 0, 0, 0.2);
}

.auth-content {
  padding: 32px;
}

.auth-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  margin-bottom: 32px;
  text-align: center;
}

.auth-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.auth-logo {
  font-size: 48px;
  margin-bottom: 8px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.auth-title-button {
  padding: 0;
  border: none;
  background: transparent;
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  cursor: pointer;
  transition: all 0.2s ease;
}

.auth-title-button:hover {
  opacity: 0.85;
  transform: scale(1.02);
}

.auth-subtitle {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

.auth-segmented {
  width: 100%;
  max-width: 200px;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.auth-submit {
  width: 100%;
  margin-top: 16px;
  height: 48px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 16px;
  letter-spacing: 0.5px;
  transition: all 0.2s ease;
}

.auth-submit:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.auth-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.footer-text {
  color: #94a3b8;
  font-size: 14px;
}

.footer-link {
  background: none;
  border: none;
  color: #3b82f6;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  padding: 0;
  transition: all 0.2s ease;
}

.footer-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

/* Element Plus 组件样式覆盖 */
:deep(.el-input__wrapper) {
  border-radius: 10px;
  padding: 4px 12px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #1e293b;
  padding-bottom: 6px;
}

:deep(.el-segmented__item) {
  border-radius: 8px;
}

:deep(.el-segmented__item-selected) {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

/* 响应式 */
@media (max-width: 480px) {
  .auth-content {
    padding: 24px;
  }

  .auth-title-button {
    font-size: 28px;
  }

  .auth-logo {
    font-size: 40px;
  }
}
</style>

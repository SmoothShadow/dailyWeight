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
      <div class="auth-header">
        <div>
          <button class="auth-title-button" type="button" @click="emit('back')">dailyWeight</button>
          <p>记录每天的体重变化</p>
        </div>
        <el-segmented
          v-model="mode"
          :options="[
            { label: '登录', value: 'login' },
            { label: '注册', value: 'register' },
          ]"
        />
      </div>

      <el-form label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" maxlength="32" placeholder="支持中文用户名" />
        </el-form-item>

        <el-form-item label="密码">
          <el-input
            v-model="form.password"
            show-password
            type="password"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-button :loading="props.loading" class="auth-submit" type="primary" @click="handleSubmit">
          {{ mode === 'login' ? '登录' : '注册' }}
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

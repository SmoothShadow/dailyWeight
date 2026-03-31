<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'
import AuthPanel from './components/AuthPanel.vue'
import WeightCalendar from './components/WeightCalendar.vue'
import { getCurrentUser, getWeights, login, logout, register, saveWeight } from './api'
import type { User, WeightRecord } from './types'

type ViewMode = 'calendar' | 'auth'

const loading = ref(false)
const saving = ref(false)
const initialized = ref(false)
const currentUser = ref<User | null>(null)
const records = ref<WeightRecord[]>([])
const currentMonth = ref('')
const authMode = ref<'login' | 'register'>('login')
const currentView = ref<ViewMode>('calendar')


onMounted(async () => {
  try {
    const response = await getCurrentUser()
    currentUser.value = response.user
  } catch {
    currentUser.value = null
  } finally {
    initialized.value = true
  }
})

async function handleAuth(payload: { mode: 'login' | 'register'; username: string; password: string }) {
  if (!payload.username || !payload.password) {
    ElMessage.error('用户名和密码不能为空')
    return
  }

  loading.value = true
  try {
    const response = payload.mode === 'login' ? await login(payload) : await register(payload)
    currentUser.value = response.user
    currentView.value = 'calendar'
    ElMessage.success(payload.mode === 'login' ? '登录成功' : '注册成功')
    if (currentMonth.value) {
      await loadWeights(currentMonth.value)
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '操作失败')
  } finally {
    loading.value = false
  }
}

async function handleLogout() {
  try {
    await logout()
    currentUser.value = null
    records.value = []
    currentView.value = 'calendar'
    ElMessage.success('已退出登录')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '退出失败')
  }
}

async function loadWeights(month: string) {
  currentMonth.value = month
  if (!currentUser.value) {
    records.value = []
    return
  }

  try {
    records.value = await getWeights(month)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载体重失败')
  }
}

async function handleSaveWeight(payload: { date: string; weight: number }) {
  if (!payload.date || Number.isNaN(payload.weight)) {
    ElMessage.error('请输入有效体重')
    return
  }

  saving.value = true
  try {
    const record = await saveWeight(payload)
    const nextRecords = records.value.filter(
      (item) => !(item.userId === record.userId && item.date === record.date),
    )
    nextRecords.push(record)
    records.value = nextRecords.sort((a, b) => {
      if (a.date === b.date) {
        return a.username.localeCompare(b.username, 'zh-CN')
      }
      return a.date.localeCompare(b.date)
    })
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    saving.value = false
  }
}

function handleShowAuth(mode: 'login' | 'register') {
  authMode.value = mode
  currentView.value = 'auth'
}

function handleBackToCalendar() {
  currentView.value = 'calendar'
}
</script>

<template>
  <div v-if="!initialized" class="app-loading">加载中...</div>

  <div v-else class="app-shell">
    <WeightCalendar
      v-if="currentView === 'calendar'"
      :loading="loading"
      :records="records"
      :saving="saving"
      :user="currentUser"
      @logout="handleLogout"
      @month-change="loadWeights"
      @save="handleSaveWeight"
      @show-auth="handleShowAuth"
    />

    <AuthPanel
      v-else
      :key="authMode"
      :loading="loading"
      :mode="authMode"
      @back="handleBackToCalendar"
      @submit="handleAuth"
    />
  </div>
</template>

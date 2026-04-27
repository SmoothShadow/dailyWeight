<script setup lang="ts">
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { computed, ref, watch } from 'vue'
import type { User, WeightRecord } from '../types'
import WeightRecordDialog from './WeightRecordDialog.vue'
import { changePassword } from '../api'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  user: User | null
  records: WeightRecord[]
  loading?: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  monthChange: [month: string]
  save: [payload: { date: string; weight: number }]
  logout: []
  showAuth: [mode: 'login' | 'register']
}>()

const selectedDate = ref(new Date())
const dialogVisible = ref(false)
const dialogDate = ref('')
const changePwdVisible = ref(false)
const changePwdForm = ref({ oldPassword: '', newPassword: '' })
const changing = ref(false)
const currentMonth = computed(() => formatMonth(selectedDate.value))
const currentYear = computed(() => `${selectedDate.value.getFullYear()}年`)
const currentMonthLabel = computed(() =>
  `${selectedDate.value.getMonth() + 1}月`,
)

// 星期几的中文显示
const weekDays = ['日', '一', '二', '三', '四', '五', '六']

watch(
  currentMonth,
  (month) => {
    emit('monthChange', month)
  },
  { immediate: true },
)

const recordsByDate = computed(() => {
  return props.records.reduce<Record<string, WeightRecord[]>>((acc, record) => {
    acc[record.date] ??= []
    acc[record.date].push(record)
    return acc
  }, {})
})

const activeRecord = computed(() => {
  if (!props.user || !dialogDate.value) {
    return null
  }

  return (recordsByDate.value[dialogDate.value] ?? []).find((record) => record.userId === props.user?.id) ?? null
})

function formatMonth(date: Date) {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  return `${year}-${month}`
}

function formatDate(date: Date) {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 格式化日期为中文显示
function formatDateChinese(dateStr: string) {
  const [year, month, day] = dateStr.split('-')
  return `${year}年${parseInt(month)}月${parseInt(day)}日`
}

// 获取星期几
function getWeekDay(dateStr: string) {
  const date = new Date(dateStr)
  return `周${weekDays[date.getDay()]}`
}

function toDate(value: string) {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function resetChangePwd() {
  changePwdForm.value = { oldPassword: '', newPassword: '' }
}

async function handleChangePasswordSubmit() {
  if (!changePwdForm.value.oldPassword || !changePwdForm.value.newPassword) {
    ElMessage.error('旧密码和新密码不能为空')
    return
  }
  if (changePwdForm.value.oldPassword === changePwdForm.value.newPassword) {
    ElMessage.error('新密码不能与旧密码相同')
    return
  }

  changing.value = true
  try {
    await changePassword({ ...changePwdForm.value })
    ElMessage.success('密码修改成功')
    changePwdVisible.value = false
    resetChangePwd()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改失败')
  } finally {
    changing.value = false
  }
}

function previousMonth() {
  selectedDate.value = new Date(selectedDate.value.getFullYear(), selectedDate.value.getMonth() - 1, 1)
}

function nextMonth() {
  selectedDate.value = new Date(selectedDate.value.getFullYear(), selectedDate.value.getMonth() + 1, 1)
}

function openDialog(date: Date) {
  dialogDate.value = formatDate(date)
  dialogVisible.value = true
}

function handleChangePasswordClick() {
  changePwdVisible.value = true
  resetChangePwd()
}

function handleCellClick(date: Date) {
  if (!props.user) {
    emit('showAuth', 'login')
    return
  }

  openDialog(date)
}

function handleRecordClick(date: string) {
  if (!props.user) {
    emit('showAuth', 'login')
    return
  }

  openDialog(toDate(date))
}

// 判断是否是今天
function isToday(dateStr: string) {
  const today = new Date()
  const [year, month, day] = dateStr.split('-').map(Number)
  return today.getFullYear() === year &&
         today.getMonth() === month - 1 &&
         today.getDate() === day
}
</script>

<template>
  <section class="calendar-page">
    <header class="calendar-toolbar">
      <div class="toolbar-left">
        <div class="toolbar-year">{{ currentYear }}</div>
        <div class="toolbar-month-label">{{ currentMonthLabel }}</div>
      </div>

      <div class="toolbar-month-switcher">
        <el-button circle @click="previousMonth">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="month-display">{{ currentMonthLabel }}</div>
        <el-button circle @click="nextMonth">
          <el-icon><ArrowRight /></el-icon>
        </el-button>
      </div>

      <div class="toolbar-actions">
        <el-dropdown v-if="props.user" trigger="click">
          <el-button class="user-btn">
            <span class="user-icon">👤</span>
            {{ props.user.username }}
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleChangePasswordClick">
                <span class="dropdown-icon">🔑</span>修改密码
              </el-dropdown-item>
              <el-dropdown-item divided @click="emit('logout')">
                <span class="dropdown-icon">🚪</span>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <div v-else class="guest-actions">
          <el-button @click="emit('showAuth', 'login')">登录</el-button>
          <el-button type="primary" @click="emit('showAuth', 'register')">注册</el-button>
        </div>
      </div>
    </header>

    <el-card class="calendar-card" shadow="never">
      <template #header>
        <div class="calendar-header">
          <span class="calendar-title">体重日历</span>
          <span class="calendar-subtitle">点击日期记录体重</span>
        </div>
      </template>
      <el-calendar v-model="selectedDate">
        <template #date-cell="{ data }">
          <div
            class="calendar-cell"
            :class="{ 'is-today': isToday(data.day) }"
            @click="handleCellClick(toDate(data.day))"
          >
            <div class="calendar-day-header">
              <span class="calendar-day-number" :class="{ 'today-number': isToday(data.day) }">
                {{ parseInt(data.day.split('-')[2]) }}
              </span>
              <span v-if="isToday(data.day)" class="today-label">今天</span>
            </div>
            <div class="calendar-records">
              <button
                v-for="record in recordsByDate[data.day] ?? []"
                :key="record.id"
                class="record-chip"
                type="button"
                @click.stop="handleRecordClick(data.day)"
              >
                <span class="record-weight">{{ record.weight }}</span>
                <span class="record-unit">kg</span>
              </button>
            </div>
          </div>
        </template>
      </el-calendar>
    </el-card>

    <WeightRecordDialog
      v-model="dialogVisible"
      :date="dialogDate"
      :initial-weight="activeRecord?.weight ?? null"
      :loading="props.saving"
      @save="emit('save', $event)"
    />

    <el-dialog
      v-model="changePwdVisible"
      title="修改密码"
      width="400px"
      class="change-pwd-dialog"
    >
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="旧密码">
          <el-input
            v-model="changePwdForm.oldPassword"
            type="password"
            autocomplete="current-password"
            placeholder="请输入旧密码"
          />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input
            v-model="changePwdForm.newPassword"
            type="password"
            autocomplete="new-password"
            placeholder="请输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="changePwdVisible = false">取消</el-button>
          <el-button type="primary" :loading="changing" @click="handleChangePasswordSubmit">
            确认修改
          </el-button>
        </div>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.calendar-toolbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 8px 4px;
}

.toolbar-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toolbar-year {
  font-size: 28px;
  font-weight: 800;
  background: linear-gradient(135deg, #1e293b 0%, #64748b 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.toolbar-month-label {
  font-size: 14px;
  color: #94a3b8;
  font-weight: 500;
}

.toolbar-month-switcher {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  justify-self: center;
}

.month-display {
  min-width: 60px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
}

.toolbar-actions {
  justify-self: end;
}

.guest-actions {
  display: flex;
  gap: 12px;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-icon {
  font-size: 16px;
}

.dropdown-icon {
  margin-right: 8px;
}

/* 日历头部 */
.calendar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.calendar-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.calendar-subtitle {
  font-size: 13px;
  color: #94a3b8;
}

/* 日历单元格 */
.calendar-cell {
  height: 100%;
  min-height: 100px;
  display: flex;
  flex-direction: column;
  padding: 8px;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.2s ease;
  background: transparent;
}

.calendar-cell:hover {
  background: #f0f9ff;
}

.calendar-cell.is-today {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.calendar-day-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.calendar-day-number {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  transition: all 0.2s ease;
}

.calendar-day-number.today-number {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.today-label {
  font-size: 11px;
  color: #3b82f6;
  font-weight: 500;
}

/* 体重记录标签 */
.calendar-records {
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow: hidden;
  margin-top: auto;
}

.record-chip {
  width: 100%;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  padding: 6px 10px;
  text-align: center;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.record-chip:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.record-weight {
  font-weight: 600;
}

.record-unit {
  font-size: 11px;
  opacity: 0.9;
}

/* 修改密码弹窗 */
.change-pwd-dialog :deep(.el-dialog__body) {
  padding-top: 16px;
}

/* 响应式 */
@media (max-width: 768px) {
  .calendar-toolbar {
    grid-template-columns: 1fr;
    gap: 12px;
    text-align: center;
  }

  .toolbar-left {
    align-items: center;
  }

  .toolbar-actions {
    justify-self: center;
  }

  .calendar-cell {
    min-height: 80px;
    padding: 6px;
  }

  .calendar-day-number {
    width: 24px;
    height: 24px;
    font-size: 12px;
  }
}
</style>

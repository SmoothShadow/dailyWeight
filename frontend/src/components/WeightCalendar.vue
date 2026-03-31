<script setup lang="ts">
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { computed, ref, watch } from 'vue'
import type { User, WeightRecord } from '../types'
import WeightRecordDialog from './WeightRecordDialog.vue'

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
const currentMonth = computed(() => formatMonth(selectedDate.value))
const currentYear = computed(() => selectedDate.value.getFullYear())
const currentMonthLabel = computed(() =>
  selectedDate.value.toLocaleDateString('zh-CN', { month: 'long' }),
)

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

function toDate(value: string) {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
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
</script>

<template>
  <section class="calendar-page">
    <header class="calendar-toolbar">
      <div class="toolbar-year">{{ currentYear }}</div>

      <div class="toolbar-month-switcher">
        <el-button circle @click="previousMonth">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <strong>{{ currentMonthLabel }}</strong>
        <el-button circle @click="nextMonth">
          <el-icon><ArrowRight /></el-icon>
        </el-button>
      </div>

      <div class="toolbar-actions">
        <el-dropdown v-if="props.user" trigger="click">
          <el-button>
            {{ props.user.username }}
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="emit('logout')">退出登录</el-dropdown-item>
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
      <el-calendar v-model="selectedDate">
        <template #date-cell="{ data }">
          <div class="calendar-cell" @click="handleCellClick(toDate(data.day))">
            <span class="calendar-day-number">{{ data.day.split('-')[2] }}</span>
            <div class="calendar-records">
              <button
                v-for="record in recordsByDate[data.day] ?? []"
                :key="record.id"
                class="record-chip"
                type="button"
                @click.stop="handleRecordClick(data.day)"
              >
                {{ record.username }}：{{ record.weight }}kg
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
  </section>
</template>

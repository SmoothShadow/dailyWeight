<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  date: string
  initialWeight?: number | null
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  save: [payload: { date: string; weight: number }]
}>()

const form = reactive({
  weight: props.initialWeight?.toString() ?? '',
})

watch(
  () => [props.initialWeight, props.modelValue, props.date],
  () => {
    form.weight = props.initialWeight != null ? props.initialWeight.toString() : ''
  },
)

const isEdit = computed(() => props.initialWeight != null)

// 格式化日期为中文显示
const formattedDate = computed(() => {
  if (!props.date) return ''
  const [year, month, day] = props.date.split('-')
  return `${year}年${parseInt(month)}月${parseInt(day)}日`
})

// 获取星期几
const weekDay = computed(() => {
  if (!props.date) return ''
  const weekDays = ['日', '一', '二', '三', '四', '五', '六']
  const date = new Date(props.date)
  return `周${weekDays[date.getDay()]}`
})

function handleSave() {
  const weight = Number(form.weight)
  emit('save', {
    date: props.date,
    weight,
  })
}
</script>

<template>
  <el-dialog
    :model-value="props.modelValue"
    width="400px"
    class="weight-dialog"
    :show-close="false"
    @close="emit('update:modelValue', false)"
  >
    <template #header>
      <div class="dialog-header">
        <div class="dialog-icon">{{ isEdit ? '✏️' : '⚖️' }}</div>
        <h3 class="dialog-title">{{ isEdit ? '编辑体重' : '记录体重' }}</h3>
        <p class="dialog-date">{{ formattedDate }} {{ weekDay }}</p>
      </div>
    </template>

    <el-form label-position="top" class="weight-form" @submit.prevent="handleSave">
      <el-form-item class="weight-input-item">
        <div class="weight-input-wrapper">
          <input
            v-model="form.weight"
            type="number"
            step="0.1"
            class="weight-input"
            placeholder="0.0"
            @keyup.enter="handleSave"
          />
          <span class="weight-unit">kg</span>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button class="cancel-btn" @click="emit('update:modelValue', false)">
          取消
        </el-button>
        <el-button
          type="primary"
          class="save-btn"
          :loading="props.loading"
          @click="handleSave"
        >
          {{ isEdit ? '保存修改' : '记录体重' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.weight-dialog :deep(.el-dialog) {
  border-radius: 20px;
  overflow: hidden;
}

.weight-dialog :deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
}

.weight-dialog :deep(.el-dialog__body) {
  padding: 0 24px;
}

.weight-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px 24px;
}

.dialog-header {
  padding: 24px 24px 20px;
  text-align: center;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
}

.dialog-icon {
  font-size: 40px;
  margin-bottom: 12px;
  animation: bounce 2s ease infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-6px); }
}

.dialog-title {
  margin: 0 0 4px;
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

.dialog-date {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.weight-form {
  padding-top: 20px;
}

.weight-input-item {
  margin-bottom: 0;
}

.weight-input-wrapper {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 8px;
  padding: 16px 0;
}

.weight-input {
  width: 160px;
  font-size: 48px;
  font-weight: 700;
  text-align: center;
  border: none;
  background: transparent;
  color: #1e293b;
  outline: none;
  caret-color: #3b82f6;
}

.weight-input::placeholder {
  color: #cbd5e1;
  font-weight: 400;
}

.weight-input:focus {
  color: #3b82f6;
}

.weight-unit {
  font-size: 20px;
  font-weight: 500;
  color: #94a3b8;
}

/* 隐藏数字输入框的上下箭头 */
.weight-input::-webkit-outer-spin-button,
.weight-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.weight-input[type=number] {
  -moz-appearance: textfield;
}

.dialog-footer {
  display: flex;
  gap: 12px;
}

.cancel-btn {
  flex: 1;
  height: 44px;
  border-radius: 12px;
  font-weight: 500;
}

.save-btn {
  flex: 2;
  height: 44px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 15px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
}

.save-btn:hover {
  background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.3);
}

/* 响应式 */
@media (max-width: 480px) {
  .weight-dialog :deep(.el-dialog) {
    width: 90% !important;
    margin: 20px auto;
  }

  .weight-input {
    font-size: 36px;
    width: 120px;
  }

  .weight-unit {
    font-size: 16px;
  }
}
</style>

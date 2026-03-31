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
    :title="isEdit ? '编辑体重' : '记录体重'"
    width="420px"
    @close="emit('update:modelValue', false)"
  >
    <el-form label-position="top" @submit.prevent="handleSave">
      <el-form-item label="日期">
        <el-input :model-value="props.date" disabled />
      </el-form-item>

      <el-form-item label="当日体重">
        <el-input v-model="form.weight" placeholder="请输入体重">
          <template #append>kg</template>
        </el-input>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button :loading="props.loading" type="primary" @click="handleSave">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

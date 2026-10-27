<script setup lang="ts">
import { ref } from 'vue'
import { api, usePolling, clone } from '../lib/api'
import { t } from '../lib/i18n'
import type { User } from '../lib/models'
const { data: members, reload, error } = usePolling<User[]>(() => '/users', [])
const open = ref(false),
  form = ref({ id: '', username: '', password: '', role: 'viewer', disabled: false }),
  saveError = ref('')
function edit(u?: User) {
  form.value = {
    id: '',
    username: '',
    role: 'viewer',
    disabled: false,
    ...(u ? clone(u) : {}),
    password: '',
  }
  saveError.value = ''
  open.value = true
}
async function save() {
  try {
    await api(
      '/users' + (form.value.id ? '/' + form.value.id : ''),
      form.value.id ? 'PUT' : 'POST',
      form.value,
    )
    open.value = false
    await reload()
  } catch (e) {
    saveError.value = String(e)
  }
}
</script>
<template>
  <section class="panel">
    <div class="section-title">
      <h2>{{ t('members') }}</h2>
      <el-button @click="edit()">+ {{ t('add') }}</el-button>
    </div>
    <el-alert v-if="error" :title="error" type="error" /><el-table :data="members"
      ><el-table-column prop="username" :label="t('username')" /><el-table-column :label="t('role')"
        ><template #default="{ row }">{{ t(row.role) }}</template></el-table-column
      ><el-table-column :label="t('enabled')"
        ><template #default="{ row }">{{ row.disabled ? '—' : '✓' }}</template></el-table-column
      ><el-table-column
        ><template #default="{ row }"
          ><el-button text @click="edit(row)">{{ t('edit') }}</el-button></template
        ></el-table-column
      ></el-table
    >
  </section>
  <el-dialog v-model="open" :title="t('members')" width="480px"
    ><el-form label-position="top" @submit.prevent="save"
      ><el-form-item :label="t('username')"><el-input v-model="form.username" /></el-form-item
      ><el-form-item :label="t('password')"
        ><el-input
          v-model="form.password"
          type="password"
          autocomplete="new-password"
          show-password /></el-form-item
      ><el-form-item :label="t('role')"
        ><el-select v-model="form.role"
          ><el-option
            v-for="role in ['viewer', 'operator', 'admin']"
            :key="role"
            :value="role"
            :label="t(role)" /></el-select></el-form-item
      ><el-checkbox v-model="form.disabled">{{ t('paused') }}</el-checkbox
      ><el-alert v-if="saveError" :title="saveError" type="error" />
      <div class="actions">
        <el-button type="primary" native-type="submit">{{ t('save') }}</el-button>
      </div></el-form
    ></el-dialog
  >
</template>

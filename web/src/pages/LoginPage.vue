<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NCard, NForm, NFormItem, NInput, NTabPane, NTabs, useMessage } from 'naive-ui'
import { api, setToken } from '../api'

const router = useRouter()
const route = useRoute()
const message = useMessage()

const tab = ref('login')
const loading = ref(false)
const form = reactive({ username: '', password: '', nickname: '' })

async function submit() {
  const username = form.username.trim()
  if (!username) return message.warning('请输入用户名')
  if (!form.password) return message.warning('请输入密码')
  loading.value = true
  try {
    const data =
      tab.value === 'login'
        ? await api.login({ username, password: form.password })
        : await api.register({
            username,
            password: form.password,
            nickname: form.nickname.trim() || username,
          })
    setToken(data.token)
    message.success(tab.value === 'login' ? '登录成功' : '注册成功，已自动登录')
    router.replace((route.query.redirect || '/surveys').toString())
  } catch (e) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <n-card class="login-card">
      <div class="brand">
        <div class="brand-logo">问</div>
        <div>
          <div class="brand-title">在线调研问卷系统</div>
          <div class="muted">数据结构课程设计 · Vue3 + Kratos + SQLite</div>
        </div>
      </div>

      <n-tabs v-model:value="tab" type="line" animated>
        <n-tab-pane name="login" tab="登录">
          <n-form label-placement="top" :show-feedback="false">
            <n-form-item label="用户名">
              <n-input v-model:value="form.username" placeholder="admin" @keyup.enter="submit" />
            </n-form-item>
            <n-form-item label="密码" style="margin-top: 12px">
              <n-input
                v-model:value="form.password"
                type="password"
                show-password-on="click"
                placeholder="admin123"
                @keyup.enter="submit"
              />
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <n-tab-pane name="register" tab="注册">
          <n-form label-placement="top" :show-feedback="false">
            <n-form-item label="用户名">
              <n-input v-model:value="form.username" placeholder="3-32 位字母数字" />
            </n-form-item>
            <n-form-item label="昵称" style="margin-top: 12px">
              <n-input v-model:value="form.nickname" placeholder="不填就用用户名" />
            </n-form-item>
            <n-form-item label="密码" style="margin-top: 12px">
              <n-input
                v-model:value="form.password"
                type="password"
                show-password-on="click"
                placeholder="至少 6 位"
                @keyup.enter="submit"
              />
            </n-form-item>
          </n-form>
        </n-tab-pane>
      </n-tabs>

      <n-button type="primary" block size="large" :loading="loading" style="margin-top: 20px" @click="submit">
        {{ tab === 'login' ? '登录' : '注册并登录' }}
      </n-button>

      <div class="muted hint">演示账号：admin / admin123</div>
    </n-card>
  </div>
</template>

<style scoped>
.login-wrap {
  min-height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
}
.login-card {
  width: 100%;
  max-width: 420px;
  border-radius: 18px;
  box-shadow: 0 18px 48px rgba(31, 39, 51, 0.08);
}
.brand {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}
.brand-logo {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.brand-title {
  font-size: 17px;
  font-weight: 700;
}
.hint {
  text-align: center;
  margin-top: 14px;
}
</style>

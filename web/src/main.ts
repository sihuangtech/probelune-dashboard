import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import './styles/base.css'
import './styles/layout.css'
import './styles/responsive.css'

import App from './App.vue'
import { i18n } from './lib/i18n'
import router from './router'

const app = createApp(App)

app.use(i18n)
app.use(ElementPlus)
app.use(router)

app.mount('#app')

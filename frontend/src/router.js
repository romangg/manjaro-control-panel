import { createMemoryHistory, createRouter } from 'vue-router'

import HomeView from './components/HomeView.vue'
import KernelsView from './components/KernelsView.vue'
import HardwareView from './components/HardwareView.vue'
import LanguageView from './components/LanguageView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/kernels', component: KernelsView },
  { path: '/hardware', component: HardwareView },
  { path: '/language', component: LanguageView },
]

const router = createRouter({
  history: createMemoryHistory(),
  routes,
})

export default router

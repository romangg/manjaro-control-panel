import { createMemoryHistory, createRouter } from 'vue-router'

import HomeView from './components/HomeView.vue'
import KernelsView from './components/KernelsView.vue'
import HardwareView from './components/HardwareView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/kernels', component: KernelsView },
  { path: '/hardware', component: HardwareView },
]

const router = createRouter({
  history: createMemoryHistory(),
  routes,
})

export default router

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goToKernels = () => {
  router.push('/kernels')
}
const goToHardware = () => {
  router.push('/hardware')
}
</script>

<template>
  <button @click="goToKernels">Kernels</button>
  <button @click="goToHardware">Hardware</button>
</template>

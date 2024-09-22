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
  <div class="h-screen flex items-center justify-center">
    <div>
      <h1 class="text-3xl pb-8 text-center">Control Panel</h1>
      <div class="flex gap-4">
        <Button label="Kernels" @click="goToKernels" />
        <Button label="Hardware" @click="goToHardware" />
      </div>
    </div>
  </div>
</template>

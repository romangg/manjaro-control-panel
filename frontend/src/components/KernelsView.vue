<script setup>
import { ref, onMounted } from 'vue'
import { KernelService } from "../../bindings/manjaro-control-panel";
import { Events } from "@wailsio/runtime";

const kernels = ref([])
const backendOpActive = ref('')

console.log("XXX SCRIPT SETUP")

const getKernels = () => {
  KernelService.Kernels().then((value) => {
    kernels.value = value;
  }).catch((err) => {
    console.log(err);
  });
}

const doInstall = (name) => {
  backendOpActive.value = name
  KernelService.Install(name)
}

const doRemove = (name) => {
  backendOpActive.value = name
  KernelService.Remove(name)
}

onMounted(() => {
  Events.On("kernelOpFinished", function (event) {
    backendOpActive.value = ''
    getKernels()
  })
  getKernels()
})
</script>

<template>
  <h1  style="activeClass">Kernels</h1>

  <div :style="{ activeClass: !backendOpActive}">
    <div v-for="kernel in kernels"  style="activeClass">
      {{ kernel.Name }} - {{ kernel.Version }} - {{ kernel.Lts }} <button v-if="!kernel.Installed"
        @click="doInstall(kernel.Name)">Install</button><button v-if="kernel.Installed"
        @click="doRemove(kernel.Name)">Remove</button>
        <div v-if="backendOpActive == kernel.Name">IN PROGRESS</div>
    </div>
  </div>
</template>

<style scoped>
.activeClass {
  background-color: green;
  color: green;
}
</style>

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

const items = [
  { label: 'Control Panel', route: '/' },
  { label: 'Kernels' },
];

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
  <Breadcrumb :home="home" :model="items">
    <template #item="{ item, props }">
      <router-link v-if="item.route" v-slot="{ href, navigate }" :to="item.route" custom>
        <a :href="href" v-bind="props.action" @click="navigate">
          <span>{{ item.label }}</span>
        </a>
      </router-link>
      <a v-else :href="item.url" :target="item.target" v-bind="props.action">
        <span>{{ item.label }}</span>
      </a>
    </template>
  </Breadcrumb>
  <DataView :value="kernels">
    <template #list="slotProps">
      <div class="flex flex-col">
        <div v-for="(kernel, index) in slotProps.items" :key="index">
          <div class="flex flex-col p-6 gap-4"
            :class="{ 'border-t border-surface-200 dark:border-surface-700': index !== 0 }">
            <div class="grid grid-cols-3 gap-4 items-center">
              <div>
                <h3 class="font-semibold text-2xl">
                  Linux {{ kernel.Version }}
                </h3>
                {{ kernel.Name }}
              </div>
              <div class="justify-self-center">
                <div class="flex flex-wrap justify-center gap-3">
                  <Message v-if="kernel.Recommended" severity="help">Recommended</Message>
                  <Message v-if="kernel.Lts" severity="info">LTS</Message>
                  <Message v-if="kernel.RealTime" severity="info">Real-time</Message>
                </div>
              </div>
              <div class="justify-self-center">
                <Button v-if="kernel.Installed" severity="danger" @click="doRemove(kernel.Name)">Remove</Button>
                <Button v-else @click="doInstall(kernel.Name)">Install</Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </DataView>
</template>

<style scoped>
.activeClass {
  background-color: green;
  color: green;
}
</style>

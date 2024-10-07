<script setup>
import { ref, useTemplateRef, onMounted } from 'vue'
import { KernelService } from "../../bindings/manjaro-control-panel";
import { Events } from "@wailsio/runtime";

const kernels = ref([])
const backendOpActive = ref('')
const backendOpLogHeader = ref('')
const backendOpLog = ref([])
const opLogShown = ref(false);
const opLogScroll = useTemplateRef('op-log-scroll')
const opLogScrollBottom = useTemplateRef('op-log-scroll-bottom')

const getKernels = () => {
  KernelService.Kernels().then((value) => {
    kernels.value = value;
    backendOpActive.value = ''
  }).catch((err) => {
    console.log(err);
    backendOpActive.value = ''
  });
}

const items = [
  { label: 'Control Panel', route: '/' },
  { label: 'Kernels' },
];

const showOpLog = () => {
  opLogShown.value = true
  setTimeout(() => {
    opLogScrollBottom.value.scrollIntoView({ behaviour: 'smooth', block: 'end' });
  });
}

const resetBackendOpLog = () => {
  backendOpLog.value = []
  backendOpLogHeader.value = ''
}

const doInstall = (name) => {
  backendOpActive.value = name
  resetBackendOpLog()
  KernelService.Install(name)
}

const doRemove = (name) => {
  backendOpActive.value = name
  resetBackendOpLog()
  KernelService.Remove(name)
}

onMounted(() => {
  Events.On("kernelOpFinished", function (event) {
    getKernels()
  })
  Events.On("kernelOpOutputLine", function (event) {
    const line = event.data[0]
    if (backendOpLog.value[backendOpLog.value.length - 1] == line) {
      return
    }

    if (opLogScroll.value) {
      let content = opLogScroll.value.$refs.content
      const height = content.scrollHeight
      const scroll = content.scrollTop
      const client = content.clientHeight

      setTimeout(() => {
        if (opLogScrollBottom.value) {
          if (Math.abs(height - scroll - client) < client * 0.15) {
            opLogScrollBottom.value.scrollIntoView({ behaviour: 'smooth', block: 'end' });
          }
        }
      });
    }
    backendOpLog.value.push(line)

    if (line.startsWith("::")
      && !line.startsWith(":: Proceed with installation?")
      && !line.startsWith(":: Do you want to remove these packages?")) {
      backendOpLogHeader.value = line.substring(2)
    }
  })

  backendOpActive.value = 'init-kernels'
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
                <button v-if="backendOpActive == kernel.Name" @click="showOpLog()" class="flex flex-wrap justify-center gap-3">
                  {{ backendOpLogHeader }}
                </button>
                <div v-else class="flex flex-wrap justify-center gap-3">
                  <Message v-if="kernel.Running" severity="contrast">Running</Message>
                  <Message v-if="kernel.Recommended" severity="help">Recommended</Message>
                  <Message v-if="kernel.Lts" severity="info">LTS</Message>
                  <Message v-if="kernel.RealTime" severity="info">Real-time</Message>
                  <Message v-if="kernel.Eol" severity="error">Unsupported</Message>
                </div>
              </div>
              <div class="justify-self-center">
                <Button v-if="kernel.Installed && backendOpActive != kernel.Name" severity="danger"
                  @click="doRemove(kernel.Name)" :disabled="kernel.Running || backendOpActive != ''">Remove</Button>
                <Button v-if="!kernel.Installed && backendOpActive != kernel.Name" @click="doInstall(kernel.Name)"
                  :disabled="backendOpActive != ''">Install</Button>
                <button>
                  <ProgressSpinner v-if="backendOpActive == kernel.Name" style="width: 50px; height: 50px"
                    strokeWidth="8" @click="showOpLog()" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </DataView>
  <Drawer class="!w-full md:!w-[768px]" v-model:visible="opLogShown" header="Operation Log">
    <ScrollPanel ref="op-log-scroll" style="width: 100%; height: 100%">
      <p v-for="line in backendOpLog">
        {{ line }}
      </p>
      <span ref="op-log-scroll-bottom" style="height: 1px;"></span>
    </ScrollPanel>
  </Drawer>
</template>

<style scoped>
.activeClass {
  background-color: green;
  color: green;
}
</style>

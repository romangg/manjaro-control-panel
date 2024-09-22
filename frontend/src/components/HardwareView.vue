<template>
  <div>
    <Breadcrumb :model="items">
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
    <div class="flex flex-wrap gap-4 py-4 px-12 justify-center">
      <Button class="" v-if="hasFreeInstall()" severity="secondary" @click="autoInstallFree()"
        :disabled="backendOpActive != ''">Auto-install open-source graphics
        driver</Button>
      <Button v-if="hasNonfreeInstall()" severity="secondary" @click="autoInstallProprietary()" :disabled="backendOpActive != ''">Auto-install
        proprietary
        graphics driver</Button>
    </div>
    <div
      v-for="(dev, index) in devices">
      <div class="flex flex-col bg-slate-50">
        <div class="flex flex-col p-6 gap-4"
          :class="{ 'border-t border-surface-200 dark:border-surface-700': index !== 0 }">
          <div>
            <h3 class="font-semibold text-2xl pb-4">
              {{ getDeviceName(dev) }} - {{ dev.Vendor_name }}
            </h3>
            <div class="p-3 grid grid-cols-3" :class="{ 'bg-slate-100' : cfgindex % 2 == 0 }" v-for="(cfg, cfgindex) in dev.Available_configs">
              <div class="text-lg">
                {{ cfg.Name }}
              </div>
              <div class="justify-self-center">
                <div class="flex flex-wrap justify-center gap-3">
                  <Message v-if="cfg.Freedriver" severity="info">Open-source</Message>
                </div>
              </div>
              <div class="justify-self-center">
                <Button v-if="isInstalled(dev, cfg.Name)" severity="danger" @click="doRemove(cfg.Name)"
                  :disabled="backendOpActive != ''">Remove</Button>
                <Button v-if="!isInstalled(dev, cfg.Name)" @click="doInstall(cfg.Name)"
                  :disabled="backendOpActive != ''">Install</Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { HwService } from "../../bindings/manjaro-control-panel";
import { Events } from "@wailsio/runtime";

const items = [
  { label: 'Control Panel', route: '/' },
  { label: 'Hardware' },
];

const devices = ref([])
const backendOpActive = ref('')

const getDevices = () => {
  HwService.Devices().then((value) => {
    devices.value = value.filter(dev => dev.Class_name == 'Display controller' && dev.Installed_configs.length + dev.Available_configs.length > 0);
  }).catch((err) => {
    console.log(err);
  });
}
getDevices()

const getDeviceName = (dev) => {
  if (dev.Device_name) {
    return dev.Device_name
  }
  if (dev.Subvendor_name) {
    return dev.Subvendor_name
  }
  if (dev.Model) {
    return dev.Model
  }
  return "Unknown"
}

const hasFreeInstall = () => {
  return devices.value.some(dev => dev.Available_configs.some(cfg => cfg.Freedriver))
}

const hasNonfreeInstall = () => {
  return devices.value.some(dev => dev.Available_configs.some(cfg => !cfg.Freedriver))
}

const isInstalled = (dev, cfg_name) => {
  return dev.Installed_configs.some(item => item.Name === cfg_name)
}

const doInstall = (name) => {
  backendOpActive.value = name

  HwService.InstallConfig(name).then((value) => {
    backendOpActive.value = ''
  }).catch((err) => {
    console.log(err);
    backendOpActive.value = ''
  });
}

const doRemove = (name) => {
  backendOpActive.value = name

  HwService.RemoveConfig(name).then((value) => {
    backendOpActive.value = ''
  }).catch((err) => {
    console.log(err);
    backendOpActive.value = ''
  });
}

const autoInstallFree = () => {
  backendOpActive.value = 'free'

  HwService.InstallFreeGpuConfig().then((value) => {
    backendOpActive.value = ''
  }).catch((err) => {
    console.log(err);
    backendOpActive.value = ''
  });
}

const autoInstallProprietary = () => {
  backendOpActive.value = 'nonfree'

  HwService.InstallProprietaryGpuConfig().then((value) => {
    backendOpActive.value = ''
  }).catch((err) => {
    console.log(err);
    backendOpActive.value = ''
  });
}

onMounted(() => {
  Events.On("hwOpFinished", function (event) {
    backendOpActive.value = ''
    getDevices()
  })
})
</script>

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
    <div
      v-for="(pkg, index) in languagePackages">
      <div class="flex flex-col bg-slate-50">
        <div class="flex flex-col p-6 gap-4"
          :class="{ 'border-t border-surface-200 dark:border-surface-700': index !== 0 }">
          <div>
            <h3 class="font-semibold text-2xl pb-4">
              {{ index }} {{ pkg.Name }}
            </h3>
            <!-- <div class="p-3 grid grid-cols-3" :class="{ 'bg-slate-100' : cfgindex % 2 == 0 }" v-for="(cfg, cfgindex) in dev.Available_configs">
              <div class="text-lg">
                {{ cfg.Name }}
              </div>
            </div> -->
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { LanguageService } from "../../bindings/manjaro-control-panel";

const items = [
  { label: 'Control Panel', route: '/' },
  { label: 'Language Packs' },
];

const languagePackages = ref([])

const getLanguagePackages = () => {
  LanguageService.Packages().then((value) => {
    languagePackages.value = value;
  }).catch((err) => {
    console.log(err);
  });
}
getLanguagePackages()
</script>

<template>
  <h3>Hardware</h3>
  <div v-for="dev in devices">
    <div v-if="dev.Installed_configs.length + dev.Available_configs.length > 0">
      {{ dev.Vendor_name }} - {{ dev.Class_name }} <br>Avail. cfgs: {{ dev.Available_configs }}<br> Installed. cfgs: {{ dev.Installed_configs }}
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { HwService } from "../../bindings/manjaro-control-panel";

const devices = ref([])

console.log("XXX HW SCRIPT SETUP")

const getDevices = () => {
  HwService.Devices().then((value) => {
    devices.value = value;
  }).catch((err) => {
    console.log(err);
  });
}
getDevices()
</script>

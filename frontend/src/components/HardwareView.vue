<template>
  <h3>Hardware</h3>
  <div v-for="dev in devices">
    {{ dev }}
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

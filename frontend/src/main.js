import { createApp } from "vue";
import router from "./router";
import App from "./App.vue";
import "./index.css";

import PrimeVue from "primevue/config";
import Aura from "@primevue/themes/aura";
import Button from "primevue/button"
import DataView from 'primevue/dataview';
import Drawer from 'primevue/drawer';
import Message from 'primevue/message';
import ScrollPanel from 'primevue/scrollpanel';
import Breadcrumb from 'primevue/breadcrumb';
import ProgressSpinner from 'primevue/progressspinner';

const app = createApp(App);
app.use(router);
app.use(PrimeVue, {
  theme: {
    preset: Aura,
  },
});
app.mount("#app");

app.component('Button', Button);
app.component('DataView', DataView);
app.component('Drawer', Drawer);
app.component('Message', Message);
app.component('ScrollPanel', ScrollPanel);
app.component('Breadcrumb', Breadcrumb);
app.component('ProgressSpinner', ProgressSpinner);

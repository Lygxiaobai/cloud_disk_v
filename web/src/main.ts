import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import { createApp } from "vue";

import App from "./App.vue";
import { router } from "./router";
import { useAuthStore } from "./stores/auth";
import { pinia } from "./stores/index";
import "./style.css";

const app = createApp(App);

app.use(pinia);
app.use(router);
app.use(ElementPlus);

const authStore = useAuthStore(pinia);
void authStore.bootstrap();

app.mount("#app");

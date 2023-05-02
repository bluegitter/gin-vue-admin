<template>
  <el-dialog
    v-model="dialogVisible"
    :before-close="beforeClose"
    @opened="doOpened"
    @close="doClose"
    :destroy-on-close="true"
  >
    <div ref="terminalContainer" class="terminal-container"></div>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, watch, defineExpose } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { useUserStore } from "@/pinia/modules/user";

const dialogVisible = ref(false);
const terminalContainer = ref(null);
let terminal = null;
let fitAddon = null;
let websocket = null;
let containerID = null;

const doOpen = (cid) => {
  console.log("doOpen: " + cid);
  dialogVisible.value = true;
  containerID = cid;
};

const doOpened = (cid) => {
  initTerminal();
};

const doClose = () => {
  dialogVisible.value = false;
  cleanupTerminal();
};

const beforeClose = (done) => {
  close();
  done();
};

const cleanupTerminal = () => {
  if (websocket && websocket.readyState === WebSocket.OPEN) {
    websocket.close();
  }
  if (terminal) {
    terminal.dispose();
    terminal = null;
  }
  if (terminalContainer.value) {
    terminalContainer.value.innerHTML = "";
  }
};

const initTerminal = () => {
  if (!terminalContainer.value) {
    console.error("Terminal container element not found");
    return;
  }
  const websocketURL = `ws://localhost:8888/docker/containers/${containerID}/console`;

  terminal = new Terminal();
  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.open(terminalContainer.value);
  fitAddon.fit();

  const userStore = useUserStore();
  const headers = {
    "x-token": userStore.token,
    "x-user-id": userStore.userInfo.ID,
  };

  debugger;
  websocket = new WebSocket(websocketURL, [], {
    headers: headers,
  });
  websocket.onopen = () => {
    console.log("  websocket.onopen, 连接成功");
    websocket.send("连接成功");
  };
  websocket.onmessage = (event) => {
    console.log("  websocket.onmessage, terminal.write");
    terminal.write(event.data);
  };
  websocket.onclose = () => {
    console.log("  websocket.onclose, terminal.write");
    terminal.write("连接已关闭");
  };

  terminal.onData((data) => {
    console.log("  terminal.onData, websocket.send");
    websocket.send(data);
  });
};

defineExpose({ doOpen });

// watch(dialogVisible, (value) => {
//   if (value) {
//     initTerminal();
//   } else {
//     cleanupTerminal();
//   }
// });
</script>

<style scoped>
.terminal-container {
  height: 100%;
  width: 100%;
}
</style>

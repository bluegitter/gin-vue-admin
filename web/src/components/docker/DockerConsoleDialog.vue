<template>
  <el-dialog
    v-model="dialogVisible"
    :before-close="beforeClose"
    fullscreen
    @opened="doOpened"
    @close="doClose"
    :destroy-on-close="true"
  >
    <div ref="terminalContainer" class="terminal-container"></div>
  </el-dialog>
</template>

<script>
// import "xterm/lib/addons/fullscreen/fullscreen.css";
// import "xterm/dist/xterm.css";
import "xterm/css/xterm.css";
</script>

<script setup>
import { reactive, ref, watch, defineExpose } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { useUserStore } from "@/pinia/modules/user";

const defaultTheme = {
  foreground: "#ffffff", // 字体
  background: "#1b212f", // 背景色
  cursor: "#ffffff", // 设置光标
  selection: "rgba(255, 255, 255, 0.3)",
  black: "#000000",
  brightBlack: "#808080",
  red: "#ce2f2b",
  brightRed: "#f44a47",
  green: "#00b976",
  brightGreen: "#05d289",
  yellow: "#e0d500",
  brightYellow: "#f4f628",
  magenta: "#bd37bc",
  brightMagenta: "#d86cd8",
  blue: "#1d6fca",
  brightBlue: "#358bed",
  cyan: "#00a8cf",
  brightCyan: "#19b8dd",
  white: "#e5e5e5",
  brightWhite: "#ffffff",
};

const dialogVisible = ref(false);
const terminalContainer = ref(null);
let terminal = null;
let websocket = null;
let containerID = null;
let fitAddon = null;

const doOpen = (cid) => {
  console.log("doOpen: " + cid);
  dialogVisible.value = true;
  containerID = cid;
};

const doOpened = (cid) => {
  initTerminal();
  window.addEventListener("resize", onWindowResize);
};

const onWindowResize = () => {
  fitAddon.fit();
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
  const hostname = location.hostname;
  const websocketURL = `ws://${hostname}:8888/docker/containers/${containerID}/console`;

  terminal = new Terminal({
    rows: 30,
    rendererType: "canvas", // 渲染类型
    convertEol: true, // 启用时，光标将设置为下一行的开头
    scrollback: 10, // 终端中的回滚量
    disableStdin: false, // 是否应禁用输入
    fontSize: 18,
    cursorBlink: true, // 光标闪烁
    cursorStyle: "bar", // 光标样式 underline
    bellStyle: "sound",
    theme: defaultTheme,
  });
  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.open(terminalContainer.value);
  fitAddon.fit();
  terminal.focus();

  const userStore = useUserStore();
  const headers = {
    "x-token": userStore.token,
    "x-user-id": userStore.userInfo.ID,
  };

  websocket = new WebSocket(websocketURL, [], {
    headers: headers,
  });
  websocket.onopen = () => {
    ws.send("uname -a\n");
  };
  websocket.onmessage = (event) => {
    terminal.write(event.data);
  };
  websocket.onclose = () => {
    terminal.write("Disconnected from server\n");
  };

  terminal.onData((data) => {
    websocket.send(data);
  });
};

defineExpose({ doOpen });
</script>

<style scoped>
.terminal-container {
  height: 100%;
  width: 100%;
}

.xterm div {
  font-family: Monaco, Menlo, Consolas, "Courier New", monospace;
  font-size: 16px;
}
</style>

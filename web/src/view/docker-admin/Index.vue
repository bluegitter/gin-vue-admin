<template>
  <div>
    <el-row>
      <el-col :span="24" class="toolbar">
        <el-button type="primary" icon="el-icon-plus" @click="newContainer"
          >New</el-button
        >
      </el-col>
    </el-row>
    <el-row>
      <el-col :span="24">
        <el-table :data="containers.value">
          <el-table-column
            prop="Id"
            label="ID"
            width="120"
            :formatter="truncateId"
          ></el-table-column>
          <el-table-column
            prop="Ports"
            label="Ports"
            :formatter="formatPorts"
          ></el-table-column>
          <el-table-column
            prop="CPUUsage"
            label="CPU Usage"
            width="100"
            :formatter="formatCpuUsage"
          ></el-table-column>
          <el-table-column
            prop="MemoryUsage"
            label="Memory Usage"
            width="150"
            :formatter="formatMemoryUsage"
          ></el-table-column>
          <el-table-column prop="State" label="State" width="100">
            <template v-slot="{ row }">
              <span :class="stateColorClass(row.State)">
                {{ row.State }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="Status"
            label="Status"
            width="200"
          ></el-table-column>
          <el-table-column label="Actions" width="260">
            <template v-slot="{ row }">
              <el-button
                type="primary"
                size="small"
                link
                @click="handleStartContainer(row.Id)"
                >Start</el-button
              >
              <el-button
                type="warning"
                size="small"
                link
                @click="handleStopContainer(row.Id)"
                >Stop</el-button
              >
              <el-button
                type="danger"
                size="small"
                link
                @click="handleRemoveContainer(row.Id)"
                >Remove</el-button
              >
              <el-button
                type="info"
                size="small"
                link
                @click="handleOpenConsole(row.Id)"
                >Console</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </el-col>
    </el-row>
    <docker-console-dialog
      ref="consoleDialog"
      :visible.sync="dialogVisible"
      :container-id="selectedContainerID"
    ></docker-console-dialog>
  </div>
</template>
<script>
export default {
  components: {
    DockerConsoleDialog,
  },
};
</script>
<script setup>
import {
  getContainerList,
  startContainer,
  stopContainer,
  removeContainer,
  getContainerStats,
  createAnacondaContainer,
} from "@/api/docker";
import DockerConsoleDialog from "@/components/docker/DockerConsoleDialog.vue";
import { onMounted, reactive, ref, nextTick } from "vue";

const containers = reactive([]);
let dialogVisible = reactive(false);
let selectedContainerID = reactive("");
const consoleDialog = ref();

const truncateId = (row, column, cellValue, index) => {
  let id = cellValue;
  const length = 12; // 设置截断长度
  if (id.startsWith("sha256:")) {
    id = id.substring(7); // 从第7个字符开始截取（不包含"sha256:"）
  }
  return id.length > length ? id.slice(0, length) : id;
};

const formatPorts = (row, column, cellValue, index) => {
  const formattedPorts = [];
  const ports = cellValue;
  ports.forEach((portInfo) => {
    if (portInfo.IP && portInfo.PublicPort) {
      formattedPorts.push(
        `${portInfo.IP}:${portInfo.PublicPort}->${portInfo.PrivatePort}/${portInfo.Type}`
      );
    } else {
      `${portInfo.PrivatePort}/${portInfo.Type}`;
    }
  });
  return formattedPorts.join(",");
};
const formatCpuUsage = (row, column, cellValue, index) => {
  const cpuUsage = cellValue;

  if (cpuUsage == undefined) return "";

  return `${cpuUsage.toFixed(2)}%`;
};

const formatMemoryUsage = (row, column, cellValue, index) => {
  const memoryUsage = cellValue;
  const memoryLimit = row["MemoryLimit"];

  if (memoryUsage == undefined || memoryLimit == undefined) return "";

  let usageHuman = humanReadable(memoryUsage);
  let limitHuman = humanReadable(memoryLimit);

  return usageHuman + " / " + limitHuman;
};
const humanReadable = (num) => {
  if (num === 0) return "";

  const i = Math.floor(Math.log(num) / Math.log(1024));
  return (
    (num / Math.pow(1024, i)).toFixed(2) * 1 +
    " " +
    ["B", "KB", "MB", "GB", "TB"][i]
  );
};
const stateColorClass = (state) => {
  return state === "running" ? "state-running" : "state-exited";
};

const fetchAndUpdateContainers = async () => {
  try {
    const resp = await getContainerList();
    console.log("getContainerList response:", resp.data); // Add this line
    containers.value = resp.data;
    containers.value.forEach((container) => {
      console.log("Fetching stats for container:", container.Id); // Add this line
      fetchContainerStats(container.Id);
    });
    console.log(containers.value);
  } catch (error) {
    console.error("Error fetching container list:", error);
  }
};

const handleStartContainer = async (containerID) => {
  loading = true; // 显示遮罩
  let resp = await startContainer(containerID);
  fetchAndUpdateContainers(); // 调用新方法以重新加载数据
  loading = false; // 取消遮罩
};

const handleStopContainer = async (containerID) => {
  let resp = await stopContainer(containerID);
  fetchAndUpdateContainers(); // 调用新方法以重新加载数据
  loading = false; // 取消遮罩
};

const handleRemoveContainer = async (containerID) => {
  $confirm("Are you sure to delete this container?", "Warning", {
    confirmButtonText: "Yes",
    cancelButtonText: "Cancel",
    type: "warning",
  })
    .then(async () => {
      // 使用异步函数
      let resp = await removeContainer(containerID);
      fetchAndUpdateContainers();
      loading = false;
      $message.info("Container instance deleted.");
    })
    .catch(() => {
      $message.info("Deletion cancelled.");
    });
};

const handleOpenConsole = (containerID) => {
  selectedContainerID = containerID;
  console.log("handleOpenConsole: " + selectedContainerID);
  consoleDialog.value.doOpen(selectedContainerID);
};

const fetchContainerStats = async (containerID) => {
  let resp = await getContainerStats(containerID);
  let stats = resp.data;
  const containerIndex = containers.value.findIndex(
    (container) => container.Id === containerID
  );
  if (containerIndex !== -1) {
    // Update container stats in the containers array
    containers.value[containerIndex] = {
      ...containers.value[containerIndex],
      CPUUsage: stats.CPUUsage,
      MemoryUsage: stats.MemoryUsage,
      MemoryLimit: stats.MemoryLimit,
    };
  }
};
// onMounted(async () => {
//   await nextTick();

// });
fetchAndUpdateContainers();

defineExpose({
  truncateId,
  formatPorts,
  formatCpuUsage,
  formatMemoryUsage,
  stateColorClass,
  handleStartContainer,
  handleStopContainer,
  handleRemoveContainer,
  handleOpenConsole,
});
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 15px;
}
.state-running,
.state-exited {
  font-weight: bold;
  text-transform: capitalize;
  color: white;
  padding: 5px 10px;
  border-radius: 12px;
  display: initial;
}

.state-running {
  background-color: green;
}

.state-exited {
  background-color: red;
}

.el-button span {
  font-weight: bold;
}
</style>

<script setup lang="ts">
import Sidebar from './components/Sidebar.vue';
import TopBar from './components/TopBar.vue';
import TelemetryCard from './components/TelemetryCard.vue';
import OperationsCard from './components/OperationsCard.vue';
import MediaCard from './components/MediaCard.vue';
import LogsTable from './components/LogsTable.vue';
import StatusBar from './components/StatusBar.vue';
</script>

<template>
  <div class="min-h-screen bg-background text-on-surface font-sans selection:bg-primary-container selection:text-on-primary-container">
    <div class="digital-grain"></div>

    <Sidebar />

    <main class="ml-64 flex flex-col h-screen">
      <TopBar />

      <RouterView />

      <StatusBar />
    </main>
  </div>
</template>

<style>
/* Custom scrollbar for the dark theme */
::-webkit-scrollbar {
  width: 4px;
}
::-webkit-scrollbar-track {
  background: surface-dim;
}
::-webkit-scrollbar-thumb {
  background: surface-variant;
}
::-webkit-scrollbar-thumb:hover {
  background: #primary-container;
}
</style>


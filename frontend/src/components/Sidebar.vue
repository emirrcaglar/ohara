<script setup lang="ts">
import { useRoute } from 'vue-router';
import { Library, Network, Terminal, User } from 'lucide-vue-next';

defineOptions({
  name: 'AppSidebar',
});

const route = useRoute();

const navItems = [
  { name: 'Library', icon: Library, path: '/library' },
  { name: 'Network', icon: Network, path: '/network' },
  { name: 'Logs', icon: Terminal, path: '/logs' },
];
</script>

<template>
  <aside class="fixed left-0 top-0 h-full flex flex-col bg-surface-container-low z-40 w-64 border-0 rounded-0">
      <RouterLink to="/" class="p-8 flex flex-col gap-1 group cursor-pointer hover:opacity-80 transition-opacity">
        <span class="text-2xl font-bold tracking-tighter text-primary-container">
          OHARA
        </span>
      </RouterLink>

    <nav class="flex-1 mt-4">
      <RouterLink
        v-for="item in navItems"
        :key="item.name"
        :to="item.path"
        class="flex items-center gap-4 px-6 py-4 transition-none"
        :class="route.path === item.path
          ? 'text-primary-container font-black border-l-4 border-primary-container bg-surface-container-high'
          : 'text-white/60 hover:bg-surface-container-high hover:text-primary'"
      >
        <component :is="item.icon" class="w-5 h-5" />
        <span class="uppercase tracking-tight text-sm">{{ item.name }}</span>
      </RouterLink>
    </nav>

    <div class="p-6 mt-auto">
      <div class="bg-surface-container-high p-4 flex items-center gap-3">
        <div class="w-10 h-10 bg-primary-container flex items-center justify-center">
          <User class="text-on-primary-container w-6 h-6" />
        </div>
        <div class="flex flex-col">
          <span class="text-xs font-bold uppercase tracking-widest text-on-surface">Admin_01</span>
          <span class="text-[10px] text-secondary">ROOT_ACCESS</span>
        </div>
      </div>
    </div>
  </aside>
</template>


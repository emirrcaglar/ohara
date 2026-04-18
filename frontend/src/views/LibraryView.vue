<script setup lang="ts">
import { onMounted } from 'vue'
import VaultHeader from '../components/VaultHeader.vue'
import VaultCard from '../components/VaultCard.vue'
import ImportCard from '../components/ImportCard.vue'
import { useMangaStore } from '../stores/manga'
import { getMangaCoverUrl } from '../api/manga'

const mangaStore = useMangaStore()

onMounted(() => {
  mangaStore.fetchLibrary()
})
</script>

<template>
  <main class="flex-1 flex flex-col">
    <section class="p-8 flex-1 bg-surface">
      <VaultHeader :totalManga="mangaStore.total" />

      <div v-if="mangaStore.loading" class="text-secondary">Loading...</div>
      <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6">
        <VaultCard
          v-for="manga in mangaStore.items"
          :key="manga.id"
          :manga="manga"
          :coverUrl="getMangaCoverUrl(manga.id)"
          category="MANGA_ARCHIVE"
          :stats="`${manga.currentPage} / ${manga.pageCount} PAGES`"
        />
        <ImportCard />
      </div>
    </section>
  </main>
</template>

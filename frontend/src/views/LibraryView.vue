<script setup lang="ts">
import VaultHeader from '../components/VaultHeader.vue';
import VaultCard from '../components/VaultCard.vue';
import ImportCard from '../components/ImportCard.vue';
</script>

<template>
  <main class="flex-1 flex flex-col">
    <section class="p-8 flex-1 bg-surface">
      <VaultHeader />

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6">
        <VaultCard
          title="Neon_Genesis_Vol_01"
          category="MANGA_ARCHIVE"
          badge=".cbz"
          stats="SIZE: 42.4 MB"
          image="https://lh3.googleusercontent.com/aida-public/AB6AXuAHxiaNd81L9Y2TccyqmvSBC2ajW_b_-5D9F5y0Sr9jWzId2oZ-g2ix9EBBPSSYUCea2tKN0EItHDuz0MPC1LL05AWDSu5VCkbyJe8q-OG9KRSd719IMfOTXk9014BA1l3fozpsXVlS-S7A6GKVAAEZNQpaRHzJUYShFGQVvGeYhHA-tgXXk1ikMrtwTSKhub4wq0cMifUSTHJW7F_n9uQidlx5bd_LuxRLrWnWwjdY2EySl5r-e6UHd3sFtrV4T-WBpWrAIlwa3VU"
        />
        <ImportCard />
      </div>
    </section>

  </main>
</template>

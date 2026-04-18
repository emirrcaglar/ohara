<script setup lang="ts">
import Sidebar from '../components/Sidebar.vue';
import TopBar from '../components/TopBar.vue';
import TelemetryCard from '../components/TelemetryCard.vue';
import OperationsCard from '../components/OperationsCard.vue';
import MediaCard from '../components/MediaCard.vue';
import LogsTable from '../components/LogsTable.vue';
import StatusBar from '../components/StatusBar.vue';
</script>

<template>
  <section class="flex-1 p-8 grid grid-cols-12 gap-8 overflow-y-auto">
    <div class="col-span-12 lg:col-span-4 flex flex-col gap-6">
      <TelemetryCard />
      <OperationsCard />
    </div>

    <div class="col-span-12 lg:col-span-8 flex flex-col gap-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <MediaCard
              title="NEON_ARCHIVE.MKV"
              subtitle="Featured_Stream"
              image="https://lh3.googleusercontent.com/aida-public/AB6AXuCzbM5EfiXX_i2YMYhHeBu6bid3tfm1hsCGos4Dip9sbVnLkUfCpz2woFiQGyrVAYpdSCP_RmeSWwVW7UvFxVCqxIaarle5bSK3z6SdrfrceouEvz_onyGZ54tAf-RbK_oeL6szQsk_0zUOJDyp1j1D59R5PkesgyyD8qhbwAYj4hiVWiy0VnM4VYAq9r-OOdGOz6klrAN6LIBCWPlkYkX5o3urEIrH8PbrykRIVeZzF1q4R6ZtqpgaZnLPFSL9BTcwn8Bg_eCGvSs"
              badge="4K_HDR"
              badge-type="primary"
            />
            <MediaCard
              title="VOYAGER_EXP.RAW"
              subtitle="Active_Ingest"
              image="https://lh3.googleusercontent.com/aida-public/AB6AXuCL0N5oE55rrvq7D9Yoll1-xxfT9v8t_QdVZzbliV3enndFMDWYltfxkv5Be8sIVQ3RzR0Vf_QIxl6j1qejYDbdIQU7jWdbnuJdAjuj7qxAhwuTdkMQFQSHhjGPmlCIdP1FYmUmHjDjV3z1y-MOOLbzj04Q8WmYLcWhFRV0QhJXxzf4c7huEQ8NM7NQTSbu0af1H3Ka4bSpiSmjnMNEPc8mdx-qDf7P7BWrguo_g25-UtmOX36DVNwmXTGe1yBQ3KvrpVNECol5es8"
              badge="SYNCING"
              badge-type="secondary"
            />
            </div>
      <LogsTable />
    </div>
  </section>
</template>

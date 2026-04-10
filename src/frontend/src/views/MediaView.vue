<script setup lang="ts">
import MediaDisplay from '../components/media/MediaDisplay.vue'
import MediaControls from '../components/media/MediaControls.vue' // You'll extract the controls block here
import QueueItem from '../components/media/QueueItem.vue'
</script>

<template>
  <main class="flex-1 flex flex-col md:flex-row bg-background overflow-hidden">

    <section class="flex-1 flex flex-col p-8 gap-6 min-h-0">
      <MediaDisplay
        title="NEURAL_DRIFT_SEQUENCER_v2"
        subLabel="Live Link 0.44"
        image="https://lh3.googleusercontent.com/aida-public/AB6AXuDGrpO_SpA6RTLDAqeu4PWLJWUB3RQ1nqdXOzLADi8aylzjEw-eddeqQCvyW6E8ZxUoB-y9KeLCZAxdXQetZcthuDitn_cZXVpdQoVSNL9HGjSnEKuV9rFLHnhsCDexCEtaNzDJI2oO8YbNbpczNzmBUJIhXpEeOxjk7omotL7RCqJRu4R316QbWwxA8CMWmxBE1OcxRz77dmDI6Qpelw8svqlHd_Dig0Mk6PzITc7vJS0N1pY8D76mR519rHa-MLSyZPMaoBv1n8s"
        bitrate="44.1KHZ / 24BIT"
      />

      <div class="bg-surface-container-low p-6">
        <MediaControls
        currentTime="02:44:11"
        remainingTime="-01:15:89"
        :progressPercent="66"
        :volume="75"
      />
        </div>
    </section>

    <aside class="w-full md:w-96 bg-surface-container-low border-l border-white/5 flex flex-col">
      <div class="p-8 pb-4 flex items-center justify-between">
        <h3 class="text-sm font-black uppercase tracking-widest text-secondary">Up_Next</h3>
        <span class="text-[10px] font-mono text-white/40">QUEUE: 08</span>
      </div>

      <div class="flex-1 overflow-y-auto px-4">
        <QueueItem
          title="Cyber_Glitch_Core.mp4"
          subtitle="ACTIVE_BUFFER"
          image="https://lh3.googleusercontent.com/aida-public/AB6AXuBJHWVqMUccxxJXZsqaop0OHWgxvcyVkIyP4EDE3ZaqJCcQKjtCkMkUM7yfrGG8Tufe8Y5fBvbSCVJPs43VF6TDy7dY3EYVOgfnhSXaixZL606eJ8GmOfAhP_YfV4aYk1vxg8Tv4PY-fi3lGsk1sQIgX_p4dVUq7XYWaWXJV6hfQTTlIzIS6Z5i2z_sDp8W9b8six2OR1u4H9sOstfxsnAk7CaKXy1WaJeJsfgeeTDEBHm2Qm1Tlm1EDgzejAGD6rraNMO1qVi5ImE"
          isActive
        />
        <QueueItem title="Data_Stream_Alpha.flac" subtitle="Aetheria_Vibes" image="https://lh3.googleusercontent.com/aida-public/AB6AXuCkMa8TsV3bJRCi4x44Wvj8dayL3u0kmZc3sRBLgZ0dCGfiOMzeVM99WzfPL28lbZuwgjFpwL2HL8jhohnkGm8EF7SRMWjdDjuByhe-w_AhDry2VaeyKJ21U-3XfTV2yLNt3LDmmzqH18VW_IN_w_XJoW8g7P5dqh92pPml0CYKzzH3s0HCjW9Bzd3auHi7_i9ZrLJ5AiLP9PdLyD_bYrrzBo6pTFWccWyPPNaR4YbaUPgBFQmguzzRU8GxcL6Sk7-ZlUs6oCQP_DM" />
      <div class="mt-8 p-4 grid grid-cols-2 gap-4">
        <BentoStat label="Uptime" labelColor="text-secondary">
          <p class="text-2xl font-black text-white leading-none tracking-tighter">99.8%</p>
        </BentoStat>

        <BentoStat label="Sync" labelColor="text-primary">
          <span class="material-symbols-outlined text-primary text-3xl">sync</span>
        </BentoStat>
      </div>
        </div>

      <div class="p-8 bg-surface-container-lowest font-mono text-[9px] text-white/20">
        SIGNAL: NOMINAL // AES-256-GCM
      </div>
    </aside>

  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { usePreferencesStore } from '../stores/preferences'
import { useTheme } from '../composables/useTheme'

const preferencesStore = usePreferencesStore()
const { isLight, toggleTheme } = useTheme()
const { rightToLeftSwipeForManga, scrollReadingForManga, isLoading, error } =
  storeToRefs(preferencesStore)

const isSavingPreference = ref(false)

const preferenceError = computed(() => error.value)
const isLoadingPreferences = computed(() => isLoading.value && !preferencesStore.hasLoaded)

onMounted(() => {
  if (!preferencesStore.hasLoaded) {
    void preferencesStore.loadPreferences()
  }
})

async function toggleRightToLeftSwipeForManga() {
  if (isLoadingPreferences.value || isSavingPreference.value) return

  isSavingPreference.value = true

  try {
    await preferencesStore.setRightToLeftSwipeForManga(!rightToLeftSwipeForManga.value)
  } catch {
    // The store rolls back optimistic state and exposes the error label.
  } finally {
    isSavingPreference.value = false
  }
}

async function toggleScrollReadingForManga() {
  if (isLoadingPreferences.value || isSavingPreference.value) return

  isSavingPreference.value = true

  try {
    await preferencesStore.setScrollReadingForManga(!scrollReadingForManga.value)
  } catch {
    // The store rolls back optimistic state and exposes the error label.
  } finally {
    isSavingPreference.value = false
  }
}
</script>

<template>
  <main
    class="flex-1 min-h-0 overflow-y-auto bg-surface-container-lowest p-4 md:p-8 flex flex-col gap-8"
  >
    <section class="flex-1 space-y-6">
      <div class="flex justify-between items-end border-b-0">
        <div>
          <h1
            class="font-display text-4xl font-black text-on-surface tracking-tighter uppercase leading-none"
          >
            SETTINGS
          </h1>
          <p class="font-body text-secondary text-xs font-bold tracking-[0.2em] mt-2 uppercase">
            SYSTEM_CONFIGURATION
          </p>
        </div>
      </div>

      <div class="bg-surface p-8">
        <div class="bg-surface-container-low p-8 grid gap-8 md:grid-cols-[1fr_18rem]">
          <div class="space-y-3">
            <span
              class="font-body text-secondary text-[10px] font-bold uppercase tracking-[0.24em]"
            >
              INTERFACE
            </span>
            <h2 class="font-display text-2xl font-black text-on-surface uppercase tracking-tighter">
              Appearance
            </h2>
            <p class="font-body text-on-surface-variant text-sm leading-6 max-w-xl">
              Switch between the default dark terminal palette and the Kinetic Terminal Light
              design. This setting is saved locally on this device.
            </p>
          </div>

          <button
            type="button"
            class="w-full bg-surface-container-high p-6 text-left transition-colors hover:bg-surface-container-highest active:scale-[0.99]"
            role="switch"
            :aria-checked="isLight"
            :aria-label="isLight ? 'Switch to dark mode' : 'Switch to light mode'"
            @click="toggleTheme"
          >
            <div class="flex h-full flex-col justify-between gap-8">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p
                    class="font-body text-[10px] font-bold text-secondary uppercase tracking-widest"
                  >
                    THEME_MODE
                  </p>
                  <p
                    class="font-display mt-2 text-lg font-black text-on-surface uppercase tracking-tighter"
                  >
                    {{ isLight ? 'Light mode' : 'Dark mode' }}
                  </p>
                </div>

                <div
                  class="w-14 h-7 bg-surface-container-lowest p-1 flex transition-all duration-75 shrink-0"
                  :class="isLight ? 'justify-end' : 'justify-start'"
                >
                  <div
                    class="w-5 h-5 transition-colors"
                    :class="isLight ? 'bg-primary-container' : 'bg-outline-variant/40'"
                  ></div>
                </div>
              </div>

              <div class="flex items-end justify-between gap-4">
                <span
                  class="font-body text-[10px] text-on-surface-variant uppercase tracking-widest"
                >
                  {{ isLight ? 'KINETIC_TERMINAL_LIGHT' : 'TERMINAL_DARK' }}
                </span>
                <span class="material-symbols-outlined text-primary-container">
                  {{ isLight ? 'light_mode' : 'dark_mode' }}
                </span>
              </div>
            </div>
          </button>
        </div>
      </div>

      <div class="bg-surface p-8">
        <div class="bg-surface-container-low p-8 grid gap-8 md:grid-cols-[1fr_18rem]">
          <div class="space-y-3">
            <span
              class="font-body text-secondary text-[10px] font-bold uppercase tracking-[0.24em]"
            >
              READER_BEHAVIOR
            </span>
            <h2 class="font-display text-2xl font-black text-on-surface uppercase tracking-tighter">
              Manga controls
            </h2>
            <p class="font-body text-on-surface-variant text-sm leading-6 max-w-xl">
              Configure directional gestures for manga reading flows. This preference is saved to
              your authenticated user profile.
            </p>
            <div
              v-if="preferenceError"
              class="inline-flex bg-secondary-container/10 px-3 py-2 font-body text-[10px] font-bold text-secondary-container uppercase tracking-widest"
            >
              {{ preferenceError }}
            </div>
          </div>

          <div class="space-y-4">
            <button
              type="button"
              class="w-full bg-surface-container-high p-6 text-left transition-colors hover:bg-surface-container-highest active:scale-[0.99]"
              role="switch"
              :aria-checked="rightToLeftSwipeForManga"
              aria-label="Right to left swipe for manga"
              :disabled="isLoadingPreferences || isSavingPreference"
              @click="toggleRightToLeftSwipeForManga"
            >
              <div class="flex h-full flex-col justify-between gap-8">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <p
                      class="font-body text-[10px] font-bold text-secondary uppercase tracking-widest"
                    >
                      SWIPE_DIRECTION
                    </p>
                    <p
                      class="font-display mt-2 text-lg font-black text-on-surface uppercase tracking-tighter"
                    >
                      Right to left swipe for manga
                    </p>
                  </div>

                  <div
                    class="w-14 h-7 bg-surface-container-lowest p-1 flex transition-all duration-75 shrink-0"
                    :class="rightToLeftSwipeForManga ? 'justify-end' : 'justify-start'"
                  >
                    <div
                      class="w-5 h-5 transition-colors"
                      :class="
                        rightToLeftSwipeForManga ? 'bg-primary-container' : 'bg-outline-variant/40'
                      "
                    ></div>
                  </div>
                </div>

                <div class="flex items-end justify-between gap-4">
                  <span
                    class="font-body text-[10px] text-on-surface-variant uppercase tracking-widest"
                  >
                    {{
                      isLoadingPreferences
                        ? 'LOADING'
                        : isSavingPreference
                          ? 'SYNCING'
                          : rightToLeftSwipeForManga
                            ? 'ENABLED'
                            : 'DISABLED'
                    }}
                  </span>
                  <span class="material-symbols-outlined text-primary-container">
                    {{ rightToLeftSwipeForManga ? 'swipe_left' : 'swipe' }}
                  </span>
                </div>
              </div>
            </button>

            <button
              type="button"
              class="w-full bg-surface-container-high p-6 text-left transition-colors hover:bg-surface-container-highest active:scale-[0.99]"
              role="switch"
              :aria-checked="scrollReadingForManga"
              aria-label="Scroll reading for manga"
              :disabled="isLoadingPreferences || isSavingPreference"
              @click="toggleScrollReadingForManga"
            >
              <div class="flex h-full flex-col justify-between gap-8">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <p
                      class="font-body text-[10px] font-bold text-secondary uppercase tracking-widest"
                    >
                      READING_MODE
                    </p>
                    <p
                      class="font-display mt-2 text-lg font-black text-on-surface uppercase tracking-tighter"
                    >
                      Scroll reading for manga
                    </p>
                  </div>

                  <div
                    class="w-14 h-7 bg-surface-container-lowest p-1 flex transition-all duration-75 shrink-0"
                    :class="scrollReadingForManga ? 'justify-end' : 'justify-start'"
                  >
                    <div
                      class="w-5 h-5 transition-colors"
                      :class="
                        scrollReadingForManga ? 'bg-primary-container' : 'bg-outline-variant/40'
                      "
                    ></div>
                  </div>
                </div>

                <div class="flex items-end justify-between gap-4">
                  <span
                    class="font-body text-[10px] text-on-surface-variant uppercase tracking-widest"
                  >
                    {{
                      isLoadingPreferences
                        ? 'LOADING'
                        : isSavingPreference
                          ? 'SYNCING'
                          : scrollReadingForManga
                            ? 'SCROLL'
                            : 'SWIPE'
                    }}
                  </span>
                  <span class="material-symbols-outlined text-primary-container">
                    {{ scrollReadingForManga ? 'vertical_align_bottom' : 'swipe' }}
                  </span>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.font-display {
  font-family: 'Space Grotesk', sans-serif;
}
.font-body {
  font-family: 'Space Grotesk', sans-serif;
}
</style>

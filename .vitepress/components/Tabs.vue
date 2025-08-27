
<template>
  <div class="tabs" ref="tabsContainer">
    <div class="tab-buttons">
      <button
        v-for="(tab, index) in tabTitles"
        :key="index"
        :class="{ active: index === activeTab }"
        @click="activeTab = index"
      >
        {{ tab }}
      </button>
    </div>
    <div class="tab-content">
      <slot />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'

const activeTab = ref(0)
const tabTitles = ref([])
const tabsContainer = ref(null)

onMounted(() => {
  const tabElements = tabsContainer.value.querySelectorAll('[data-tab-name]')
  tabTitles.value = Array.from(tabElements).map(el => el.getAttribute('data-tab-name'))

  const updateVisibility = () => {
    tabElements.forEach((el, index) => {
      el.style.display = index === activeTab.value ? 'block' : 'none'
    })
  }

  updateVisibility()
  watch(activeTab, updateVisibility)
})
</script>

<style scoped>
.tabs {
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
}

.tab-buttons {
  display: flex;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--vp-c-divider);
  background-color: var(--vp-c-bg-soft);
  padding: 4px;
}

.tab-buttons button {
  padding: 8px 16px;
  margin-right: 4px;
  cursor: pointer;
  background-color: transparent;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px 4px 0 0;
  color: var(--vp-c-text-2);
  font-weight: 500;
}

.tab-buttons button.active {
  background-color: var(--vp-c-bg);
  color: var(--vp-c-brand);
  font-weight: bold;
  border-bottom: none;
}

.tab-content {
  padding: 16px;
}
</style>

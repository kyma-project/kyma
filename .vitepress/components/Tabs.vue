<template>
  <div class="tabs" ref="tabsContainer">
    <div class="tab-buttons">
      <button
        v-for="(tab, index) in tabTitles"
        :key="index"
        :class="{ active: index === activeTab }"
        @click="activeTab = index"
        v-html="tab"
      >
      </button>
    </div>
    <div class="tab-content">
      <div v-if="isPropDriven" v-html="parsedTabs[activeTab]?.content"></div>
      <slot v-else />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'

const props = defineProps({
  tabsData: {
    type: String,
    required: false
  }
})

const activeTab = ref(0)
const tabTitles = ref([])
const tabsContainer = ref(null)
const parsedTabs = ref([])

const isPropDriven = computed(() => !!props.tabsData);

function decodeBase64(str) {
  // Use Buffer on server-side (SSR), and atob on client-side
  if (typeof window === 'undefined') {
    return Buffer.from(str, 'base64').toString('utf-8');
  }
  return decodeURIComponent(window.atob(str).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));
}

onMounted(() => {
  if (isPropDriven.value) {
    try {
      const decoded = decodeBase64(props.tabsData);
      const tabs = JSON.parse(decoded);
      parsedTabs.value = tabs;
      tabTitles.value = tabs.map(tab => tab.label);
    } catch (e) {
      console.error('Failed to parse tabs data:', e);
    }
  } else {
    // Old slot-based logic for backward compatibility
    const tabElements = tabsContainer.value.querySelectorAll('[data-tab-name]')
    tabTitles.value = Array.from(tabElements).map(el => el.getAttribute('data-tab-name'))

    const updateVisibility = () => {
      tabElements.forEach((el, index) => {
        el.style.display = index === activeTab.value ? 'block' : 'none'
      })
    }
    watch(activeTab, updateVisibility, { immediate: true })
  }
})
</script>

<style scoped>
.tabs {
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  margin-top: 1em;
  margin-bottom: 1em;
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

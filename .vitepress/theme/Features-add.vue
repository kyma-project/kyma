<template>
  <div v-if="features && features.length" class="fetures-add">
    <h2 v-html="description"></h2>
    <div class="fetures-add-grid">
        
      <div class="feature-card" v-for="(feature, index) in features" :key="index">
        <div v-if="feature.icon" class="icon-wrapper" :style="{ width: feature.icon.width, height: feature.icon.height }">
          <img :src="withBase(feature.icon.src)" :alt="feature.title" class="icon" :style="{ width: feature.icon.width, height: feature.icon.height }"/>
        </div>
        <div class="text-content">
          <h2 class="title" v-html="feature.title"></h2>
          <p v-if="feature.details" class="details" v-html="feature.details"></p>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { withBase, useData } from 'vitepress'
import { computed } from 'vue'

const { frontmatter } = useData()
const features = computed(() => frontmatter.value.features_additonal)
const description = computed(() => frontmatter.value.features_additonal_description)
</script>

<style scoped>
.feature-card {
  display: flex;
  flex-direction: column;
  padding: 1.5rem;
  border: 1px solid var(--vp-c-bg-soft);
  border-radius: 8px;
  background-color: var(--vp-c-bg);
  box-shadow: 0 0 12px 2px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.3s ease;
  align-items: center;
  height: 100%;
  box-sizing: border-box;
}

.feature-card:hover {
  box-shadow: 0 0 12px 10px rgba(0, 0, 0, 0.1);
}

.icon {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.icon img {
  max-width: 100%;
  max-height: 100%;
}

.title {
  margin: 0.5rem 0;
  font-size: 1rem;
  font-weight: 600;
  text-align: center;
}

.details {
  margin: 0;
  font-size: 0.875rem;
  color: var(--vp-c-text-1);
  text-align: center;
}
.fetures-add {
  padding: 2rem 0;
  border-top: 1px solid var(--vp-c-divider);
  margin-top: 2rem;
}
.fetures-add h2 {
  text-align: center;
  margin-bottom: 2rem;
  font-size: 1.5rem;
  font-weight: 600;
}
.fetures-add-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  justify-items: center;
}
.feature-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%; /* or a fixed height if needed */
}

.icon-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 1rem;
}
</style>

<template>
  <div class="tab-slot" :data-tab-name="name">
    <slot />
  </div>
</template>

<script setup>
const props = defineProps({
  name: {
    type: String,
    required: true
  }
})
</script>

<style scoped>
.tab-slot {
  display: none;
}
</style>

<script setup>
import DefaultTheme from 'vitepress/theme'
import MyFooter from './MyFooter.vue'
import FeaturesAdd from './Features-add.vue'

const { Layout } = DefaultTheme
</script>

<template>
  <Layout>
    <!-- Slot for content after the main features on the homepage -->
    <template #home-features-after>
      <FeaturesAdd />
    </template>
    
    <!-- Slot for content at the bottom of every page -->
    <template #layout-bottom>
      <MyFooter />
    </template>
  </Layout>
</template>

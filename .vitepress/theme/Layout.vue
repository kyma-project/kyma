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

<script setup>
import { onMounted } from 'vue'
import DefaultTheme from 'vitepress/theme'
import MyFooter from './MyFooter.vue'
import FeaturesAdd from './Features-add.vue'

const { Layout } = DefaultTheme
onMounted(() => {
  const { protocol, hostname, port, pathname, hash } = window.location

  // Rebuild base URL
  let baseUrl = `${protocol}//${hostname}`
  if (port) {
    baseUrl += `:${port}`
  }
  // Include pathname if exists
  if (pathname){
    baseUrl += pathname
  }
  else {
    baseUrl += '/'
  }

  if (hash.startsWith('#/')) {
    const pathAndQuery = hash.slice(2) // remove "#/"
    const [path, query] = pathAndQuery.split('?')

    const pathSegments = path.split('/')
    if (pathSegments.length >= 2) {
      const root = pathSegments[0] // module name, es: "busola", "cli", "istio"... This is how standard are organised folders in external-content
      const rest = pathSegments.slice(1).join('/') // es: this is the rest, cointening user/...

      // Build a new URL with "docs" inside
      let newUrl = `external-content/${root}/docs/${rest}.html`

      // Translate anchor (id=...)
      if (query && query.startsWith('id=')) {
        const anchor = query.split('=')[1]
        newUrl += `#${anchor}`
      }

      const href= baseUrl + newUrl
      window.location.href = href
    } else {
      console.warn('[DEBUG] Unexpected hash format:', hash)
    }
  }
})

</script>

<template>
  <div v-if="html" v-html="html"></div>
  <div v-else-if="error">
    <h3>❌ Errore nel caricamento del contenuto.</h3>
    <p>Could not fetch content from the remote repository.</p>
  </div>
  <div v-else>
    <p>⏳ Caricamento...</p>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { useRoute, useData } from 'vitepress';
import MarkdownIt from 'markdown-it';
import path from 'path-browserify';

const html = ref('');
const error = ref(false);
const route = useRoute();
const { theme } = useData();

const md = new MarkdownIt();

async function fetchContent() {
  html.value = '';
  error.value = false;

  try {
    const currentPath = route.path;
    const aliases = theme.value.alias;
    let url = '';

    if (!aliases) {
      throw new Error("`themeConfig.alias` is not defined in .vitepress/config.mjs");
    }

    let aliasFound;
    for (const [aliasPath, rawUrl] of Object.entries(aliases)) {
      const regex = new RegExp(aliasPath.replace('(.*)', '(.*)'));
      const match = currentPath.match(regex);

      if (match) {
        aliasFound = rawUrl;
        const remotePath = match[1] || 'README.md';
        url = rawUrl.replace('$1', remotePath);
        if (url.endsWith('/')) {
          url = path.join(url, 'README.md');
        }
        else if (!url.endsWith('.md')) {
          //replace current extension
          url = url.replace(/.[^.]+$/, '.md');
        }
        aliasFound = url.replace(/.[^/]+$/, '/');
        break;
      }
    }

    if (!url) {
      throw new Error(`No matching alias found for path: ${currentPath}`);
    }

    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`Failed to fetch ${url}: ${res.statusText}`);
    }

    let markdown = await res.text();
    html.value = md.render(markdown);

  } catch (e) {
    console.error(e);
    error.value = true;
  }
}

onMounted(fetchContent);
watch(() => route.path, fetchContent);
</script>

import DefaultTheme from 'vitepress/theme'
import DynamicContentLoader from '../components/DynamicContentLoader.vue'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    app.component('DynamicContentLoader', DynamicContentLoader)
  }
}

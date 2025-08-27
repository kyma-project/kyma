import DefaultTheme from 'vitepress/theme'
import Tabs from '../components/Tabs.vue'
import Tab from '../components/Tab.vue'


export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    app.component('Tabs', Tabs)
    app.component('Tab', Tab)
  }
}

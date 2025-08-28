import DefaultTheme from 'vitepress/theme'
import Layout from './Layout.vue'
import './style.css'
import Tabs from '../components/Tabs.vue'
import Tab from '../components/Tab.vue'

export default {
  ...DefaultTheme,
  Layout,
  enhanceApp({ app, router, siteData }) {
    app.component('Tabs', Tabs)
    app.component('Tab', Tab)
    // You can register global components here if needed
  }
}

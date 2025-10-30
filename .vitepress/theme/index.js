import DefaultTheme from 'vitepress/theme';
import Layout from './Layout.vue';
import './style.css';
import Tabs from '../components/Tabs.vue';
import Tab from '../components/Tab.vue';

export default {
  ...DefaultTheme,
  Layout,
  enhanceApp({ app, router, siteData }) {
    app.component('Tabs', Tabs);
    app.component('Tab', Tab);
    // You can register global components here if needed
    if (typeof window !== 'undefined') {
      const stylesheets = [
        {
          rel: 'icon',
          type: 'image/png',
          href: './assets/favicon-16x16.png',
          sizes: '16x16',
        },
        {
          rel: 'icon',
          type: 'image/png',
          href: './assets/favicon-32x32.png',
          sizes: '32x32',
        },
        { rel: 'shortcut icon', href: './assets/favicon.ico' },
        {
          rel: 'apple-touch-icon',
          sizes: '180x180',
          href: './assets/apple-touch-icon.png',
        },
      ];

      stylesheets.forEach((attrs) => {
        const link = document.createElement('link');
        Object.entries(attrs).forEach(([key, value]) => {
          link.setAttribute(key, value);
        });
        document.head.appendChild(link);
      });
    }
  },
};

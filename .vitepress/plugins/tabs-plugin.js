export default function tabsPlugin(md) {
  const originalRender = md.render;

  md.render = function (src, env) {
    const tabBlockRegex =
      /<!--\s*tabs:start\s*-->([\s\S]*?)<!--\s*tabs:end\s*-->/g;

    const newSrc = src.replace(tabBlockRegex, (match, content) => {
      const tabs = [];
      const headingRegex = /^####\s+(?:\*\*|__)(.*?)(?:\*\*|__)$/;
      let currentTab = null;
      let currentContent = '';

      content
        .trim()
        .split('\n')
        .forEach((line) => {
          const headingMatch = line.match(headingRegex);
          if (headingMatch) {
            if (currentTab) {
              tabs.push({
                label: currentTab,
                content: md.render(currentContent.trim()),
              });
            }
            currentTab = headingMatch[1].trim();
            currentContent = '';
          } else {
            currentContent += line + '\n';
          }
        });

      if (currentTab) {
        tabs.push({
          label: currentTab,
          content: md.render(currentContent.trim()),
        });
      }

      if (tabs.length > 0) {
        const tabsData = Buffer.from(JSON.stringify(tabs)).toString('base64');
        return `<Tabs tabs-data="${tabsData}" />`;
      }

      return ''; // Return empty string if no tabs were found inside the block
    });

    return originalRender.call(this, newSrc, env);
  };
}

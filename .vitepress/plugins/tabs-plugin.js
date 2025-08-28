export default function tabsPlugin(md) {
  function tabsRule(state, startLine, endLine, silent) {
    let pos = state.bMarks[startLine] + state.tShift[startLine];
    let max = state.eMarks[startLine];

    if (state.src.slice(pos, max).trim() !== '<!-- tabs:start -->') {
      return false;
    }

    if (silent) {
      return true;
    }

    let nextLine = startLine + 1;
    let endLineNumber = -1;
    while (nextLine < endLine) {
      pos = state.bMarks[nextLine] + state.tShift[nextLine];
      max = state.eMarks[nextLine];
      if (state.src.slice(pos, max).trim() === '<!-- tabs:end -->') {
        endLineNumber = nextLine;
        break;
      }
      nextLine++;
    }

    if (endLineNumber === -1) {
      return false;
    }

    const contentStart = state.bMarks[startLine + 1];
    const contentEnd = state.bMarks[endLineNumber];
    const content = state.src.slice(contentStart, contentEnd);

    const tabs = [];
    const headingRegex = /^####\s+(.*)/;

    let currentTab = null;
    let currentContent = '';

    content.split('\n').forEach(line => {
      const match = line.match(headingRegex);
      if (match) {
        if (currentTab) {
          tabs.push({ label: currentTab, content: md.render(currentContent.trim()) });
        }
        currentTab = match[1];
        currentContent = '';
      } else {
        currentContent += line + '\n';
      }
    });

    if (currentTab) {
      tabs.push({ label: currentTab, content: md.render(currentContent.trim()) });
    }

    const token = state.push('tabs_block', 'Tabs', 0);
    token.info = JSON.stringify(tabs);
    token.map = [startLine, endLineNumber + 1];
    
    state.line = endLineNumber + 1;

    return true;
  }

  md.block.ruler.before('fence', 'tabs', tabsRule);

  md.renderer.rules.tabs_block = (tokens, idx) => {
    const token = tokens[idx];
    const tabs = JSON.parse(token.info);
    const tabsData = encodeURIComponent(JSON.stringify(tabs));
    return `<Tabs :tabs='JSON.parse(decodeURIComponent("${tabsData}"))' />`;
  };
}

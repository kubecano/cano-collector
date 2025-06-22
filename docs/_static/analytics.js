// Analytics and tracking functions for Cano-Collector documentation

function setupCopyListener() {
  let codeCells = document.querySelectorAll("code");
  codeCells.forEach(element => element.addEventListener('copy', (event) => {
    reportCopy(element);
  }));
  let copyButtons = document.querySelectorAll("button[data-clipboard-target*=code]");
  copyButtons.forEach(element => element.addEventListener('click', (event) => {
    reportCopy(element);
  }));
}

function trackPageViewEvent() {
  const pageUrl = window.location.href;
  trackEvent('DocsPageview', {'pageUrl': pageUrl});
  if (pageUrl.includes('installation')) {
    trackEvent('InstallationPageview', {'pageUrl': pageUrl});
  }
}

function trackEvent(event, properties) {
  // Google Analytics tracking
  if (typeof gtag !== 'undefined') {
    gtag('event', event, properties);
  }
}

function reportCopy(baseElement) {
  // don't track users who ask not to be tracked
  if (navigator.doNotTrack === "1") {
    return
  }
  let id_element = (baseElement.closest('div[id^=cb-]')); // corresponds to the :name: in the code-blocks. prefix cb-
  if (id_element) {
    const path = window.location.pathname;
    const page = path.split("/").pop();
    if (page && page.endsWith('html')) {
      trackEvent('copied from a codeblock in the docs');
      trackEvent('copied from a codeblock on: ' + page);
    }
    trackEvent('copied from codeblock: ' + id_element.getAttribute('id'));
  }
} 
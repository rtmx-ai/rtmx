// RTMX Dashboard Application
// Alpine.js app state and htmx configuration

function rtmxApp() {
  return {
    page: 'status',
    connected: true,
    ws: null,

    init() {
      // Determine initial page from URL
      var path = window.location.pathname;
      if (path === '/' || path === '') {
        this.page = 'status';
      } else {
        this.page = path.replace(/^\//, '').split('/')[0];
      }

      // Handle browser back/forward
      window.addEventListener('popstate', function() {
        var p = window.location.pathname;
        this.page = (p === '/' || p === '') ? 'status' : p.replace(/^\//, '').split('/')[0];
      }.bind(this));
    },

    navigate(pageName) {
      this.page = pageName;
    }
  };
}

// htmx configuration
document.addEventListener('DOMContentLoaded', function() {
  // Add loading indicators
  document.body.addEventListener('htmx:beforeRequest', function() {
    var indicator = document.getElementById('loading-indicator');
    if (indicator) indicator.style.display = 'block';
  });

  document.body.addEventListener('htmx:afterRequest', function() {
    var indicator = document.getElementById('loading-indicator');
    if (indicator) indicator.style.display = 'none';
  });
});

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

// D3.js force-directed graph renderer
function renderGraph(data) {
  var container = document.getElementById('graph-container');
  var svg = d3.select('#graph-svg');
  if (!container || !svg.node()) return;

  svg.selectAll('*').remove();

  var width = container.clientWidth;
  var height = container.clientHeight || 600;

  svg.attr('viewBox', [0, 0, width, height]);

  var statusColor = {
    'COMPLETE': '#16a34a',
    'PARTIAL': '#d97706',
    'MISSING': '#dc2626',
    'NOT_STARTED': '#71717a'
  };

  var nodes = (data.nodes || []).map(function(n) { return Object.assign({}, n); });
  var links = (data.edges || []).map(function(e) { return {source: e.source, target: e.target}; });

  var simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links).id(function(d) { return d.id; }).distance(60))
    .force('charge', d3.forceManyBody().strength(-80))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(12));

  var g = svg.append('g');

  // Zoom
  svg.call(d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', function(event) { g.attr('transform', event.transform); }));

  var link = g.append('g')
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', '#9ca3af')
    .attr('stroke-opacity', 0.4)
    .attr('stroke-width', 1)
    .attr('marker-end', 'url(#arrow)');

  // Arrow marker
  svg.append('defs').append('marker')
    .attr('id', 'arrow')
    .attr('viewBox', '0 -3 6 6')
    .attr('refX', 14)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-3L6,0L0,3')
    .attr('fill', '#9ca3af');

  var node = g.append('g')
    .selectAll('g')
    .data(nodes)
    .join('g')
    .call(d3.drag()
      .on('start', function(event, d) {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x; d.fy = d.y;
      })
      .on('drag', function(event, d) { d.fx = event.x; d.fy = event.y; })
      .on('end', function(event, d) {
        if (!event.active) simulation.alphaTarget(0);
        d.fx = null; d.fy = null;
      }));

  node.append('circle')
    .attr('r', 6)
    .attr('fill', function(d) { return statusColor[d.status] || '#71717a'; })
    .attr('stroke', '#fff')
    .attr('stroke-width', 1.5);

  node.append('title')
    .text(function(d) { return d.id + ' (' + d.group + ')'; });

  // Show labels on hover
  node.on('mouseover', function(event, d) {
    d3.select(this).select('circle').attr('r', 9);
    g.append('text')
      .attr('class', 'hover-label')
      .attr('x', d.x + 12)
      .attr('y', d.y + 4)
      .attr('fill', '#374151')
      .attr('font-size', '11px')
      .text(d.id);
  }).on('mouseout', function() {
    d3.select(this).select('circle').attr('r', 6);
    g.selectAll('.hover-label').remove();
  });

  simulation.on('tick', function() {
    link
      .attr('x1', function(d) { return d.source.x; })
      .attr('y1', function(d) { return d.source.y; })
      .attr('x2', function(d) { return d.target.x; })
      .attr('y2', function(d) { return d.target.y; });
    node.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
  });
}

// Re-render graph after htmx swaps content
document.addEventListener('htmx:afterSwap', function(event) {
  var graphScript = event.detail.target.querySelector('script');
  if (graphScript && graphScript.textContent.indexOf('renderGraph') >= 0) {
    eval(graphScript.textContent);
  }
});

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

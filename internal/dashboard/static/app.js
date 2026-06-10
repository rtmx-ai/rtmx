// RTMX Dashboard Application
// Alpine.js app state and htmx configuration

function rtmxApp() {
  return {
    page: 'status',
    connected: true,
    ws: null,
    panelOpen: false,
    panelReqId: '',
    paletteOpen: false,

    init() {
      // Determine initial page from URL
      var path = window.location.pathname;
      if (path === '/' || path === '') {
        this.page = 'status';
      } else {
        this.page = path.replace(/^\//, '').split('/')[0];
      }

      // Handle browser back/forward
      window.addEventListener('popstate', function(e) {
        var p = window.location.pathname;
        this.page = (p === '/' || p === '') ? 'status' : p.replace(/^\//, '').split('/')[0];
        // REQ-DASH-029: Close panel on back if it was a panel push
        if (this.panelOpen && !(e.state && e.state.panel)) {
          this.panelOpen = false;
        }
      }.bind(this));

      // REQ-DASH-031: Cmd+K / Ctrl+K opens command palette
      // REQ-DASH-035: j/k row selection
      // REQ-DASH-036: g-prefix page shortcuts
      var gPending = false;
      window.addEventListener('keydown', function(e) {
        // Skip shortcuts when focused on input/select/textarea
        var tag = document.activeElement && document.activeElement.tagName;
        if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') {
          // Only Cmd+K bypasses input suppression
          if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            this.paletteOpen = !this.paletteOpen;
            if (this.paletteOpen) {
              this.$nextTick(function() {
                var input = document.getElementById('palette-input');
                if (input) { input.value = ''; input.focus(); }
              });
            }
          }
          return;
        }

        // Cmd+K / Ctrl+K
        if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
          e.preventDefault();
          this.paletteOpen = !this.paletteOpen;
          if (this.paletteOpen) {
            this.$nextTick(function() {
              var input = document.getElementById('palette-input');
              if (input) { input.value = ''; input.focus(); }
            });
          }
          return;
        }

        // g-prefix shortcuts
        if (gPending) {
          gPending = false;
          var pageMap = { s: 'status', r: 'requirements', d: 'graph', k: 'kanban', v: 'releases', h: 'health', a: 'agents' };
          var urlMap = { s: '/', r: '/requirements', d: '/graph', k: '/kanban', v: '/releases', h: '/health', a: '/agents' };
          var partialMap = { s: '/partials/status', r: '/partials/requirements', d: '/partials/graph', k: '/partials/kanban', v: '/partials/releases', h: '/partials/health', a: '/partials/agents' };
          if (pageMap[e.key]) {
            e.preventDefault();
            this.navigate(pageMap[e.key]);
            htmx.ajax('GET', partialMap[e.key], { target: '#content', swap: 'innerHTML', pushUrl: urlMap[e.key] });
          }
          return;
        }
        if (e.key === 'g') {
          gPending = true;
          setTimeout(function() { gPending = false; }, 1000);
          return;
        }

        // ? key opens help overlay
        if (e.key === '?') {
          e.preventDefault();
          window.dispatchEvent(new CustomEvent('toggle-help'));
          return;
        }

        // j/k row selection
        if (e.key === 'j' || e.key === 'k') {
          e.preventDefault();
          var rows = document.querySelectorAll('.data-table tbody tr');
          if (rows.length === 0) return;
          var currentIdx = -1;
          rows.forEach(function(row, i) {
            if (row.classList.contains('row-selected')) currentIdx = i;
          });
          var nextIdx;
          if (e.key === 'j') {
            nextIdx = currentIdx < rows.length - 1 ? currentIdx + 1 : 0;
          } else {
            nextIdx = currentIdx > 0 ? currentIdx - 1 : rows.length - 1;
          }
          rows.forEach(function(row) { row.classList.remove('row-selected'); });
          rows[nextIdx].classList.add('row-selected');
          rows[nextIdx].scrollIntoView({ block: 'nearest' });
        }
      }.bind(this));
    },

    navigate(pageName) {
      this.page = pageName;
    },

    openPanel(reqId) {
      this.panelReqId = reqId;
      this.panelOpen = true;
    },

    closePanel() {
      this.panelOpen = false;
      this.panelReqId = '';
    }
  };
}

// REQ-DASH-014: Toast notification manager
function toastManager() {
  var nextId = 0;
  return {
    toasts: [],

    add(detail) {
      var id = ++nextId;
      var toast = { id: id, type: detail.type || 'info', message: detail.message || '', visible: true };
      this.toasts.push(toast);
      // Max 5 visible
      while (this.toasts.length > 5) {
        this.toasts.shift();
      }
      // Auto-dismiss after 4 seconds
      var self = this;
      setTimeout(function() { self.remove(id); }, 4000);
    },

    remove(id) {
      var idx = this.toasts.findIndex(function(t) { return t.id === id; });
      if (idx >= 0) {
        this.toasts[idx].visible = false;
        var self = this;
        setTimeout(function() {
          self.toasts = self.toasts.filter(function(t) { return t.id !== id; });
        }, 300);
      }
    }
  };
}

// Global showToast helper
function showToast(type, message) {
  window.dispatchEvent(new CustomEvent('show-toast', { detail: { type: type, message: message } }));
}

// REQ-DASH-017: Inline select editor for status/priority cells
function inlineSelect(reqId, field, currentValue) {
  return {
    editing: false,
    loading: false,
    original: currentValue,

    startEdit() {
      this.editing = true;
      this.$nextTick(function() { this.$refs.input.focus(); }.bind(this));
    },

    save(newValue) {
      if (newValue === this.original) {
        this.editing = false;
        return;
      }
      this.editing = false;
      this.loading = true;
      var self = this;
      var body = {};
      body[field] = newValue;
      fetch('/api/requirements/' + reqId, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      }).then(function(resp) {
        self.loading = false;
        if (resp.ok) {
          htmx.ajax('GET', '/partials/requirements/row/' + reqId, { target: '#row-' + reqId, swap: 'outerHTML' });
          showToast('success', 'Saved ' + reqId);
        } else {
          resp.json().then(function(data) {
            showToast('error', data.error || 'Save failed');
          });
        }
      }).catch(function() {
        self.loading = false;
        showToast('error', 'Network error');
      });
    },

    cancel() {
      this.editing = false;
    }
  };
}

// REQ-DASH-018: Inline text editor for assignee/sprint cells
function inlineText(reqId, field, currentValue) {
  return {
    editing: false,
    loading: false,
    value: currentValue,
    original: currentValue,

    startEdit() {
      this.editing = true;
      this.$nextTick(function() { this.$refs.input.focus(); this.$refs.input.select(); }.bind(this));
    },

    save(newValue) {
      this.editing = false;
      if (newValue === this.original) return;
      this.value = newValue;
      this.original = newValue;
      this.loading = true;
      var self = this;
      var body = {};
      body[field] = newValue;
      fetch('/api/requirements/' + reqId, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      }).then(function(resp) {
        self.loading = false;
        if (resp.ok) {
          showToast('success', 'Saved ' + reqId);
        } else {
          resp.json().then(function(data) {
            showToast('error', data.error || 'Save failed');
          });
        }
      }).catch(function() {
        self.loading = false;
        showToast('error', 'Network error');
      });
    },

    cancel() {
      this.editing = false;
      this.value = this.original;
    }
  };
}

// REQ-DASH-040/042: Bulk action bar
function bulkActionBar() {
  return {
    selectedCount: 0,

    init() {
      var self = this;
      // Watch for checkbox changes
      document.addEventListener('change', function(e) {
        if (e.target.classList.contains('bulk-select') || e.target.classList.contains('bulk-select-all')) {
          self.selectedCount = document.querySelectorAll('.bulk-select:checked').length;
        }
      });
    },

    getSelectedIds() {
      var ids = [];
      document.querySelectorAll('.bulk-select:checked').forEach(function(cb) {
        ids.push(cb.value);
      });
      return ids;
    },

    clearSelection() {
      document.querySelectorAll('.bulk-select, .bulk-select-all').forEach(function(cb) { cb.checked = false; });
      this.selectedCount = 0;
    },

    executeBulk(updates) {
      var ids = this.getSelectedIds();
      if (ids.length === 0) return;
      var self = this;
      fetch('/api/requirements/bulk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ req_ids: ids, updates: updates })
      }).then(function(resp) {
        if (resp.ok) {
          showToast('success', 'Updated ' + ids.length + ' requirements');
          self.clearSelection();
          // Refresh the table
          htmx.ajax('GET', '/partials/requirements', { target: '#content', swap: 'innerHTML' });
        } else {
          resp.json().then(function(data) {
            showToast('error', data.error || 'Bulk update failed');
          });
        }
      }).catch(function() {
        showToast('error', 'Network error');
      });
    },

    bulkSetStatus() {
      var status = prompt('Set status to (COMPLETE, PARTIAL, MISSING, NOT_STARTED):');
      if (status) this.executeBulk({ status: status });
    },

    bulkSetPriority() {
      var priority = prompt('Set priority to (P0, HIGH, MEDIUM, LOW):');
      if (priority) this.executeBulk({ priority: priority });
    },

    bulkSetAssignee() {
      var assignee = prompt('Set assignee to:');
      if (assignee !== null) this.executeBulk({ assignee: assignee });
    },

    bulkSetSprint() {
      var sprint = prompt('Set sprint to:');
      if (sprint !== null) this.executeBulk({ sprint: sprint });
    }
  };
}

// REQ-DASH-033: Palette arrow key navigation
function paletteKeyNav(event) {
  var results = document.querySelectorAll('.palette-result');
  if (results.length === 0) return;

  var currentIdx = -1;
  results.forEach(function(el, i) {
    if (el.classList.contains('selected')) currentIdx = i;
  });

  if (event.key === 'ArrowDown') {
    event.preventDefault();
    var nextIdx = currentIdx < results.length - 1 ? currentIdx + 1 : 0;
    results.forEach(function(el) { el.classList.remove('selected'); });
    results[nextIdx].classList.add('selected');
    results[nextIdx].scrollIntoView({ block: 'nearest' });
  } else if (event.key === 'ArrowUp') {
    event.preventDefault();
    var prevIdx = currentIdx > 0 ? currentIdx - 1 : results.length - 1;
    results.forEach(function(el) { el.classList.remove('selected'); });
    results[prevIdx].classList.add('selected');
    results[prevIdx].scrollIntoView({ block: 'nearest' });
  } else if (event.key === 'Enter') {
    event.preventDefault();
    var selected = document.querySelector('.palette-result.selected');
    if (selected) selected.click();
  }
}

// REQ-DASH-037: Keyboard shortcut help overlay
function keyboardHelpOverlay() {
  return {
    helpOpen: false,
    toggle() { this.helpOpen = !this.helpOpen; },
    shortcuts: [
      { keys: 'Cmd+K', desc: 'Open command palette' },
      { keys: 'g s', desc: 'Go to Status' },
      { keys: 'g r', desc: 'Go to Requirements' },
      { keys: 'g d', desc: 'Go to Graph' },
      { keys: 'g k', desc: 'Go to Kanban' },
      { keys: 'g v', desc: 'Go to Releases' },
      { keys: 'g h', desc: 'Go to Health' },
      { keys: 'g a', desc: 'Go to Agents' },
      { keys: 'j / k', desc: 'Select next/prev row' },
      { keys: 'Escape', desc: 'Close panel/palette' },
      { keys: '?', desc: 'Toggle this help' }
    ]
  };
}

// REQ-DASH-022: Per-column filter dropdown component
function columnFilter(paramName, currentValue, options) {
  var sel = {};
  if (currentValue) {
    currentValue.split(',').forEach(function(v) {
      v = v.trim();
      if (v) sel[v] = true;
    });
  }
  return {
    open: false,
    selected: sel,
    paramName: paramName,

    toggle() {
      this.open = !this.open;
    },

    toggleValue(val) {
      if (this.selected[val]) {
        delete this.selected[val];
      } else {
        this.selected[val] = true;
      }
    },

    apply() {
      this.open = false;
      // Direct htmx request with all current params
      var params = new URLSearchParams();
      var csvValue = Object.keys(this.selected).join(',');
      if (csvValue) {
        params.set(this.paramName, csvValue);
      } else {
        params.delete(this.paramName);
      }
      // Preserve other filters from the form
      var searchInput = document.querySelector('input[name="search"]');
      if (searchInput && searchInput.value) params.set('search', searchInput.value);
      htmx.ajax('GET', '/partials/requirements?' + params.toString(), { target: '#content', swap: 'innerHTML' });
    }
  };
}

// REQ-DASH-023: Remove a single filter chip
function removeFilterChip(btn) {
  var chip = btn.closest('.filter-chip');
  var param = chip.getAttribute('data-param');
  var value = chip.getAttribute('data-value');

  // Get current param value and remove the specific value
  var params = new URLSearchParams();
  var searchInput = document.querySelector('input[name="search"]');
  if (searchInput && searchInput.value) params.set('search', searchInput.value);

  // Collect current filter values from all chips except the one being removed
  var chips = document.querySelectorAll('.filter-chip');
  var paramValues = {};
  chips.forEach(function(c) {
    var p = c.getAttribute('data-param');
    var v = c.getAttribute('data-value');
    if (p === param && v === value) return; // skip the one being removed
    if (!paramValues[p]) paramValues[p] = [];
    paramValues[p].push(v);
  });
  for (var p in paramValues) {
    params.set(p, paramValues[p].join(','));
  }

  htmx.ajax('GET', '/partials/requirements?' + params.toString(), { target: '#content', swap: 'innerHTML' });
}

// REQ-DASH-023: Clear all filters
function clearAllFilters() {
  var params = new URLSearchParams();
  var searchInput = document.querySelector('input[name="search"]');
  if (searchInput && searchInput.value) params.set('search', searchInput.value);
  htmx.ajax('GET', '/partials/requirements?' + params.toString(), { target: '#content', swap: 'innerHTML' });
}

// REQ-DASH-048: Dagre layered (Sugiyama) layout replacing force simulation
function renderGraph(data) {
  var container = document.getElementById('graph-container');
  var svg = d3.select('#graph-svg');
  if (!container || !svg.node()) return;

  svg.selectAll('*').remove();

  var statusColor = {
    'COMPLETE': '#16a34a',
    'PARTIAL': '#d97706',
    'MISSING': '#dc2626',
    'NOT_STARTED': '#71717a'
  };

  var nodes = (data.nodes || []);
  var edges = (data.edges || []);

  if (nodes.length === 0) return;

  // Create dagre graph
  var g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: 'TB', ranksep: 80, nodesep: 40, marginx: 20, marginy: 20 });
  g.setDefaultEdgeLabel(function() { return {}; });

  // Add nodes with dimensions
  var nodeMap = {};
  nodes.forEach(function(n) {
    nodeMap[n.id] = n;
    g.setNode(n.id, { label: n.id, width: 120, height: 32 });
  });

  // Add edges
  edges.forEach(function(e) {
    g.setEdge(e.source, e.target);
  });

  // Compute layout
  dagre.layout(g);

  // Calculate bounding box and auto-size viewBox
  var padding = 40;
  var minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  g.nodes().forEach(function(id) {
    var n = g.node(id);
    if (n) {
      minX = Math.min(minX, n.x - n.width / 2);
      minY = Math.min(minY, n.y - n.height / 2);
      maxX = Math.max(maxX, n.x + n.width / 2);
      maxY = Math.max(maxY, n.y + n.height / 2);
    }
  });
  var vbW = (maxX - minX) + padding * 2;
  var vbH = (maxY - minY) + padding * 2;
  svg.attr('viewBox', [(minX - padding), (minY - padding), vbW, vbH].join(' '));

  var root = svg.append('g').attr('class', 'graph-root');

  // Zoom and pan (no drag on nodes -- dagre positions are deterministic)
  svg.call(d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', function(event) { root.attr('transform', event.transform); }));

  // Arrow marker
  svg.append('defs').append('marker')
    .attr('id', 'arrow')
    .attr('viewBox', '0 -3 6 6')
    .attr('refX', 6)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-3L6,0L0,3')
    .attr('fill', '#9ca3af');

  // Draw edges as lines (straight from source to target center)
  var linkGroup = root.append('g').attr('class', 'edges');
  edges.forEach(function(e) {
    var srcNode = g.node(e.source);
    var tgtNode = g.node(e.target);
    if (!srcNode || !tgtNode) return;
    linkGroup.append('line')
      .attr('x1', srcNode.x)
      .attr('y1', srcNode.y + srcNode.height / 2)
      .attr('x2', tgtNode.x)
      .attr('y2', tgtNode.y - tgtNode.height / 2)
      .attr('stroke', '#9ca3af')
      .attr('stroke-opacity', 0.4)
      .attr('stroke-width', 1)
      .attr('marker-end', 'url(#arrow)');
  });

  // Draw nodes as circles positioned by dagre
  var nodeGroup = root.append('g').attr('class', 'nodes');
  nodes.forEach(function(n) {
    var pos = g.node(n.id);
    if (!pos) return;
    var ng = nodeGroup.append('g')
      .attr('transform', 'translate(' + pos.x + ',' + pos.y + ')');
    ng.append('circle')
      .attr('r', 6)
      .attr('fill', statusColor[n.status] || '#71717a')
      .attr('stroke', '#fff')
      .attr('stroke-width', 1.5);
    ng.append('title')
      .text(n.id + ' (' + n.group + ')');
    ng.on('mouseover', function() {
      d3.select(this).select('circle').attr('r', 9);
    }).on('mouseout', function() {
      d3.select(this).select('circle').attr('r', 6);
    });
  });
}

// REQ-DASH-091: Reusable tooltip component
function tooltip() {
  return {
    visible: false,
    content: [],
    x: 0,
    y: 0,
    show: function(event, items) {
      this.content = items || [];
      this.x = event.clientX + 12;
      this.y = event.clientY + 12;
      // Viewport-aware repositioning
      var vw = window.innerWidth;
      var vh = window.innerHeight;
      if (this.x + 320 > vw) this.x = event.clientX - 320 - 12;
      if (this.y + 200 > vh) this.y = event.clientY - 200 - 12;
      this.visible = true;
    },
    hide: function() {
      this.visible = false;
    }
  };
}

// REQ-DASH-066: View switcher Alpine component
function viewSwitcher() {
  var params = new URLSearchParams(window.location.search);
  return {
    active: params.get('view') || 'dag',
    switchView: function(view) {
      this.active = view;
      var url = new URL(window.location);
      url.searchParams.set('view', view);
      history.replaceState(null, '', url);
    }
  };
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

  // REQ-DASH-028/029: Auto-open panel when htmx loads content into detail-panel-body
  document.body.addEventListener('htmx:afterSwap', function(event) {
    if (event.detail.target && event.detail.target.id === 'detail-panel-body') {
      var app = document.body.__x;
      if (app && app.$data) {
        app.$data.panelOpen = true;
        // REQ-DASH-029: Push URL state for panel
        var reqIdEl = event.detail.target.querySelector('.req-id');
        if (reqIdEl) {
          var reqId = reqIdEl.textContent.trim();
          if (reqId) {
            history.pushState({ panel: reqId }, '', '/requirements/' + reqId);
          }
        }
      }
    }
  });

  // REQ-DASH-015: Listen for showToast events from HX-Trigger headers
  document.body.addEventListener('showToast', function(event) {
    var detail = event.detail || {};
    showToast(detail.type || 'info', detail.message || '');
  });
});

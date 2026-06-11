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

// REQ-DASH-050: Effort-based radius computation
function effortRadius(effortWeeks) {
  if (!effortWeeks || effortWeeks <= 0) return 12;
  if (effortWeeks < 0.25) return 8;
  if (effortWeeks > 4.0) return 28;
  return Math.round(8 + (effortWeeks - 0.25) * (20 / 3.75));
}

// REQ-DASH-051: Status to CSS class mapping
function statusClass(status) {
  var map = { 'COMPLETE': 'node-complete', 'PARTIAL': 'node-partial', 'MISSING': 'node-missing', 'NOT_STARTED': 'node-not-started' };
  return map[status] || 'node-not-started';
}

// REQ-DASH-052: Priority to CSS class mapping
function priorityClass(priority) {
  var map = { 'P0': 'priority-p0', 'HIGH': 'priority-high', 'MEDIUM': 'priority-medium', 'LOW': 'priority-low' };
  return map[priority] || 'priority-low';
}

// REQ-DASH-083: Build SVG path from dagre edge points with rounded corners
function edgePath(points) {
  if (!points || points.length < 2) return '';
  var d = 'M' + points[0].x + ',' + points[0].y;
  var cornerRadius = 5;
  for (var i = 1; i < points.length; i++) {
    if (i < points.length - 1 && cornerRadius > 0) {
      // Compute rounded corner via quadratic bezier at intermediate points
      var prev = points[i - 1];
      var curr = points[i];
      var next = points[i + 1];
      var dx1 = curr.x - prev.x, dy1 = curr.y - prev.y;
      var dx2 = next.x - curr.x, dy2 = next.y - curr.y;
      var len1 = Math.sqrt(dx1 * dx1 + dy1 * dy1);
      var len2 = Math.sqrt(dx2 * dx2 + dy2 * dy2);
      var r = Math.min(cornerRadius, len1 / 2, len2 / 2);
      if (r > 0 && len1 > 0 && len2 > 0) {
        var cx1 = curr.x - (dx1 / len1) * r;
        var cy1 = curr.y - (dy1 / len1) * r;
        var cx2 = curr.x + (dx2 / len2) * r;
        var cy2 = curr.y + (dy2 / len2) * r;
        d += ' L' + cx1 + ',' + cy1;
        d += ' Q' + curr.x + ',' + curr.y + ' ' + cx2 + ',' + cy2;
      } else {
        d += ' L' + curr.x + ',' + curr.y;
      }
    } else {
      d += ' L' + points[i].x + ',' + points[i].y;
    }
  }
  return d;
}

// REQ-DASH-048: Dagre layered (Sugiyama) layout with full visual encoding
function renderGraph(data) {
  var container = document.getElementById('graph-container');
  var svg = d3.select('#graph-svg');
  if (!container || !svg.node()) return;

  svg.selectAll('*').remove();

  var nodes = (data.nodes || []);
  var edges = (data.edges || []);
  var cycles = (data.cycles || []);

  if (nodes.length === 0) return;

  // Build cycle membership lookup for highlight mode
  var cycleMembership = {}; // nodeId -> array of cycle arrays
  cycles.forEach(function(cycle) {
    cycle.forEach(function(id) {
      if (!cycleMembership[id]) cycleMembership[id] = [];
      cycleMembership[id].push(cycle);
    });
  });

  // Create dagre graph
  var g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: 'TB', ranksep: 80, nodesep: 40, marginx: 20, marginy: 20 });
  g.setDefaultEdgeLabel(function() { return {}; });

  // REQ-DASH-050: Add nodes with effort-proportional dimensions
  var nodeMap = {};
  nodes.forEach(function(n) {
    nodeMap[n.id] = n;
    var r = effortRadius(n.effort_weeks);
    // REQ-DASH-052: Account for priority ring stroke width in dagre dimensions
    var strokeW = 0;
    if (n.priority === 'P0') strokeW = 3;
    else if (n.priority === 'HIGH') strokeW = 2;
    else if (n.priority === 'MEDIUM') strokeW = 1;
    // REQ-DASH-056: Account for label height (id + 12px gap)
    var nodeW = 2 * r + 2 * strokeW;
    var nodeH = 2 * r + 2 * strokeW + 24;
    g.setNode(n.id, { label: n.id, width: Math.max(nodeW, 80), height: nodeH });
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

  // REQ-DASH-084: Zoom and pan on SVG group, no node drag
  var zoomBehavior = d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', function(event) {
      root.attr('transform', event.transform);
      // REQ-DASH-056: Toggle label detail based on zoom level
      var scale = event.transform.k;
      svg.classed('labels-expanded', scale > 1.5);
      svg.classed('labels-hidden', scale < 0.5);
    });
  svg.call(zoomBehavior);
  // REQ-DASH-084: Double-click to reset zoom
  svg.on('dblclick.zoom', function() {
    svg.transition().duration(500).call(zoomBehavior.transform, d3.zoomIdentity);
  });

  // REQ-DASH-053: Arrow markers for normal and critical edges
  var defs = svg.append('defs');
  defs.append('marker')
    .attr('id', 'arrow-normal')
    .attr('class', 'marker-normal')
    .attr('viewBox', '0 -3 6 6')
    .attr('refX', 6).attr('refY', 0)
    .attr('markerWidth', 6).attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path').attr('d', 'M0,-3L6,0L0,3');
  defs.append('marker')
    .attr('id', 'arrow-critical')
    .attr('class', 'marker-critical')
    .attr('viewBox', '0 -3 6 6')
    .attr('refX', 6).attr('refY', 0)
    .attr('markerWidth', 6).attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path').attr('d', 'M0,-3L6,0L0,3');

  // REQ-DASH-053: Draw normal edges first, then critical path edges on top
  var linkGroup = root.append('g').attr('class', 'edges');
  var normalEdges = edges.filter(function(e) { return !e.is_critical_path; });
  var criticalEdges = edges.filter(function(e) { return e.is_critical_path; });

  // REQ-DASH-083: Render edges as SVG paths using dagre edge points
  function renderEdge(e, isCritical) {
    var edgeData = g.edge(e.source, e.target);
    if (!edgeData) return;
    var points = edgeData.points;
    if (!points || points.length < 2) return;
    var pathD = edgePath(points);
    linkGroup.append('path')
      .attr('d', pathD)
      .attr('fill', 'none')
      .attr('class', isCritical ? 'edge-critical' : 'edge-normal')
      .attr('marker-end', isCritical ? 'url(#arrow-critical)' : 'url(#arrow-normal)')
      .attr('data-source', e.source)
      .attr('data-target', e.target);
  }

  normalEdges.forEach(function(e) { renderEdge(e, false); });
  criticalEdges.forEach(function(e) { renderEdge(e, true); });

  // Draw nodes with visual encoding
  var nodeGroup = root.append('g').attr('class', 'nodes');
  var cycleHighlightActive = false;

  nodes.forEach(function(n) {
    var pos = g.node(n.id);
    if (!pos) return;
    var r = effortRadius(n.effort_weeks);

    // Build CSS classes for the node group
    var classes = ['graph-node', statusClass(n.status), priorityClass(n.priority)];
    if (n.is_blocked) classes.push('node-blocked');
    if (n.is_cycle_member) classes.push('node-cycle');

    var ng = nodeGroup.append('g')
      .attr('transform', 'translate(' + pos.x + ',' + pos.y + ')')
      .attr('class', classes.join(' '))
      .attr('data-id', n.id);

    // REQ-DASH-055: Cycle member glow effect (rendered behind node)
    if (n.is_cycle_member) {
      ng.append('circle')
        .attr('class', 'cycle-glow')
        .attr('r', r + 6)
        .attr('fill', 'none')
        .attr('stroke', '#f59e0b')
        .attr('stroke-width', 3)
        .attr('stroke-opacity', 0.5);
    }

    // REQ-DASH-050/051/052: Node circle with effort radius, status fill, priority ring
    ng.append('circle')
      .attr('class', 'node-circle')
      .attr('r', r);

    // REQ-DASH-055: Warning triangle badge for cycle members
    if (n.is_cycle_member) {
      var badgeSize = Math.max(8, r * 0.5);
      var bx = r * 0.7;
      var by = -r * 0.7;
      var tri = ng.append('g')
        .attr('class', 'cycle-badge')
        .attr('transform', 'translate(' + bx + ',' + by + ')');
      tri.append('polygon')
        .attr('points', function() {
          var h = badgeSize;
          var w = badgeSize;
          return '0,' + (-h/2) + ' ' + (w/2) + ',' + (h/2) + ' ' + (-w/2) + ',' + (h/2);
        });
      tri.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '0.35em')
        .attr('font-size', Math.max(6, badgeSize * 0.7) + 'px')
        .attr('fill', '#0a0a0a')
        .attr('font-weight', 'bold')
        .text('!');
    }

    // REQ-DASH-056: Always-visible node label (requirement ID)
    ng.append('text')
      .attr('class', 'node-label-id')
      .attr('y', r + 12)
      .text(n.id);

    // REQ-DASH-056: Zoom-expanded description label
    if (n.description) {
      var desc = n.description.length > 40 ? n.description.substring(0, 40) + '...' : n.description;
      ng.append('text')
        .attr('class', 'node-label-desc')
        .attr('y', r + 23)
        .text(desc);
    }

    // REQ-DASH-084: Hover interaction (scale enlargement, no drag)
    ng.on('mouseover', function() {
      d3.select(this).select('.node-circle')
        .transition().duration(100)
        .attr('r', r * 1.1);
    }).on('mouseout', function() {
      d3.select(this).select('.node-circle')
        .transition().duration(100)
        .attr('r', r);
    });

    // REQ-DASH-055: Click cycle member to highlight full cycle
    if (n.is_cycle_member) {
      ng.on('click', function() {
        if (cycleHighlightActive) {
          // Restore all nodes and edges
          nodeGroup.selectAll('.graph-node').style('opacity', null);
          linkGroup.selectAll('path').style('opacity', null);
          cycleHighlightActive = false;
          return;
        }
        // Find cycles this node belongs to
        var memberCycles = cycleMembership[n.id] || [];
        var highlightSet = {};
        memberCycles.forEach(function(cycle) {
          cycle.forEach(function(id) { highlightSet[id] = true; });
        });
        // Dim non-cycle nodes and edges
        nodeGroup.selectAll('.graph-node').style('opacity', function() {
          return highlightSet[d3.select(this).attr('data-id')] ? 1.0 : 0.3;
        });
        linkGroup.selectAll('path').style('opacity', function() {
          var src = d3.select(this).attr('data-source');
          var tgt = d3.select(this).attr('data-target');
          return (highlightSet[src] && highlightSet[tgt]) ? 1.0 : 0.3;
        });
        cycleHighlightActive = true;
      });
    }
  });

  // REQ-DASH-059: Hover tooltip using shared tooltip DOM element
  var tooltipEl = document.getElementById('graph-tooltip');
  var hoverTimer = null;
  function showGraphTooltip(event, n) {
    if (!tooltipEl) return;
    var desc = n.description || '';
    if (desc.length > 80) desc = desc.substring(0, 80) + '...';
    var assignee = n.assignee || 'Unassigned';
    tooltipEl.className = 'graph-tooltip-card';
    tooltipEl.innerHTML = '<div class="tooltip-id">' + n.id + '</div>' +
      '<div class="tooltip-desc">' + desc + '</div>' +
      '<div class="tooltip-meta">' +
        '<span class="status-' + statusClass(n.status).replace('node-', '') + '">' + n.status + '</span>' +
        '<span>' + (n.priority || 'LOW') + '</span>' +
        '<span>' + (n.effort_weeks || 0) + 'w</span>' +
        '<span>' + assignee + '</span>' +
      '</div>';
    var x = event.clientX + 12;
    var y = event.clientY + 12;
    if (x + 320 > window.innerWidth) x = event.clientX - 320 - 12;
    if (y + 100 > window.innerHeight) y = event.clientY - 100 - 12;
    tooltipEl.style.left = x + 'px';
    tooltipEl.style.top = y + 'px';
    tooltipEl.style.display = 'block';
  }
  function hideGraphTooltip() {
    if (tooltipEl) tooltipEl.style.display = 'none';
    if (hoverTimer) { clearTimeout(hoverTimer); hoverTimer = null; }
  }

  // REQ-DASH-058/059/062: Add interaction events to each node
  var selectedNodeId = null;
  var highlightActive = false;
  nodeGroup.selectAll('.graph-node')
    .on('mouseover', function(event) {
      var el = d3.select(this);
      var nid = el.attr('data-id');
      var nd = nodeMap[nid];
      if (!nd) return;
      var r = effortRadius(nd.effort_weeks);
      el.select('.node-circle').transition().duration(100).attr('r', r * 1.1);
      hoverTimer = setTimeout(function() { showGraphTooltip(event, nd); }, 200);
    })
    .on('mousemove', function(event) {
      var nid = d3.select(this).attr('data-id');
      var nd = nodeMap[nid];
      if (nd && tooltipEl && tooltipEl.style.display === 'block') showGraphTooltip(event, nd);
    })
    .on('mouseout', function() {
      var nid = d3.select(this).attr('data-id');
      var nd = nodeMap[nid];
      var r = nd ? effortRadius(nd.effort_weeks) : 6;
      d3.select(this).select('.node-circle').transition().duration(100).attr('r', r);
      hideGraphTooltip();
    })
    .on('click', function(event) {
      event.stopPropagation();
      hideGraphTooltip();
      var nid = d3.select(this).attr('data-id');

      // REQ-DASH-058: Toggle detail panel on click
      if (selectedNodeId === nid) {
        // Toggle off
        nodeGroup.selectAll('.graph-node').classed('node-selected', false);
        selectedNodeId = null;
        var app = document.querySelector('[x-data]');
        if (app && app.__x) app.__x.$data.panelOpen = false;
      } else {
        nodeGroup.selectAll('.graph-node').classed('node-selected', false);
        d3.select(this).classed('node-selected', true);
        selectedNodeId = nid;
        htmx.ajax('GET', '/partials/detail/' + nid, { target: '#detail-panel-body', swap: 'innerHTML' });
      }

      // REQ-DASH-062: Dependency chain highlight
      if (highlightActive && selectedNodeId === null) {
        // Clear highlight
        nodeGroup.selectAll('.graph-node')
          .classed('highlight-upstream', false)
          .classed('highlight-downstream', false)
          .classed('highlight-selected', false)
          .style('opacity', null);
        linkGroup.selectAll('path').style('opacity', null);
        highlightActive = false;
        return;
      }
      if (selectedNodeId) {
        // Build upstream and downstream sets from edges
        var upstreamSet = {};
        var downstreamSet = {};
        function findUpstream(id) {
          edges.forEach(function(e) {
            if (e.target === id && !upstreamSet[e.source]) {
              upstreamSet[e.source] = true;
              findUpstream(e.source);
            }
          });
        }
        function findDownstream(id) {
          edges.forEach(function(e) {
            if (e.source === id && !downstreamSet[e.target]) {
              downstreamSet[e.target] = true;
              findDownstream(e.target);
            }
          });
        }
        findUpstream(nid);
        findDownstream(nid);
        var chainSet = {};
        chainSet[nid] = true;
        for (var k in upstreamSet) chainSet[k] = true;
        for (var k2 in downstreamSet) chainSet[k2] = true;

        nodeGroup.selectAll('.graph-node')
          .classed('highlight-upstream', function() { return upstreamSet[d3.select(this).attr('data-id')] || false; })
          .classed('highlight-downstream', function() { return downstreamSet[d3.select(this).attr('data-id')] || false; })
          .classed('highlight-selected', function() { return d3.select(this).attr('data-id') === nid; })
          .style('opacity', function() {
            return chainSet[d3.select(this).attr('data-id')] ? 1.0 : 0.3;
          });
        linkGroup.selectAll('path').style('opacity', function() {
          var src = d3.select(this).attr('data-source');
          var tgt = d3.select(this).attr('data-target');
          return (chainSet[src] && chainSet[tgt]) ? 1.0 : 0.3;
        });
        highlightActive = true;
      }
    });

  // REQ-DASH-055/062: Click background to exit highlight modes
  svg.on('click', function(event) {
    if (event.target === svg.node()) {
      if (cycleHighlightActive) {
        nodeGroup.selectAll('.graph-node').style('opacity', null);
        linkGroup.selectAll('path').style('opacity', null);
        cycleHighlightActive = false;
      }
      if (highlightActive) {
        nodeGroup.selectAll('.graph-node')
          .classed('highlight-upstream', false)
          .classed('highlight-downstream', false)
          .classed('highlight-selected', false)
          .style('opacity', null);
        linkGroup.selectAll('path').style('opacity', null);
        highlightActive = false;
      }
      if (selectedNodeId) {
        nodeGroup.selectAll('.graph-node').classed('node-selected', false);
        selectedNodeId = null;
      }
      hideGraphTooltip();
    }
  });

  // REQ-DASH-062: Escape clears highlight
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && (highlightActive || selectedNodeId)) {
      nodeGroup.selectAll('.graph-node')
        .classed('highlight-upstream', false)
        .classed('highlight-downstream', false)
        .classed('highlight-selected', false)
        .classed('node-selected', false)
        .style('opacity', null);
      linkGroup.selectAll('path').style('opacity', null);
      highlightActive = false;
      selectedNodeId = null;
    }
  });

  // REQ-DASH-064: Expose zoom controls globally
  window._graphZoomBehavior = zoomBehavior;
  window._graphSvg = svg;
  window._graphRoot = root;
  window._graphG = g;
  window._graphNodes = nodes;
  window._graphNodeGroup = nodeGroup;
  window._graphLinkGroup = linkGroup;
  window._graphEdges = edges;
  window._graphNodeMap = nodeMap;

  // Update zoom level indicator
  function updateZoomIndicator(k) {
    var el = document.getElementById('zoom-level');
    if (el) el.textContent = Math.round(k * 100) + '%';
  }
  svg.on('zoom.indicator', null);
  zoomBehavior.on('zoom.indicator', function(event) {
    updateZoomIndicator(event.transform.k);
  });
}

// REQ-DASH-064: Zoom control functions
function graphZoom(action) {
  var svg = window._graphSvg;
  var zb = window._graphZoomBehavior;
  var g = window._graphG;
  if (!svg || !zb) return;

  if (action === 'in') {
    svg.transition().duration(200).call(zb.scaleBy, 1.25);
  } else if (action === 'out') {
    svg.transition().duration(200).call(zb.scaleBy, 0.8);
  } else if (action === 'fit') {
    // Fit all nodes into view with 40px padding
    var container = document.getElementById('graph-container');
    if (!container || !g) return;
    var cw = container.clientWidth;
    var ch = container.clientHeight;
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
    var gw = maxX - minX + padding * 2;
    var gh = maxY - minY + padding * 2;
    var scale = Math.min(cw / gw, ch / gh, 4);
    scale = Math.max(scale, 0.1);
    var tx = (cw - gw * scale) / 2 - (minX - padding) * scale;
    var ty = (ch - gh * scale) / 2 - (minY - padding) * scale;
    svg.transition().duration(200).call(zb.transform, d3.zoomIdentity.translate(tx, ty).scale(scale));
  }
}

// REQ-DASH-064: Keyboard shortcuts for zoom (+ - 0)
document.addEventListener('keydown', function(e) {
  var tag = document.activeElement && document.activeElement.tagName;
  if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') return;
  if (e.key === '+' || e.key === '=') { e.preventDefault(); graphZoom('in'); }
  else if (e.key === '-') { e.preventDefault(); graphZoom('out'); }
  else if (e.key === '0') { e.preventDefault(); graphZoom('fit'); }
});

// REQ-DASH-065: Graph search-to-highlight
function graphSearch(query) {
  var nodeGroup = window._graphNodeGroup;
  var linkGroup = window._graphLinkGroup;
  var nodes = window._graphNodes;
  var nodeMap = window._graphNodeMap;
  if (!nodeGroup || !nodes) return;

  var countEl = document.getElementById('graph-search-count');
  if (!query || query.trim() === '') {
    nodeGroup.selectAll('.graph-node').classed('search-match', false).classed('search-dim', false);
    if (countEl) countEl.textContent = '';
    return;
  }

  var q = query.toLowerCase();
  var matchCount = 0;
  var firstMatch = null;
  nodeGroup.selectAll('.graph-node').each(function() {
    var el = d3.select(this);
    var nid = el.attr('data-id');
    var nd = nodeMap[nid];
    var text = (nid + ' ' + (nd ? nd.description || '' : '')).toLowerCase();
    var isMatch = text.indexOf(q) >= 0;
    el.classed('search-match', isMatch).classed('search-dim', !isMatch);
    if (isMatch) {
      matchCount++;
      if (!firstMatch) firstMatch = nid;
    }
  });

  if (countEl) countEl.textContent = matchCount + ' match' + (matchCount !== 1 ? 'es' : '');

  // Pan to first match
  if (firstMatch && window._graphG && window._graphSvg && window._graphZoomBehavior) {
    var pos = window._graphG.node(firstMatch);
    if (pos) {
      var container = document.getElementById('graph-container');
      if (container) {
        var cw = container.clientWidth;
        var ch = container.clientHeight;
        var tx = cw / 2 - pos.x;
        var ty = ch / 2 - pos.y;
        window._graphSvg.transition().duration(200).call(
          window._graphZoomBehavior.transform,
          d3.zoomIdentity.translate(tx, ty).scale(1)
        );
      }
    }
  }
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

// REQ-DASH-086: Timeline D3 SVG rendering
function renderTimeline(data) {
  var svg = d3.select('#timeline-svg');
  if (!svg.node() || !data) return;
  svg.selectAll('*').remove();

  var statusColor = {
    'COMPLETE': 'var(--rtmx-complete)',
    'PARTIAL': 'var(--rtmx-partial)',
    'MISSING': 'var(--rtmx-missing)',
    'NOT_STARTED': 'var(--rtmx-not-started)'
  };

  var sprints = data.sprints || [];
  var categories = data.categories || [];
  var reqs = data.requirements || [];
  if (sprints.length === 0 || categories.length === 0) return;

  var margin = { top: 40, right: 20, bottom: 20, left: 120 };
  var colWidth = Math.max(120, 800 / sprints.length);
  var rowHeight = 22;

  // Group reqs by category then sprint
  var catSprReqs = {};
  categories.forEach(function(cat) { catSprReqs[cat] = {}; });
  reqs.forEach(function(r) {
    var cat = r.category;
    var spr = r.sprint || 'Unscheduled';
    if (!catSprReqs[cat]) catSprReqs[cat] = {};
    if (!catSprReqs[cat][spr]) catSprReqs[cat][spr] = [];
    catSprReqs[cat][spr].push(r);
  });

  // Calculate heights per category
  var catHeights = {};
  var totalHeight = margin.top;
  categories.forEach(function(cat) {
    var maxRows = 0;
    sprints.forEach(function(spr) {
      var arr = (catSprReqs[cat] || {})[spr] || [];
      maxRows = Math.max(maxRows, arr.length);
    });
    catHeights[cat] = Math.max(1, maxRows) * rowHeight + 10;
    totalHeight += catHeights[cat] + 30;
  });
  totalHeight += margin.bottom;
  var totalWidth = margin.left + sprints.length * colWidth + margin.right;

  svg.attr('width', totalWidth).attr('height', Math.max(400, totalHeight))
     .attr('viewBox', '0 0 ' + totalWidth + ' ' + Math.max(400, totalHeight));

  // Sprint column headers
  sprints.forEach(function(spr, i) {
    svg.append('text')
      .attr('class', 'timeline-sprint-label')
      .attr('x', margin.left + i * colWidth + colWidth / 2)
      .attr('y', 20)
      .text(spr);
  });

  // Render rows per category
  var y = margin.top;
  categories.forEach(function(cat) {
    svg.append('text')
      .attr('class', 'timeline-category-label')
      .attr('x', 8)
      .attr('y', y + 14)
      .text(cat);

    sprints.forEach(function(spr, si) {
      var arr = (catSprReqs[cat] || {})[spr] || [];
      arr.forEach(function(r, ri) {
        var barW = Math.max(20, (r.effort_weeks || 0.5) / 4 * (colWidth - 8));
        barW = Math.min(barW, colWidth - 4);
        var bx = margin.left + si * colWidth + 2;
        var by = y + ri * rowHeight + 2;
        svg.append('rect')
          .attr('class', 'timeline-bar')
          .attr('x', bx)
          .attr('y', by)
          .attr('width', barW)
          .attr('height', rowHeight - 4)
          .attr('rx', 3)
          .attr('fill', statusColor[r.status] || statusColor['NOT_STARTED']);
        if (barW > 40) {
          svg.append('text')
            .attr('class', 'timeline-label')
            .attr('x', bx + 4)
            .attr('y', by + 13)
            .text(r.id);
        }
      });
    });
    y += catHeights[cat] + 30;
  });
}

// REQ-DASH-087: Treemap D3 squarify rendering
function renderTreemap(data) {
  var svgEl = document.getElementById('treemap-svg');
  var svg = d3.select('#treemap-svg');
  if (!svg.node() || !data) return;
  svg.selectAll('*').remove();

  var statusColor = {
    'COMPLETE': 'var(--rtmx-complete)',
    'PARTIAL': 'var(--rtmx-partial)',
    'MISSING': 'var(--rtmx-missing)',
    'NOT_STARTED': 'var(--rtmx-not-started)'
  };

  var container = svgEl.parentElement;
  var w = container ? container.clientWidth : 800;
  var h = Math.max(400, Math.round(w * 9 / 16));
  svg.attr('width', w).attr('height', h).attr('viewBox', '0 0 ' + w + ' ' + h);

  // Build d3 hierarchy
  var root = d3.hierarchy(data)
    .sum(function(d) { return d.effort_weeks || 0; })
    .sort(function(a, b) { return b.value - a.value; });

  d3.treemap()
    .tile(d3.treemapSquarify)
    .size([w, h])
    .padding(2)
    .paddingTop(18)
    (root);

  // Render category groups
  var groups = root.children || [];
  groups.forEach(function(group) {
    svg.append('rect')
      .attr('x', group.x0)
      .attr('y', group.y0)
      .attr('width', group.x1 - group.x0)
      .attr('height', group.y1 - group.y0)
      .attr('fill', 'none')
      .attr('stroke', 'var(--rtmx-border)')
      .attr('stroke-width', 2);
    svg.append('text')
      .attr('class', 'treemap-group-label')
      .attr('x', group.x0 + 4)
      .attr('y', group.y0 + 13)
      .text(group.data.name);

    // Render requirement cells
    var leaves = group.children || [];
    leaves.forEach(function(leaf) {
      var lw = leaf.x1 - leaf.x0;
      var lh = leaf.y1 - leaf.y0;
      var g = svg.append('g').attr('class', 'treemap-cell');
      g.append('rect')
        .attr('x', leaf.x0)
        .attr('y', leaf.y0)
        .attr('width', lw)
        .attr('height', lh)
        .attr('fill', statusColor[leaf.data.status] || statusColor['NOT_STARTED'])
        .attr('rx', 2);
      if (lw > 60 && lh > 20) {
        g.append('text')
          .attr('class', 'treemap-cell-text')
          .attr('x', leaf.x0 + 3)
          .attr('y', leaf.y0 + 13)
          .text(leaf.data.id);
      }
      g.append('title')
        .text(leaf.data.id + ': ' + (leaf.data.title || '') + ' (' + leaf.data.status + ', ' + leaf.data.effort_weeks + 'w)');
      g.on('click', function() {
        htmx.ajax('GET', '/partials/detail/' + leaf.data.id, { target: '#detail-panel-body', swap: 'innerHTML' });
      });
    });
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

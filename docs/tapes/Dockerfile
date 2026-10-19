# Reproducible VHS recordings with generic paths (no host HOME/PWD leakage).
# Expect build context = rtmx/ repo root after `make build` (bin/rtmx present).
FROM ghcr.io/charmbracelet/vhs:latest
USER root
RUN apt-get update && apt-get install -y --no-install-recommends jq ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
COPY bin/rtmx /usr/local/bin/rtmx
COPY docs/tapes/agent-mcp-session /usr/local/bin/agent-mcp-session
COPY docs/tapes/list-rtmx-mcp-tools /usr/local/bin/list-rtmx-mcp-tools
COPY docs/tapes/stubs/gh /usr/local/bin/gh
RUN chmod +x /usr/local/bin/rtmx /usr/local/bin/agent-mcp-session /usr/local/bin/list-rtmx-mcp-tools /usr/local/bin/gh

ENV HOME=/home/demo \
	USER=demo
RUN useradd -m -d /home/demo -s /bin/bash demo \
	&& mkdir -p /workspace/.cursor /workspace/.rtmx /workspace/docs/assets /workspace/docs/tapes \
	&& printf '%s\n' 'export PS1="> "' >> /home/demo/.bashrc \
	&& chown -R demo:demo /home/demo /workspace

WORKDIR /workspace
USER demo

COPY --chown=demo:demo docs/tapes/fixtures/mcp.json /workspace/.cursor/mcp.json
COPY --chown=demo:demo docs/tapes/fixtures/database.csv /workspace/.rtmx/database.csv
COPY --chown=demo:demo docs/tapes/fixtures/rtmx.yaml /workspace/rtmx.yaml
COPY --chown=demo:demo docs/tapes/fixtures/verify-results.json /workspace/results.json
COPY --chown=demo:demo docs/tapes/*.tape /workspace/docs/tapes/

RUN mkdir -p /workspace/internal/demo && printf '%s\n' \
	'package demo' \
	'' \
	'import "testing"' \
	'' \
	'func TestDemo(t *testing.T) {}' \
	> /workspace/internal/demo/demo_test.go

ENTRYPOINT []
CMD ["bash", "-lc", "vhs docs/tapes/workflow.tape && vhs docs/tapes/commands.tape && vhs docs/tapes/agent-loop.tape && vhs docs/tapes/mcp-setup.tape"]

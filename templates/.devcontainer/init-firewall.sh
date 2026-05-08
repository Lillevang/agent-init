#!/usr/bin/env bash
# init-firewall.sh — egress allowlist for agent containers.
#
# Run via sudo from postCreateCommand. Drops all egress except to a curated
# set of registries and the Anthropic/OpenAI APIs. Resolves DNS once at
# startup; if registry IPs rotate, you'll need to re-run this.
#
# This is a starting point. Audit and adjust to your needs.

set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo "Must run as root (sudo)" >&2
    exit 1
fi

# Allowlist of hostnames the agent legitimately needs
ALLOWED_HOSTS=(
    # Agent backends
    "api.anthropic.com"
    "api.openai.com"
    # Package registries
    "registry.npmjs.org"
    "pypi.org"
    "files.pythonhosted.org"
    "deb.debian.org"
    "security.debian.org"
    "deb.nodesource.com"
    # Source hosts
    "github.com"
    "api.github.com"
    "objects.githubusercontent.com"
    "codeload.github.com"
    "raw.githubusercontent.com"
    # Container registries (in case agent builds images)
    "ghcr.io"
    "registry-1.docker.io"
    "auth.docker.io"
    "production.cloudflare.docker.com"
)

echo "→ Configuring egress firewall..."

# Flush existing rules
iptables -F OUTPUT
iptables -P OUTPUT DROP

# Allow loopback
iptables -A OUTPUT -o lo -j ACCEPT

# Allow established/related (return traffic for connections we initiate)
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow DNS (otherwise nothing else works)
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 53 -j ACCEPT

# Resolve and allow each host
for host in "${ALLOWED_HOSTS[@]}"; do
    ips=$(getent ahosts "$host" | awk '{print $1}' | sort -u || true)
    if [[ -z "$ips" ]]; then
        echo "  ⚠  could not resolve $host"
        continue
    fi
    for ip in $ips; do
        iptables -A OUTPUT -d "$ip" -p tcp --dport 443 -j ACCEPT
        iptables -A OUTPUT -d "$ip" -p tcp --dport 80 -j ACCEPT
    done
    echo "  ✓ allowed $host"
done

# Log dropped packets (rate-limited) so we can see what gets blocked
iptables -A OUTPUT -m limit --limit 5/min -j LOG --log-prefix "FIREWALL DROP: " --log-level 4

echo "✓ Firewall configured. Egress restricted to allowlist."

#!/bin/sh
# Block cloud instance metadata and the Azure wireserver from Docker sandboxes.
# Host-originated traffic does not traverse DOCKER-USER, so the host agent can
# still read IMDS. Keep this destination list identical to
# internal/container/metadata_guard.go and docker/docker-host/setup.sh.
set -u

drop_v4="169.254.169.254/32 169.254.169.253/32 169.254.170.2/32 168.63.129.16/32"
drop_v6="fd00:ec2::254/128"

ensure() {
    bin=$1
    dest=$2
    required=$3
    if ! command -v "$bin" >/dev/null 2>&1; then
        echo "rexec-metadata-guard: $bin is not installed" >&2
        if [ "$required" = "yes" ]; then
            return 1
        fi
        return 0
    fi
    i=0
    while [ "$i" -lt 30 ]; do
        if "$bin" -C DOCKER-USER -d "$dest" -j DROP >/dev/null 2>&1; then
            return 0
        fi
        if "$bin" -I DOCKER-USER 1 -d "$dest" -j DROP >/dev/null 2>&1; then
            echo "rexec-metadata-guard: dropped $dest via $bin"
            return 0
        fi
        i=$((i + 1))
        sleep 1
    done
    echo "rexec-metadata-guard: failed to drop $dest via $bin" >&2
    return 1
}

fail=0
for dest in $drop_v4; do
    ensure iptables "$dest" yes || fail=1
done
for dest in $drop_v6; do
    ensure ip6tables "$dest" no || fail=1
done
exit "$fail"

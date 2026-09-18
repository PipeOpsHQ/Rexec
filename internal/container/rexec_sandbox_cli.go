package container

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
)

const inSandboxCLIInstallTimeout = 15 * time.Second

// inSandboxCLIInstallScript writes the in-sandbox `rexec` helper to PATH.
// This is the CLI users run inside a terminal (`rexec tools`, `rexec info`),
// not the host rexec-cli used to create sandboxes.
func inSandboxCLIInstallScript() string {
	return `#!/bin/sh
set +e
mkdir -p /root/.local/bin /home/user/.local/bin /usr/local/bin /usr/bin 2>/dev/null || true

cat > /root/.local/bin/rexec << 'REXECCLI'
#!/bin/sh
VERSION="2.1.0"
CYAN='\033[1;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

detect_pkg_manager() {
    if command -v apt-get >/dev/null 2>&1; then echo "apt"
    elif command -v apk >/dev/null 2>&1; then echo "apk"
    elif command -v dnf >/dev/null 2>&1; then echo "dnf"
    elif command -v yum >/dev/null 2>&1; then echo "yum"
    elif command -v pacman >/dev/null 2>&1; then echo "pacman"
    elif command -v zypper >/dev/null 2>&1; then echo "zypper"
    else echo "unknown"
    fi
}

get_pkg_name() {
    PKG="$1"
    PM="$2"
    case "$PKG" in
        nodejs) echo "nodejs" ;;
        python) [ "$PM" = "pacman" ] && echo "python" || echo "python3" ;;
        pip) [ "$PM" = "apk" ] && echo "py3-pip" || ([ "$PM" = "pacman" ] && echo "python-pip" || echo "python3-pip") ;;
        docker) [ "$PM" = "apt" ] && echo "docker.io" || echo "docker" ;;
        neovim|nvim) echo "neovim" ;;
        ripgrep|rg) echo "ripgrep" ;;
        *) echo "$PKG" ;;
    esac
}

do_install() {
    PKG="$1"
    if [ -z "$PKG" ]; then
        printf "${RED}Error: No package specified${NC}\n"
        echo "Usage: rexec install <package>"
        return 1
    fi
    PM=$(detect_pkg_manager)
    if [ "$PM" = "unknown" ]; then
        printf "${RED}Error: No supported package manager found${NC}\n"
        return 1
    fi
    ACTUAL_PKG=$(get_pkg_name "$PKG" "$PM")
    printf "${CYAN}Installing $PKG ($ACTUAL_PKG)...${NC}\n"
    case "$PM" in
        apt)
            export DEBIAN_FRONTEND=noninteractive
            apt-get update -qq >/dev/null 2>&1
            apt-get install -y "$ACTUAL_PKG"
            ;;
        apk) apk add --no-cache "$ACTUAL_PKG" ;;
        dnf) dnf install -y "$ACTUAL_PKG" ;;
        yum) yum install -y "$ACTUAL_PKG" ;;
        pacman) pacman -Sy --noconfirm "$ACTUAL_PKG" ;;
        zypper) zypper --non-interactive install "$ACTUAL_PKG" ;;
    esac
}

show_tools() {
    export PATH="$HOME/.local/bin:/root/.local/bin:/usr/local/bin:$PATH"
    printf "${CYAN}=== Installed Tools ===${NC}\n"
    echo ""
    printf "${YELLOW}System:${NC}\n"
    for cmd in zsh git curl wget vim nano htop jq tmux fzf ripgrep neofetch; do
        command -v "$cmd" >/dev/null 2>&1 && printf "  ${GREEN}✓${NC} $cmd\n"
    done
    echo ""
    printf "${YELLOW}AI & Dev:${NC}\n"
    for cmd in python3 node go rustc docker kubectl tgpt aichat mods gum aider opencode claude gemini llm; do
        command -v "$cmd" >/dev/null 2>&1 && printf "  ${GREEN}✓${NC} $cmd\n"
    done
    echo ""
}

show_help() {
    printf "${CYAN}rexec${NC} v${VERSION}\n"
    echo "Usage: rexec [command]"
    echo ""
    echo "Commands:"
    echo "  tools           List installed tools"
    echo "  info            System info"
    echo "  install <pkg>   Install a package"
    echo "  search <term>   Search packages"
    echo "  help            Show this help"
}

CMD="$1"
[ $# -gt 0 ] && shift
case "$CMD" in
    tools|ls) show_tools ;;
    info)
        if command -v neofetch >/dev/null 2>&1; then neofetch
        else
            echo "Host: $(hostname)"
            [ -f /etc/os-release ] && grep PRETTY_NAME /etc/os-release | cut -d'"' -f2
        fi
        ;;
    install|i) do_install "$@" ;;
    search|s)
        TERM="$1"
        [ -z "$TERM" ] && echo "Usage: rexec search <term>" && exit 1
        PM=$(detect_pkg_manager)
        case "$PM" in
            apt) apt-cache search "$TERM" | head -20 ;;
            apk) apk search "$TERM" | head -20 ;;
            *) echo "Search not supported on this OS" ;;
        esac
        ;;
    help|--help|-h|"") show_help ;;
    *) echo "Unknown command: $CMD"; show_help ;;
esac
REXECCLI

chmod +x /root/.local/bin/rexec 2>/dev/null || true
cp /root/.local/bin/rexec /usr/local/bin/rexec 2>/dev/null || true
chmod +x /usr/local/bin/rexec 2>/dev/null || true
cp /root/.local/bin/rexec /usr/bin/rexec 2>/dev/null || true
chmod +x /usr/bin/rexec 2>/dev/null || true
cp /root/.local/bin/rexec /home/user/.local/bin/rexec 2>/dev/null || true
chmod +x /home/user/.local/bin/rexec 2>/dev/null || true
chown -R user:user /home/user/.local 2>/dev/null || true

if [ -x /usr/local/bin/rexec ] || [ -x /usr/bin/rexec ]; then
    echo "[[REXEC_STATUS]]rexec CLI ready"
    exit 0
fi
echo "[[REXEC_STATUS]]rexec CLI install failed"
exit 1
`
}

// InstallInSandboxCLI copies the in-sandbox rexec helper onto PATH.
func InstallInSandboxCLI(ctx context.Context, cli client.APIClient, containerID string) error {
	ctx, cancel := context.WithTimeout(ctx, inSandboxCLIInstallTimeout)
	defer cancel()

	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          []string{"/bin/sh", "-c", inSandboxCLIInstallScript()},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("create exec: %w", err)
	}

	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("attach exec: %w", err)
	}
	defer attachResp.Close()
	_, _ = io.Copy(io.Discard, attachResp.Reader)

	inspect, err := cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
	if err != nil {
		return fmt.Errorf("inspect exec: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("rexec helper install exited %d", inspect.ExitCode)
	}
	return nil
}

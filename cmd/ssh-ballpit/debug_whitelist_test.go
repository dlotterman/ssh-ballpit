package main

import (
	"path/filepath"
	"testing"
)

func TestDebugWhitelistExample(t *testing.T) {
	for _, file := range []string{"debug_whitelist.yaml", "debug_whitelist_extended.yaml"} {
		t.Run(file, func(t *testing.T) { checkCommonDebugRules(t, loadDebugRules(t, file)) })
	}
}

func loadDebugRules(t *testing.T, file string) Config {
	t.Helper()
	patterns, err := readCommandPatterns(filepath.Join("..", "..", "examples", file))
	if err != nil {
		t.Fatal(err)
	}
	return Config{commandRules: commandRules{whitelist: patterns}}
}

func checkCommonDebugRules(t *testing.T, config Config) {
	t.Helper()
	for _, command := range []string{
		"uname -a", "df -hT", "du -cs -- /var/log", "ps -axuf",
		"ss -tnap", "ss -N lab_net-1 -tlnp", "ip netns list",
		"ip -n lab_net-1 route show", "systemctl --no-pager --plain status sshd.service",
		"systemctl --user --no-pager --plain list-timers --all",
		"journalctl --no-pager -n 100 -u sshd.service",
	} {
		source, safe, err := localCommandDecision(config, command)
		if source != "whitelist" || !safe || err != nil {
			t.Errorf("example did not allow %q: source=%q safe=%t err=%v", command, source, safe, err)
		}
	}
	for _, command := range []string{
		"df -cs", "ss -K", "ss -N ../other -tlnp", "ss -N lab -K",
		"ip netns exec lab ss -tlnp", "systemctl restart sshd.service",
		"journalctl --no-pager -f -u sshd.service",
		"ss -tnap; touch /tmp/marker", "ss -tnap > /tmp/marker",
	} {
		source, safe, _ := localCommandDecision(config, command)
		if source == "whitelist" || safe {
			t.Errorf("example bypassed review for %q", command)
		}
	}
}

func TestExtendedDebugRules(t *testing.T) {
	config := loadDebugRules(t, "debug_whitelist_extended.yaml")
	for _, command := range []string{
		"systemctl --system --no-pager --plain is-system-running",
		"systemctl --system --no-pager --plain status user@1000.service",
		"systemctl --user show --property=Paths example.path",
		"systemctl is-enabled local-fs.target sshd.service",
		"systemd-analyze --system --no-pager verify /etc/systemd/system/example.service",
		"journalctl --system --no-pager --output=short-iso -b -1 -u example.service -n 100",
		"journalctl -k -b -1 --since \"2026-10-24 01:00:00\" --until \"2026-10-24 02:00:00\" --no-pager -o short-iso",
		"smartctl -d sat -n standby,3 -i /dev/sdb",
		"cat /sys/fs/xfs/dm-3/error/metadata/EIO/max_retries",
		"cat /sys/bus/usb/devices/3-2.1/power/runtime_status",
		"cat /sys/class/drm/card1/device/gpu_busy_percent",
		"lsusb -t", "findmnt -rn -o SOURCE,TARGET,FSTYPE", "ls -l /var/crash",
		"df -h /", "sysctl net.ipv4.ip_forward", "podman ps",
	} {
		source, safe, err := localCommandDecision(config, command)
		if source != "whitelist" || !safe || err != nil {
			t.Errorf("expected whitelist match for %q, got %q (%v)", command, source, err)
		}
	}
	for _, command := range []string{
		"systemctl --system --no-pager --plain status --all",
		"systemctl is-enabled sshd.service --root=/tmp",
		"systemctl --user show --property=Environment example.service",
		"systemctl --system --no-pager --plain restart sshd.service",
		"journalctl --rotate", "smartctl -t long /dev/sdb",
		"sysctl net.ipv4.ip_forward=1", "podman exec app id",
		"cat /sys/class/drm/card1/device/config", "ls -al /home/../etc",
		"cat /run/example/rclone.conf", "rclone listremotes --config /run/example/rclone.conf",
		"tcpdump -i any", "lsusb -t && id", "lsusb -t > /tmp/output",
	} {
		source, safe, _ := localCommandDecision(config, command)
		if source == "whitelist" || safe {
			t.Errorf("unexpected whitelist match for %q", command)
		}
	}
}

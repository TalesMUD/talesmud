package gamemode

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a YAML time.Duration. "30s" and "10m" are accepted.
type Duration time.Duration

// Duration returns the standard library value.
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// UnmarshalYAML accepts a duration string.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	var text string
	if err := value.Decode(&text); err != nil {
		var n int64
		if err2 := value.Decode(&n); err2 != nil {
			return err
		}
		*d = Duration(n)
		return nil
	}
	text = strings.TrimSpace(text)
	if text == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(text)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// SSHBanConfig is the failed-auth ban window.
type SSHBanConfig struct {
	Failures int      `yaml:"failures"`
	Window   Duration `yaml:"window"`
	Ban      Duration `yaml:"ban"`
}

// SSHKeysConfig controls linked public keys.
type SSHKeysConfig struct {
	Enabled       bool  `yaml:"enabled"`
	MaxPerAccount int   `yaml:"max_per_account"`
	WebManage     *bool `yaml:"web_manage"`
}

// WebManageOn reports whether the web profile may add and revoke keys.
// An omitted value is on.
func (c SSHKeysConfig) WebManageOn() bool {
	if c.WebManage == nil {
		return true
	}
	return *c.WebManage
}

// SSHDeviceConfig is the device-code login.
type SSHDeviceConfig struct {
	Enabled              bool     `yaml:"enabled"`
	TTL                  Duration `yaml:"ttl"`
	ActivateURL          string   `yaml:"activate_url"`
	MaxPendingPerIP      int      `yaml:"max_pending_per_ip"`
	LookupsPerUserPer10m int      `yaml:"lookups_per_user_per_10m"`
	LookupsPerIPPer10m   int      `yaml:"lookups_per_ip_per_10m"`
}

// SSHGuestConfig is the extra SSH guest cap. It does not turn guests on by itself.
type SSHGuestConfig struct {
	Enabled       bool     `yaml:"enabled"`
	MaxConcurrent int      `yaml:"max_concurrent"`
	PerIPPerHour  int      `yaml:"per_ip_per_hour"`
	MaxSession    Duration `yaml:"max_session"`
}

// SSHDoorConfig is presentation for the text client over SSH.
type SSHDoorConfig struct {
	Splash         string `yaml:"splash"`
	ActivateScreen string `yaml:"activate_screen"`
	CharsetDefault string `yaml:"charset_default"`
	LetterboxFill  string `yaml:"letterbox_fill"`
}

// SSHMudConfig is presentation for classic play over SSH.
type SSHMudConfig struct {
	Splash  string `yaml:"splash"`
	History int    `yaml:"history"`
}

// SSHConfig is the process SSH listener. The zero value is off.
type SSHConfig struct {
	Enabled             bool            `yaml:"enabled"`
	Listen              string          `yaml:"listen"`
	HostKeyPath         string          `yaml:"host_key_path"`
	PublicHost          string          `yaml:"public_host"`
	PublicPort          int             `yaml:"public_port"`
	MaxConnections      int             `yaml:"max_connections"`
	MaxPerIP            int             `yaml:"max_per_ip"`
	NewConnsPerIPPerMin int             `yaml:"new_conns_per_ip_per_min"`
	AuthFailBan         SSHBanConfig    `yaml:"auth_fail_ban"`
	AuthTimeout         Duration        `yaml:"auth_timeout"`
	IdleTimeout         Duration        `yaml:"idle_timeout"`
	MaxSession          Duration        `yaml:"max_session"`
	Keys                SSHKeysConfig   `yaml:"keys"`
	Device              SSHDeviceConfig `yaml:"device"`
	Guest               SSHGuestConfig  `yaml:"guest"`
	Door                SSHDoorConfig   `yaml:"door"`
	Mud                 SSHMudConfig    `yaml:"mud"`
}

// GuestsConfig is shared guest policy. PersistentEffects defaults off, so a
// guest does not write long-lived shared records.
type GuestsConfig struct {
	PersistentEffects bool `yaml:"persistent_effects"`
}

// SSH returns the loaded SSH settings.
func SSH() SSHConfig { return current.SSH }

// GuestEffectsAllowed reports whether a guest may write long-lived shared
// records. A real account always may.
func GuestEffectsAllowed(isGuest bool) bool {
	if !isGuest {
		return true
	}
	return current.Guests.PersistentEffects
}

func normalizeSSH(cfg SSHConfig) SSHConfig {
	cfg.Listen = strings.TrimSpace(cfg.Listen)
	cfg.HostKeyPath = strings.TrimSpace(cfg.HostKeyPath)
	cfg.PublicHost = strings.TrimSpace(cfg.PublicHost)
	cfg.Device.ActivateURL = strings.TrimSpace(cfg.Device.ActivateURL)
	cfg.Door.Splash = strings.TrimSpace(cfg.Door.Splash)
	cfg.Door.ActivateScreen = strings.TrimSpace(cfg.Door.ActivateScreen)
	cfg.Door.CharsetDefault = strings.TrimSpace(cfg.Door.CharsetDefault)
	if cfg.Door.CharsetDefault == "" {
		cfg.Door.CharsetDefault = "auto"
	}
	cfg.Mud.Splash = strings.TrimSpace(cfg.Mud.Splash)
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 100
	}
	if cfg.MaxPerIP <= 0 {
		cfg.MaxPerIP = 5
	}
	if cfg.NewConnsPerIPPerMin <= 0 {
		cfg.NewConnsPerIPPerMin = 20
	}
	if cfg.AuthFailBan.Failures <= 0 {
		cfg.AuthFailBan.Failures = 10
	}
	if cfg.AuthFailBan.Window.Duration() <= 0 {
		cfg.AuthFailBan.Window = Duration(10 * time.Minute)
	}
	if cfg.AuthFailBan.Ban.Duration() <= 0 {
		cfg.AuthFailBan.Ban = Duration(15 * time.Minute)
	}
	if cfg.AuthTimeout.Duration() <= 0 {
		cfg.AuthTimeout = Duration(30 * time.Second)
	}
	if cfg.IdleTimeout.Duration() <= 0 {
		cfg.IdleTimeout = Duration(30 * time.Minute)
	}
	if cfg.Keys.MaxPerAccount <= 0 {
		cfg.Keys.MaxPerAccount = 10
	}
	if cfg.Device.TTL.Duration() <= 0 {
		cfg.Device.TTL = Duration(10 * time.Minute)
	}
	if cfg.Device.MaxPendingPerIP <= 0 {
		cfg.Device.MaxPendingPerIP = 3
	}
	if cfg.Device.LookupsPerUserPer10m <= 0 {
		cfg.Device.LookupsPerUserPer10m = 10
	}
	if cfg.Device.LookupsPerIPPer10m <= 0 {
		cfg.Device.LookupsPerIPPer10m = 20
	}
	if cfg.Guest.MaxConcurrent <= 0 {
		cfg.Guest.MaxConcurrent = 10
	}
	if cfg.Guest.PerIPPerHour <= 0 {
		cfg.Guest.PerIPPerHour = 5
	}
	if cfg.Guest.MaxSession.Duration() <= 0 {
		cfg.Guest.MaxSession = Duration(30 * time.Minute)
	}
	if cfg.Mud.History <= 0 {
		cfg.Mud.History = 20
	}
	return cfg
}

func applySSHEnv(cfg SSHConfig) SSHConfig {
	if v, ok := envBool("SSH_ENABLED"); ok {
		cfg.Enabled = v
	}
	if v := strings.TrimSpace(os.Getenv("SSH_LISTEN")); v != "" {
		cfg.Listen = v
	}
	if v := strings.TrimSpace(os.Getenv("SSH_HOST_KEY_PATH")); v != "" {
		cfg.HostKeyPath = v
	}
	if v := strings.TrimSpace(os.Getenv("SSH_PUBLIC_HOST")); v != "" {
		cfg.PublicHost = v
	}
	if v := strings.TrimSpace(os.Getenv("SSH_PUBLIC_PORT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.PublicPort = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("SSH_ACTIVATE_URL")); v != "" {
		cfg.Device.ActivateURL = v
	}
	if v, ok := envBool("SSH_GUEST_ENABLED"); ok {
		cfg.Guest.Enabled = v
	}
	if v, ok := envBool("SSH_DEVICE_ENABLED"); ok {
		cfg.Device.Enabled = v
	}
	return cfg
}

func envBool(key string) (bool, bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return false, false
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

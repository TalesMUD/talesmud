package contenthealth

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// PackRule is one content-pack check. when is shorthand for eq on each field.
type PackRule struct {
	ID       string                 `yaml:"id" json:"id"`
	Title    string                 `yaml:"title,omitempty" json:"title,omitempty"`
	Severity string                 `yaml:"severity" json:"severity"`
	Entity   string                 `yaml:"entity" json:"entity"`
	When     map[string]interface{} `yaml:"when,omitempty" json:"when,omitempty"`
	Require  []Condition            `yaml:"require" json:"require"`
	Message  string                 `yaml:"message" json:"message"`
}

// Condition is one field operator. Exactly one operator is set.
type Condition struct {
	Field string
	Op    string
	Value interface{}
}

// UnmarshalYAML reads field plus one of eq, ne, gte, lte, in, exists, regex.
func (c *Condition) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	field, _ := raw["field"].(string)
	c.Field = field
	for _, op := range []string{"eq", "ne", "gte", "lte", "in", "exists", "regex"} {
		if v, ok := raw[op]; ok {
			c.Op = op
			c.Value = v
			return nil
		}
	}
	return fmt.Errorf("condition on %q has no operator", c.Field)
}

type packFile struct {
	Rules []PackRule `yaml:"rules"`
}

// LoadPackRules reads every YAML file in dir. A missing directory is no rules.
func LoadPackRules(dir string) ([]PackRule, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	var rules []PackRule
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		var file packFile
		if err := yaml.Unmarshal(raw, &file); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		for _, rule := range file.Rules {
			if err := validatePackRule(rule); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

func validatePackRule(rule PackRule) error {
	if rule.ID == "" {
		return fmt.Errorf("pack rule is missing an id")
	}
	for _, cond := range rule.Require {
		if cond.Op == "regex" {
			if _, err := regexp.Compile(fmt.Sprint(cond.Value)); err != nil {
				return fmt.Errorf("rule %s: %w", rule.ID, err)
			}
		}
	}
	return nil
}

func packHit(rule PackRule, entity any, entityType, id, name string) (Hit, bool) {
	raw := asMap(entity)
	if !whenMatches(rule.When, raw) {
		return Hit{}, false
	}
	var failed *Condition
	for i := range rule.Require {
		if !conditionHolds(rule.Require[i], raw) {
			failed = &rule.Require[i]
			break
		}
	}
	if failed == nil {
		return Hit{}, false
	}
	severity := rule.Severity
	if severity == "" {
		severity = "error"
	}
	message := rule.Message
	if message == "" {
		message = rule.ID
	}
	return Hit{
		RuleID:     rule.ID,
		Severity:   severity,
		EntityType: entityType,
		EntityID:   id,
		EntityName: name,
		Field:      failed.Field,
		Message:    message,
	}, true
}

func whenMatches(when map[string]interface{}, raw map[string]interface{}) bool {
	for field, want := range when {
		got, ok := lookup(raw, field)
		if !ok || !valuesEqual(got, want) {
			return false
		}
	}
	return true
}

func conditionHolds(c Condition, raw map[string]interface{}) bool {
	got, ok := lookup(raw, c.Field)
	switch c.Op {
	case "eq":
		if !ok {
			return false
		}
		return valuesEqual(got, c.Value)
	case "ne":
		if !ok {
			return true
		}
		return !valuesEqual(got, c.Value)
	case "gte":
		if !ok {
			return false
		}
		return cmpNumber(got, c.Value, func(a, b float64) bool { return a >= b })
	case "lte":
		if !ok {
			return false
		}
		return cmpNumber(got, c.Value, func(a, b float64) bool { return a <= b })
	case "exists":
		want := true
		if b, isBool := c.Value.(bool); isBool {
			want = b
		}
		if want {
			return ok
		}
		return !ok
	case "in":
		if !ok {
			return false
		}
		list, isList := c.Value.([]interface{})
		if !isList {
			return false
		}
		for _, item := range list {
			if valuesEqual(got, item) {
				return true
			}
		}
		return false
	case "regex":
		if !ok {
			return false
		}
		re, err := regexp.Compile(fmt.Sprint(c.Value))
		if err != nil {
			return false
		}
		return re.MatchString(fmt.Sprint(got))
	default:
		return false
	}
}

func cmpNumber(got, want interface{}, pred func(float64, float64) bool) bool {
	left, lok := asFloat(got)
	right, rok := asFloat(want)
	if !lok || !rok {
		return false
	}
	return pred(left, right)
}

func lookup(raw map[string]interface{}, path string) (interface{}, bool) {
	if raw == nil || path == "" {
		return nil, false
	}
	var cur interface{} = raw
	for _, part := range strings.Split(path, ".") {
		obj, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		next, ok := obj[part]
		if !ok || next == nil {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

func asMap(entity any) map[string]interface{} {
	if entity == nil {
		return nil
	}
	if m, ok := entity.(map[string]interface{}); ok {
		return m
	}
	return canonicalMap(entity)
}

func valuesEqual(a, b interface{}) bool {
	if af, ok := asFloat(a); ok {
		if bf, ok := asFloat(b); ok {
			return af == bf
		}
	}
	ab, aok := a.(bool)
	bb, bok := b.(bool)
	if aok || bok {
		return aok && bok && ab == bb
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

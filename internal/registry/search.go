package registry

import "strings"

// SearchOptions controls how Search filters and ranks results.
type SearchOptions struct {
	Query    string // substring match against name, description, category, tags
	Category string // exact category filter (empty = all)
	Source   string // repo filter, e.g. "github/awesome-copilot" (empty = all)
	Limit    int    // 0 = no limit
}

// Search returns registry entries that match the given options.
// Matching is case-insensitive substring search across name, description,
// category, and tags.
func Search(r *Registry, opts SearchOptions) []SkillEntry {
	q := strings.ToLower(opts.Query)
	cat := strings.ToLower(opts.Category)
	src := strings.ToLower(opts.Source)

	var results []SkillEntry
	for _, e := range r.Skills {
		if cat != "" && strings.ToLower(e.Category) != cat {
			continue
		}
		if src != "" && strings.ToLower(e.Source.Repo) != src {
			continue
		}
		if q != "" && !matchesQuery(e, q) {
			continue
		}
		results = append(results, e)
		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}
	}
	return results
}

func matchesQuery(e SkillEntry, q string) bool {
	if strings.Contains(strings.ToLower(e.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Description), q) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Category), q) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

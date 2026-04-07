package ops

import (
	"sort"
	"strings"

	"github.com/Denver2003/codex-multiaccount-switcher/internal/config"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/domain"
	"github.com/Denver2003/codex-multiaccount-switcher/internal/store"
)

func ListProfiles(paths *config.Resolver) ([]domain.ProfileSummary, error) {
	profiles, err := store.New(paths).VisibleProfiles()
	if err != nil {
		return nil, err
	}

	sort.Slice(profiles, func(i, j int) bool {
		left := strings.ToLower(profiles[i].Label)
		right := strings.ToLower(profiles[j].Label)
		if left != right {
			return left < right
		}

		return profiles[i].ID < profiles[j].ID
	})

	return profiles, nil
}

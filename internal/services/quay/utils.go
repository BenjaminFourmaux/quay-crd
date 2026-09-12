package quay

import (
	"encoding/json"
	"sort"
)

func ParseOrganization(body []byte) (Organization, error) {
	var rawOrg struct {
		Name                string          `json:"name"`
		Email               string          `json:"email,omitempty"`
		Avatar              Avatar          `json:"avatar"`
		IsAdmin             bool            `json:"is_admin"`
		IsMember            bool            `json:"is_member"`
		Teams               map[string]Team `json:"teams"`
		OrderedTeams        []string        `json:"ordered_teams"`
		InvoiceEmail        bool            `json:"invoice_email"`
		InvoiceEmailAddress *string         `json:"invoice_email_address"`
		TagExpirationS      int             `json:"tag_expiration_s"`
		IsFreeAccount       bool            `json:"is_free_account"`
	}

	if err := json.Unmarshal(body, &rawOrg); err != nil {
		return Organization{}, err
	}

	org := Organization{
		Name:                rawOrg.Name,
		Email:               rawOrg.Email,
		Avatar:              rawOrg.Avatar,
		IsAdmin:             rawOrg.IsAdmin,
		IsMember:            rawOrg.IsMember,
		InvoiceEmail:        rawOrg.InvoiceEmail,
		InvoiceEmailAddress: rawOrg.InvoiceEmailAddress,
		TagExpirationS:      rawOrg.TagExpirationS,
		IsFreeAccount:       rawOrg.IsFreeAccount,
	}

	org.Teams = make([]Team, 0, len(rawOrg.Teams))
	seenTeams := make(map[string]struct{}, len(rawOrg.Teams))

	for _, teamName := range rawOrg.OrderedTeams {
		if team, ok := rawOrg.Teams[teamName]; ok {
			org.Teams = append(org.Teams, team)
			seenTeams[teamName] = struct{}{}
		}
	}

	if len(org.Teams) < len(rawOrg.Teams) {
		remaining := make([]string, 0, len(rawOrg.Teams)-len(org.Teams))
		for teamName := range rawOrg.Teams {
			if _, ok := seenTeams[teamName]; !ok {
				remaining = append(remaining, teamName)
			}
		}
		sort.Strings(remaining)
		for _, teamName := range remaining {
			org.Teams = append(org.Teams, rawOrg.Teams[teamName])
		}
	}

	return org, nil
}

func ParseMembers(body []byte) ([]Member, error) {
	var rawMembers struct {
		Name    string   `json:"name"`
		Members []Member `json:"members"`
		CanEdit bool     `json:"can_edit"`
	}

	if err := json.Unmarshal(body, &rawMembers); err != nil {
		return []Member{}, err
	}

	return rawMembers.Members, nil
}

package sirius

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
)

type TeamMember struct {
	ID   int    `json:"id"`
	Name string `json:"displayName"`
}

type TeamCollection struct {
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Members     []struct {
		ID          int    `json:"id"`
		DisplayName string `json:"displayName"`
		Email       string `json:"email"`
	} `json:"members"`
	TeamType *struct {
		Handle string `json:"handle"`
		Label  string `json:"label"`
	} `json:"teamType"`
}

type ReturnedTeamCollection struct {
	Id        int
	Members   []TeamMember
	Name      string
	Type      string
	TypeLabel string
	Selector  string
	Teams     []ReturnedTeamCollection
}

func (r ReturnedTeamCollection) GetAssigneesForFilter() map[int]TeamMember {
	assignees := r.Members
	for _, team := range r.Teams {
		assignees = append(assignees, team.Members...)
	}
	sort.Slice(assignees, func(i, j int) bool {
		return assignees[i].Name < assignees[j].Name
	})
	deduped := map[int]TeamMember{}
	for _, assignee := range assignees {
		if _, exists := deduped[assignee.ID]; !exists {
			deduped[assignee.ID] = assignee
		}
	}
	return deduped
}

func (r ReturnedTeamCollection) HasTeam(id int) bool {
	if r.Id == id {
		return true
	}
	for _, t := range r.Teams {
		if t.Id == id {
			return true
		}
	}
	return false
}

func (m TeamMember) IsSelected(selectedAssignees []string) bool {
	for _, a := range selectedAssignees {
		id, _ := strconv.Atoi(a)
		if m.ID == id {
			return true
		}
	}
	return false
}

func (c *Client) GetTeamsForSelection(ctx Context) ([]ReturnedTeamCollection, error) {
	var v []TeamCollection
	var q []ReturnedTeamCollection

	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/teams", nil)
	if err != nil {
		return q, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logErrorRequest(req, err)
		return q, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logResponse(req, resp, err)
		return q, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		c.logResponse(req, resp, err)
		return q, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		c.logResponse(req, resp, err)
		return q, err
	}

	layTeam := ReturnedTeamCollection{
		Members:  []TeamMember{},
		Name:     "Lay deputy team",
		Selector: "lay-team",
		Teams:    []ReturnedTeamCollection{},
	}

	proTeam := ReturnedTeamCollection{
		Members:  []TeamMember{},
		Name:     "Professional deputy team",
		Selector: "pro-team",
		Teams:    []ReturnedTeamCollection{},
	}

	for _, t := range v {
		if t.TeamType == nil {
			continue
		}

		team := ReturnedTeamCollection{
			Id:        t.ID,
			Name:      t.DisplayName,
			Type:      t.TeamType.Handle,
			TypeLabel: t.TeamType.Label,
			Selector:  strconv.Itoa(t.ID),
			Teams:     []ReturnedTeamCollection{},
		}

		for _, m := range t.Members {
			team.Members = append(team.Members, TeamMember{
				ID:   m.ID,
				Name: m.DisplayName,
			})
		}

		if t.TeamType.Handle == "LAY" {
			layTeam.Members = append(layTeam.Members, team.Members...)
			layTeam.Teams = append(layTeam.Teams, team)
		} else if t.TeamType.Handle == "PRO" {
			proTeam.Members = append(proTeam.Members, team.Members...)
			proTeam.Teams = append(proTeam.Teams, team)
		}

		q = append(q, team)
	}

	q = append(q, layTeam, proTeam)

	sort.Slice(q, func(i, j int) bool {
		return q[i].Name < q[j].Name
	})

	return q, err
}

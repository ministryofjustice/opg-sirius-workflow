package sirius

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TeamMembers struct {
	TeamMembersId   int    `json:"id"`
	TeamMembersName string `json:"name"`
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
	Id               int
	Members          []TeamMembers
	Name             string
	UserSelectedTeam int
	SelectedTeamId   int
	Type             string
	TypeLabel        string
	TeamTypeHandle   string
	TeamTypeLabel    string
}

type TeamStoredData struct {
	TeamId       int
	SelectedTeam int
}

func (c *Client) GetTeamSelection(ctx Context, loggedInTeamId int, selectedTeamName int, selectedTeamMembers TeamSelected) ([]ReturnedTeamCollection, error) {
	var v []TeamCollection
	var q []ReturnedTeamCollection
	var k TeamStoredData

	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/teams", nil)
	if err != nil {
		return q, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return q, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return q, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return q, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return q, err
	}

	if selectedTeamName == 0 && k.TeamId == 0 {
		k.TeamId = loggedInTeamId
	} else {
		k.TeamId = selectedTeamName
	}

	if selectedTeamMembers.selectedTeamToAssignTask == 0 && k.SelectedTeam == 0 {
		k.SelectedTeam = loggedInTeamId
	} else {
		k.SelectedTeam = selectedTeamMembers.selectedTeamToAssignTask
	}

	teams := make([]ReturnedTeamCollection, len(v))
	fmt.Println(v)
	for i, t := range v {

		fmt.Println("v i teamtype")
		fmt.Println(v[i].TeamType)
		teams[i] = ReturnedTeamCollection{
			Id:   t.ID,
			Name: t.DisplayName,
			Type: "",
		}

		for _, m := range t.Members {
			teams[i].Members = append(teams[i].Members, TeamMembers{
				TeamMembersId:   m.ID,
				TeamMembersName: m.Email,
			})
		}

		for i := range teams {
			teams[i].UserSelectedTeam = k.TeamId
			teams[i].SelectedTeamId = k.SelectedTeam

		}
		if t.TeamType != nil {
			teams[i].Type = t.TeamType.Handle
			teams[i].TypeLabel = t.TeamType.Label
		}
	}

	teams = filterOutNonLayTeams(teams)
	return teams, err
}

func filterOutNonLayTeams(v []ReturnedTeamCollection) []ReturnedTeamCollection {
	var filteredTeams []ReturnedTeamCollection
	for _, s := range v {
		if len(s.Type) != 0 {
			filteredTeams = append(filteredTeams, s)
		}
	}
	return filteredTeams
}

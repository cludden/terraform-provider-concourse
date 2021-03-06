package concourse

import (
	"fmt"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"github.com/hashicorp/terraform/helper/schema"
)

// teamIDAsString converts a given numeric team ID, which is required, because Terraform resource data IDs must be
// strings.
func teamIDAsString(id int) string {
	return fmt.Sprintf("%d", id)
}

func teamExists(concourse concourse.Client, name string) (bool, error) {
	teams, err := concourse.ListTeams()
	if err != nil {
		return false, fmt.Errorf("unable to list teams: %v", err)
	}
	for _, team := range teams {
		if team.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func resourceTeamCreate(d *schema.ResourceData, m interface{}) error {
	concourse := m.(Config).Concourse()

	name := d.Get("name").(string)
	authUsers := d.Get("auth_users").([]interface{})
	authGroups := d.Get("auth_groups").([]interface{})

	var users, groups []string

	for _, user := range authUsers {
		users = append(users, user.(string))
	}

	for _, group := range authGroups {
		groups = append(groups, group.(string))
	}

	t := atc.Team{
		Name: name,
		Auth: map[string]map[string][]string{
			"member": map[string][]string{
				"users":  users,
				"groups": groups,
			},
		},
	}
	team, created, updated, err := concourse.Team(name).CreateOrUpdate(t)
	if err != nil {
		return fmt.Errorf("could not create team: %v", err)
	}
	if !created && !updated {
		return fmt.Errorf("could not create/update team %s: neither 'created' nor 'updated' was set to true", name)
	}
	d.SetId(teamIDAsString(team.ID))
	return resourceTeamRead(d, m)
}

func resourceTeamRead(d *schema.ResourceData, m interface{}) error {
	concourse := m.(Config).Concourse()

	id := d.Id()
	name := d.Get("name").(string)

	teams, err := concourse.ListTeams()
	if err != nil {
		return fmt.Errorf("unable to list teams: %v", err)
	}

	for _, team := range teams {
		strID := teamIDAsString(team.ID)

		// To simplify things, we allow either the (internal) resource ID
		// or the name to be used when importing a team resource.
		if id == strID || id == team.Name || (name != "" && name == team.Name) {
			d.SetId(strID)
			if err := d.Set("name", team.Name); err != nil {
				return err
			}

			for role, auth := range team.Auth {
				if role == "member" {
					users, ok := auth["users"]
					if ok {
						authUsers := make([]interface{}, len(users))
						for i, user := range users {
							authUsers[i] = user
						}
						d.Set("auth_users", authUsers)
					} else {
						d.Set("auth_users", []interface{}{})
					}

					groups, ok := auth["groups"]
					if ok {
						authGroups := make([]interface{}, len(groups))
						for i, group := range groups {
							authGroups[i] = group
						}
						d.Set("auth_groups", authGroups)
					} else {
						d.Set("auth_groups", []interface{}{})
					}
				}
			}

			return nil
		}
	}

	// If a team with the given ID/name cannot be found, it has probably been already been deleted.
	// We will have to update the state then...
	d.SetId("")
	return nil

}

func resourceTeamUpdate(d *schema.ResourceData, m interface{}) error {
	concourse := m.(Config).Concourse()

	newName := ""
	if d.HasChange("name") {
		newName = d.Get("name").(string)
	}
	teams, err := concourse.ListTeams()
	id := d.Id()
	if err != nil {
		return fmt.Errorf("unable to list teams: %v", err)
	}

	var t *atc.Team
	for _, team := range teams {
		if id == teamIDAsString(team.ID) {
			t = &team
			oldName := team.Name
			if newName != "" && oldName != newName {
				t.Name = newName
				_, err := concourse.Team(team.Name).RenameTeam(team.Name, newName)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if t == nil {
		return fmt.Errorf("team with ID %s not found", d.Id())
	}

	if d.HasChange("auth_users") || d.HasChange("auth_groups") {
		authUsers := d.Get("auth_users").([]interface{})
		authGroups := d.Get("auth_groups").([]interface{})
		var users, groups []string

		for _, user := range authUsers {
			users = append(users, user.(string))
		}

		for _, group := range authGroups {
			groups = append(groups, group.(string))
		}

		t.Auth = map[string]map[string][]string{
			"member": map[string][]string{
				"users":  users,
				"groups": groups,
			},
		}

		_, created, updated, err := concourse.Team(t.Name).CreateOrUpdate(*t)
		if err != nil {
			return fmt.Errorf("could not create team: %v", err)
		}
		if !created && !updated {
			return fmt.Errorf("could not create/update team %s: neither 'created' nor 'updated' was set to true", t.Name)
		}
	}

	return resourceTeamRead(d, m)
}

func resourceTeamDelete(d *schema.ResourceData, m interface{}) error {
	concourse := m.(Config).Concourse()
	name := d.Get("name").(string)
	return concourse.Team(name).DestroyTeam(name)
}

func resourceTeamExists(d *schema.ResourceData, m interface{}) (bool, error) {
	id := d.Id()
	concourse := m.(Config).Concourse()

	teams, err := concourse.ListTeams()
	if err != nil {
		return false, fmt.Errorf("unable to list teams: %v", err)
	}
	for _, team := range teams {
		teamID := teamIDAsString(team.ID)
		if teamID == id {
			return true, nil
		}
	}
	return false, nil
}

func resourceTeamState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	nameOrID := d.Id()
	if err := resourceTeamRead(d, m); err != nil {
		return nil, err
	}
	if d.Id() == "" {
		return nil, fmt.Errorf("team with ID or name %s not found", nameOrID)
	}
	return []*schema.ResourceData{d}, nil

}

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceTeamCreate,
		Read:   resourceTeamRead,
		Update: resourceTeamUpdate,
		Delete: resourceTeamDelete,
		Exists: resourceTeamExists,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Team name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"auth_users": {
				Description: "User access / authorization",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:          schema.TypeString,
					MinItems:      0,
					PromoteSingle: true,
				},
				Optional: true,
			},
			"auth_groups": {
				Description: "Group access / authorization",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:          schema.TypeString,
					MinItems:      0,
					PromoteSingle: true,
				},
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: resourceTeamState,
		},
	}
}

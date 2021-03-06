package services

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/Juniper/contrail/pkg/common"
	"github.com/Juniper/contrail/pkg/models"
)

func TestCreateEventJSONEncoding(t *testing.T) {
	e := &Event{
		Request: &Event_CreateVirtualNetworkRequest{
			CreateVirtualNetworkRequest: &CreateVirtualNetworkRequest{
				VirtualNetwork: &models.VirtualNetwork{
					UUID: "vn_uuid",
				},
			},
		},
	}
	m, err := json.Marshal(e)
	assert.NoError(t, err, "marhsal event failed")
	var i map[string]interface{}
	err = json.Unmarshal(m, &i)
	assert.NoError(t, err, "unmarhsal event failed")
	assert.Equal(t, "virtual_network", i["kind"])
	assert.Equal(t, "CREATE", i["operation"])

	var d Event
	err = json.Unmarshal(m, &d)
	assert.NoError(t, err, "unmarhsal event failed")
	request := d.GetCreateVirtualNetworkRequest()
	assert.Equal(t, "vn_uuid", request.GetVirtualNetwork().GetUUID())

	d2, err := NewEvent(&EventOption{
		Kind:      i["kind"].(string),
		Operation: i["operation"].(string),
		Data:      i["data"].(map[string]interface{}),
	})
	if assert.NoError(t, err) {
		request = d2.GetCreateVirtualNetworkRequest()
		assert.Equal(t, "vn_uuid", request.GetVirtualNetwork().GetUUID())
	}
}

func TestDeleteEventJSONEncoding(t *testing.T) {
	e := &Event{
		Request: &Event_DeleteVirtualNetworkRequest{
			DeleteVirtualNetworkRequest: &DeleteVirtualNetworkRequest{
				ID: "vn_uuid",
			},
		},
	}
	m, err := json.Marshal(e)
	assert.NoError(t, err, "marhsal event failed")
	fmt.Println(string(m))
	var i map[string]interface{}
	err = json.Unmarshal(m, &i)
	assert.NoError(t, err, "unmarhsal event failed")
	assert.Equal(t, "virtual_network", i["kind"])
	assert.Equal(t, "DELETE", i["operation"])
	assert.Equal(t, "vn_uuid", i["data"].(map[string]interface{})["uuid"])

	var d Event
	err = json.Unmarshal(m, &d)
	assert.NoError(t, err, "unmarhsal event failed")
	request := d.GetDeleteVirtualNetworkRequest()
	assert.Equal(t, "vn_uuid", request.ID)
}

func TestCreateEventYAMLEncoding(t *testing.T) {
	t.Skip("TODO: Fix me")
	e := &Event{
		Request: &Event_CreateVirtualNetworkRequest{
			CreateVirtualNetworkRequest: &CreateVirtualNetworkRequest{
				VirtualNetwork: &models.VirtualNetwork{
					UUID: "vn_uuid",
				},
			},
		},
	}
	m, err := yaml.Marshal(e)
	fmt.Println(string(m))
	assert.NoError(t, err, "marhsal event failed")
	var i map[string]interface{}
	err = yaml.Unmarshal(m, &i)
	assert.NoError(t, err, "unmarhsal event failed")
	assert.Equal(t, "virtual_network", i["kind"])
	assert.Equal(t, "CREATE", i["operation"])

	var d Event
	err = yaml.Unmarshal(m, &d)
	assert.NoError(t, err, "unmarhsal event failed")
	request := d.GetCreateVirtualNetworkRequest()
	assert.Equal(t, "vn_uuid", request.GetVirtualNetwork().GetUUID())
	i = common.YAMLtoJSONCompat(i).(map[string]interface{})
	d2, err := NewEvent(&EventOption{
		Kind:      i["kind"].(string),
		Operation: i["operation"].(string),
		Data:      i["data"].(map[string]interface{}),
	})
	if assert.NoError(t, err) {
		request = d2.GetCreateVirtualNetworkRequest()
		assert.Equal(t, "vn_uuid", request.GetVirtualNetwork().GetUUID())
	}
}

func TestSortEventListByDependency(t *testing.T) {
	tests := []struct {
		name        string
		events      []*Event
		sortedOrder []string
		fails       bool
	}{
		{
			name:        "no events",
			events:      []*Event{},
			sortedOrder: []string{},
		},
		{
			name: "single event",
			events: []*Event{
				virtualNetworkCreateEvent(virtualNetwork(&vnParameters{uuid: "vn-uuid"})),
			},
			sortedOrder: []string{"vn-uuid"},
		},
		{
			name: "reference dependency",
			events: []*Event{
				virtualNetworkCreateEvent(virtualNetwork(&vnParameters{
					uuid: "vn-uuid",
					networkPolicyRefs: []*models.VirtualNetworkNetworkPolicyRef{
						{
							UUID: "network-policy-uuid",
						},
					},
				})),
				networkPolicyCreateEvent(networkPolicy("network-policy-uuid")),
			},
			sortedOrder: []string{"network-policy-uuid", "vn-uuid"},
		},
		{
			name: "parent-child dependency",
			events: []*Event{
				virtualNetworkCreateEvent(virtualNetwork(&vnParameters{
					uuid:       "vn-uuid",
					parentUUID: "project-uuid",
					parentType: "project",
				})),
				projectCreateEvent(project("project-uuid")),
			},

			sortedOrder: []string{"project-uuid", "vn-uuid"},
		},
		{
			name: "circular dependency",
			events: []*Event{
				virtualNetworkCreateEvent(virtualNetwork(&vnParameters{
					uuid: "vn-one-uuid",
					virtualNetworkRefs: []*models.VirtualNetworkVirtualNetworkRef{
						{
							UUID: "vn-two-uuid",
						},
					},
				})),
				virtualNetworkCreateEvent(virtualNetwork(&vnParameters{
					uuid: "vn-two-uuid",
					virtualNetworkRefs: []*models.VirtualNetworkVirtualNetworkRef{
						{
							UUID: "vn-one-uuid",
						},
					},
				})),
			},
			fails: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventList := EventList{tt.events}

			err := eventList.Sort()

			if tt.fails {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for i, e := range eventList.Events {
					assert.Equal(t, tt.sortedOrder[i], e.GetResource().GetUUID())
				}
			}
		})
	}
}

func projectCreateEvent(p *models.Project) *Event {
	return &Event{
		Request: &Event_CreateProjectRequest{
			&CreateProjectRequest{
				Project: p,
			},
		},
	}
}

func virtualNetworkCreateEvent(vn *models.VirtualNetwork) *Event {
	return &Event{
		Request: &Event_CreateVirtualNetworkRequest{
			&CreateVirtualNetworkRequest{
				VirtualNetwork: vn,
			},
		},
	}
}

func networkPolicyCreateEvent(np *models.NetworkPolicy) *Event {
	return &Event{
		Request: &Event_CreateNetworkPolicyRequest{
			CreateNetworkPolicyRequest: &CreateNetworkPolicyRequest{
				NetworkPolicy: np,
			},
		},
	}
}

func project(uuid string) *models.Project {
	p := models.MakeProject()
	p.UUID = uuid
	return p
}

type vnParameters struct {
	uuid               string
	parentUUID         string
	parentType         string
	networkPolicyRefs  []*models.VirtualNetworkNetworkPolicyRef
	virtualNetworkRefs []*models.VirtualNetworkVirtualNetworkRef
}

func virtualNetwork(p *vnParameters) *models.VirtualNetwork {
	vn := models.MakeVirtualNetwork()
	vn.UUID = p.uuid
	vn.ParentUUID = p.parentUUID
	vn.ParentType = p.parentType
	if len(p.networkPolicyRefs) > 0 {
		vn.NetworkPolicyRefs = p.networkPolicyRefs
	}
	if len(p.virtualNetworkRefs) > 0 {
		vn.VirtualNetworkRefs = p.virtualNetworkRefs
	}
	return vn
}

func networkPolicy(uuid string) *models.NetworkPolicy {
	np := models.MakeNetworkPolicy()
	np.UUID = uuid
	return np
}

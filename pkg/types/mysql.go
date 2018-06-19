package types

import (
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func truePtr() *bool {
	b := true
	return &b
}

// Plan is the default mysql plan
var Plan = osb.Plan{
	Name:        "default",
	ID:          "86064792-7ea2-467b-af93-ac9694d96d5b",
	Description: "The default MySQL for an artifact",
	Free:        truePtr(),
	Schemas: &osb.Schemas{
		ServiceInstance: &osb.ServiceInstanceSchema{
			Create: &osb.InputParametersSchema{
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						// "color": map[string]interface{}{
						// 	"type":    "string",
						// 	"default": "Clear",
						// 	"enum": []string{
						// 		"Clear",
						// 		"Beige",
						// 		"Grey",
						// 	},
						// },
						"size": map[string]interface{}{
							"type":    "integer",
							"default": 20,
						},
						"Artifact": map[string]interface{}{
							"type": "string",
						},
						"DeploymentType": map[string]interface{}{
							"type":    "string",
							"default": "kube",
						},
					},
				},
			},
		},
	},
}

package schema

type CommandSchema struct {
	Commands map[string]CommandDoc `json:"commands"`
}

type CommandDoc struct {
	Description string             `json:"description"`
	Usage       string             `json:"usage"`
	Flags       map[string]FlagDoc `json:"flags,omitempty"`
}

type FlagDoc struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

func BuildCommandSchema() CommandSchema {
	return CommandSchema{
		Commands: map[string]CommandDoc{
			"query": {
				Description: "Query multiple records from a runtime model.",
				Usage:       "mc query <project.database.model|database.model>",
				Flags: map[string]FlagDoc{
					"where":   {Type: "json", Required: false, Description: "JSON where filter."},
					"select":  {Type: "string[]", Required: false, Description: "Selected fields."},
					"orderBy": {Type: "json", Required: false, Description: "JSON orderBy expression."},
					"take":    {Type: "int", Required: false, Description: "Page size."},
					"skip":    {Type: "int", Required: false, Description: "Records to skip."},
				},
			},
			"get": {
				Description: "Query a single record from a runtime model.",
				Usage:       "mc get <project.database.model|database.model>",
				Flags: map[string]FlagDoc{
					"where":  {Type: "json", Required: true, Description: "JSON where filter."},
					"select": {Type: "string[]", Required: false, Description: "Selected fields."},
				},
			},
			"count": {
				Description: "Count records from a runtime model.",
				Usage:       "mc count <project.database.model|database.model>",
				Flags: map[string]FlagDoc{
					"where": {Type: "json", Required: false, Description: "JSON where filter."},
				},
			},
			"aggregate": {
				Description: "Aggregate records from a runtime model.",
				Usage:       "mc aggregate <project.database.model|database.model>",
				Flags: map[string]FlagDoc{
					"where":  {Type: "json", Required: false, Description: "JSON where filter."},
					"fields": {Type: "string[]", Required: false, Description: "Aggregation fields."},
				},
			},
			"describe": {
				Description: "Describe a runtime model using GraphQL introspection.",
				Usage:       "mc describe <project.database.model|database.model>",
			},
			"schema commands": {
				Description: "Export static CLI command schema.",
				Usage:       "mc schema commands",
			},
		},
	}
}

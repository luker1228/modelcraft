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
			"auth login": {
				Description: "Login with end-user credentials or a PAT token.",
				Usage:       "mc auth login [--token mc_pat_xxx | --username <u> --password <p>]",
				Flags: map[string]FlagDoc{
					"token":       {Type: "string", Required: false, Description: "Personal Access Token (mc_pat_xxx)."},
					"username":    {Type: "string", Required: false, Description: "End-user username."},
					"password":    {Type: "string", Required: false, Description: "End-user password."},
					"org":         {Type: "string", Required: false, Description: "Organization slug (auto-resolved when omitted)."},
					"server":      {Type: "string", Required: false, Description: "Gateway base URL."},
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
				},
			},
			"auth logout": {
				Description: "Logout and clear local credentials.",
				Usage:       "mc auth logout",
				Flags: map[string]FlagDoc{
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
				},
			},
			"auth status": {
				Description: "Show current authentication status (org, user, projects, current project).",
				Usage:       "mc auth status",
				Flags: map[string]FlagDoc{
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
				},
			},
			"auth refresh": {
				Description: "Refresh access token using the stored refresh token.",
				Usage:       "mc auth refresh",
				Flags: map[string]FlagDoc{
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
				},
			},
			"auth switch-project": {
				Description: "Set the local default project context.",
				Usage:       "mc auth switch-project <slug>",
				Flags: map[string]FlagDoc{
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
				},
			},
			"run": {
				Description: "Execute a raw GraphQL query against a runtime model endpoint.",
				Usage:       "mc run <project.database.model|database.model> [query]",
				Flags: map[string]FlagDoc{
					"credentials": {Type: "string", Required: false, Description: "Path to credentials file."},
					"project":     {Type: "string", Required: false, Description: "Project slug override."},
				},
			},
			"describe": {
				Description: "Describe a runtime model's fields and types via GraphQL introspection.",
				Usage:       "mc describe <project.database.model|database.model>",
			},
			"catalog projects": {
				Description: "List projects accessible to the current end-user.",
				Usage:       "mc catalog projects",
			},
			"catalog databases": {
				Description: "List databases within a project.",
				Usage:       "mc catalog databases [--project <slug>]",
			},
			"catalog models": {
				Description: "List models within a database.",
				Usage:       "mc catalog models --database <name> [--project <slug>]",
			},
			"schema commands": {
				Description: "Export static CLI command schema.",
				Usage:       "mc schema commands",
			},
		},
	}
}

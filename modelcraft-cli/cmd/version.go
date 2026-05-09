package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand(info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:  "version",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			type versionData struct {
				Version   string `json:"version"`
				Commit    string `json:"commit"`
				BuildTime string `json:"buildTime"`
			}

			payload := struct {
				OK   bool        `json:"ok"`
				Data versionData `json:"data"`
			}{
				OK: true,
				Data: versionData{
					Version:   info.Version,
					Commit:    info.Commit,
					BuildTime: info.BuildTime,
				},
			}

			encoded, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal version response: %w", err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", encoded)
			return err
		},
	}
}

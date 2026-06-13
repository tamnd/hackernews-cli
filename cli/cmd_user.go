package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/hackernews-cli/hackernews"
)

func (a *App) userCmd() *cobra.Command {
	var submissions bool
	cmd := &cobra.Command{
		Use:   "user <username>",
		Short: "Show an HN user profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := hackernews.ParseUsername(args[0])
			if err != nil {
				return codeError(exitUsage, err)
			}
			a.progressf("fetching user %q...", name)
			user, ids, err := a.client.User(cmd.Context(), name)
			if err != nil {
				return mapFetchErr(err)
			}
			if err := a.render([]hackernews.User{user}); err != nil {
				return err
			}
			if submissions && len(ids) > 0 {
				n := a.effectiveLimit(20)
				a.progressf("resolving %d submissions (showing %d)...", len(ids), n)
				stories, err := a.client.UserSubmissions(cmd.Context(), ids, n)
				if err != nil {
					return mapFetchErr(err)
				}
				return a.render(stories)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&submissions, "submissions", false, "also list the user's submissions")
	return cmd
}

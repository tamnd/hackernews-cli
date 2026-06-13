package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/hackernews-cli/hackernews"
)

func (a *App) itemCmd() *cobra.Command {
	var depth int
	cmd := &cobra.Command{
		Use:   "item <id>",
		Short: "Fetch a story and its comment tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := hackernews.ParseItemID(args[0])
			if err != nil {
				return codeError(exitUsage, err)
			}
			a.progressf("fetching item %d (depth %d)...", id, depth)
			story, comments, err := a.client.Item(cmd.Context(), id, depth)
			if err != nil {
				return mapFetchErr(err)
			}
			if err := a.render([]hackernews.Story{story}); err != nil {
				return err
			}
			if len(comments) > 0 {
				return a.render(comments)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 1, "comment tree depth (0=item only, -1=full tree)")
	return cmd
}

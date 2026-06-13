package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tamnd/hackernews-cli/hackernews"
)

func (a *App) searchCmd() *cobra.Command {
	var (
		tags        string
		sort        string
		since       string
		minPoints   int
		minComments int
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search HN via Algolia",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			opts := hackernews.SearchOptions{
				Query:       args[0],
				Tags:        tags,
				Sort:        sort,
				Since:       since,
				MinPoints:   minPoints,
				MinComments: minComments,
				Limit:       n,
			}
			a.progressf("searching for %q...", args[0])
			hits, err := a.client.Search(cmd.Context(), opts)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(hits, len(hits))
		},
	}
	cmd.Flags().StringVar(&tags, "tags", "story", "Algolia tags filter: story, comment, ask_hn, show_hn, front_page, job")
	cmd.Flags().StringVar(&sort, "sort", "relevance", "sort order: relevance or date")
	cmd.Flags().StringVar(&since, "since", "", "only results newer than this duration (e.g. 24h, 7d)")
	cmd.Flags().IntVar(&minPoints, "points", 0, "minimum points")
	cmd.Flags().IntVar(&minComments, "comments", 0, "minimum comment count")
	return cmd
}

func (a *App) updatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "updates",
		Short: "Show recently changed items and profiles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			u, err := a.client.Updates(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "changed items:    %d\n", len(u.Items))
			_, _ = fmt.Fprintf(w, "changed profiles: %d\n", len(u.Profiles))
			if len(u.Items) > 0 {
				_, _ = fmt.Fprintln(w)
				for _, id := range u.Items {
					_, _ = fmt.Fprintln(w, id)
				}
			}
			return nil
		},
	}
}

func (a *App) maxItemCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "maxitem",
		Short: "Print the current maximum item id",
		RunE: func(cmd *cobra.Command, _ []string) error {
			id, err := a.client.MaxItem(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}

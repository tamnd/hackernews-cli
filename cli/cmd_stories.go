package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) storiesCmd(name, short, endpoint string, defaultLimit int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(defaultLimit)
			a.progressf("fetching %d %s stories...", n, name)
			stories, err := a.client.StoryList(cmd.Context(), endpoint, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
	return cmd
}

func (a *App) jobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jobs",
		Short: "HN job listings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d job listings...", n)
			stories, err := a.client.StoryList(cmd.Context(), "jobstories", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stories, len(stories))
		},
	}
}

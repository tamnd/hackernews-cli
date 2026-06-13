package cli

import (
	"errors"

	"github.com/tamnd/hackernews-cli/hackernews"
)

func isNotFound(err error) bool {
	return errors.Is(err, hackernews.ErrNotFound)
}

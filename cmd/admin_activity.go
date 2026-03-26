package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use: "activity", Short: "View audit log",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		path := "/v1/activity"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	activityCmd.Flags().Int("limit", 50, "Max results")
	rootCmd.AddCommand(activityCmd)
}

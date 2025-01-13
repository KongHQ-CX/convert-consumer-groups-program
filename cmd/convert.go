/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/konghq-cx/convert-consumer-groups-program/internal/converter"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source-file")
		destination, _ := cmd.Flags().GetString("output-file")
		converter.Convert(source, destination)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringP("source-file", "s", "", "Source file")
	convertCmd.Flags().StringP("output-file", "o", "", "Output file")
	convertCmd.MarkFlagRequired("source-file")
	convertCmd.MarkFlagRequired("output-file")
}

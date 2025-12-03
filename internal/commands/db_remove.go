package commands

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var (
	removeClusterName string
	removeKPIName     string
	removeAll         bool
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Delete data from the database",
	Long: `Delete clusters, KPI metrics, or error records from the database.

	WARNING: All remove operations are immediate and cannot be undone.`,
}

var removeClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Remove a cluster and all its data",
	Long: `Delete a cluster record and all associated KPI metrics from the database.

WARNING: This operation cannot be undone. All metric samples for the cluster will be permanently deleted.`,
	Example: `  # Remove a cluster
  kpi-collector db remove clusters --name="old-cluster"`,
	RunE: runRemoveClusters,
}

var removeKPIsCmd = &cobra.Command{
	Use:   "kpis",
	Short: "Remove KPI metrics",
	Long:  `Delete KPI metrics from the database, optionally filtered by cluster and KPI name.`,
	Example: `  # Remove all KPIs from a cluster
  kpi-collector db remove kpis --cluster-name="mycluster1"
  
  # Remove specific KPI from a cluster
  kpi-collector db remove kpis --cluster-name="mycluster1" --name="cpu-system"`,
	RunE: runRemoveKPIs,
}

var removeErrorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Clear error counts",
	Long:  `Reset error counts for KPI queries.`,
	Example: `  # Clear errors for a specific KPI
  kpi-collector db remove errors --name="cpu-system"
  
  # Clear all errors
  kpi-collector db remove errors --all`,
	RunE: runRemoveErrors,
}

func init() {
	dbCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeClustersCmd)
	removeCmd.AddCommand(removeKPIsCmd)
	removeCmd.AddCommand(removeErrorsCmd)

	// Flags for 'remove clusters'
	removeClustersCmd.Flags().StringVar(&removeClusterName, "name", "",
		"cluster name to remove (required)")
	_ = removeClustersCmd.MarkFlagRequired("name")

	// Flags for 'remove kpis'
	removeKPIsCmd.Flags().StringVar(&removeClusterName, "cluster-name", "",
		"cluster name (required)")
	removeKPIsCmd.Flags().StringVar(&removeKPIName, "name", "",
		"KPI name to remove (optional)")
	_ = removeKPIsCmd.MarkFlagRequired("cluster-name")

	// Flags for 'remove errors'
	removeErrorsCmd.Flags().StringVar(&removeKPIName, "name", "",
		"KPI name to clear errors for")
	removeErrorsCmd.Flags().BoolVar(&removeAll, "all", false,
		"clear all error records")
}

func runRemoveClusters(cmd *cobra.Command, args []string) error {
	db, _, err := connectToDB()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Check if cluster exists
	clusters, err := listClusters(db, removeClusterName)
	if err != nil {
		return fmt.Errorf("failed to query cluster: %w", err)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("cluster '%s' not found", removeClusterName)
	}

	cluster := clusters[0]

	// Delete metrics first
	result, err := db.Exec("DELETE FROM query_results WHERE cluster_id = ?", cluster.ID)
	if err != nil {
		return fmt.Errorf("failed to delete metrics: %w", err)
	}
	metricsDeleted, _ := result.RowsAffected()

	// Delete cluster
	_, err = db.Exec("DELETE FROM clusters WHERE id = ?", cluster.ID)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	fmt.Printf("✓ Deleted cluster '%s' and %s metric samples.\n",
		cluster.Name, humanize.Comma(metricsDeleted))
	return nil
}

func runRemoveKPIs(cmd *cobra.Command, args []string) error {
	db, _, err := connectToDB()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Get cluster ID
	clusters, err := listClusters(db, removeClusterName)
	if err != nil {
		return fmt.Errorf("failed to query cluster: %w", err)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("cluster '%s' not found", removeClusterName)
	}
	cluster := clusters[0]

	// Build DELETE query
	query := "DELETE FROM query_results WHERE cluster_id = ?"
	queryArgs := []interface{}{cluster.ID}

	if removeKPIName != "" {
		query += " AND kpi_id = ?"
		queryArgs = append(queryArgs, removeKPIName)
	}

	// Execute deletion
	result, err := db.Exec(query, queryArgs...)
	if err != nil {
		return fmt.Errorf("failed to delete metrics: %w", err)
	}
	deleted, _ := result.RowsAffected()

	if deleted == 0 {
		fmt.Println("No metrics found matching the criteria.")
		return nil
	}

	fmt.Printf("✓ Deleted %s metric samples.\n", humanize.Comma(deleted))
	return nil
}

func runRemoveErrors(cmd *cobra.Command, args []string) error {
	db, _, err := connectToDB()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	var query string
	var queryArgs []interface{}

	if removeAll {
		query = "DELETE FROM query_errors"
	} else if removeKPIName != "" {
		query = "DELETE FROM query_errors WHERE kpi_id = ?"
		queryArgs = append(queryArgs, removeKPIName)
	} else {
		return fmt.Errorf("must specify either --name or --all")
	}

	result, err := db.Exec(query, queryArgs...)
	if err != nil {
		return fmt.Errorf("failed to delete errors: %w", err)
	}

	deleted, _ := result.RowsAffected()
	if deleted == 0 {
		fmt.Println("No error records found.")
		return nil
	}

	fmt.Printf("✓ Cleared %d error record(s).\n", deleted)
	return nil
}

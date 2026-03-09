//go:build unit
// +build unit

package cli

import (
	"reflect"
	"strings"
	"testing"
)

// ============================================================================
// Test Command Structure Existence
// ============================================================================

func TestPublishCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := PublishCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"Upload", "Release", "Rollout", "Promote", "Halt", "Rollback",
		"Status", "Tracks", "Capabilities", "Listing", "Details", "Images",
		"Assets", "Deobfuscation", "Testers", "Builds", "BetaGroups", "InternalShare",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PublishCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PublishCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("PublishCmd.%s should have help tag", name)
		}
	}

	actualFields := v.NumField()
	if actualFields != len(expectedSubcommands) {
		t.Errorf("PublishCmd has %d fields, expected %d", actualFields, len(expectedSubcommands))
	}
}

func TestReviewsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := ReviewsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"List", "Get", "Reply", "ResponseGet", "ResponseDelete",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("ReviewsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("ReviewsCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("ReviewsCmd.%s should have help tag", name)
		}
	}
}

func TestVitalsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := VitalsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"Crashes", "Anrs", "Errors", "Metrics", "Anomalies", "Query", "Capabilities",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("VitalsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("VitalsCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("VitalsCmd.%s should have help tag", name)
		}
	}
}

func TestConfigCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := ConfigCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"Init", "Doctor", "Path", "Get", "Set", "Print", "Export", "Import", "Completion",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("ConfigCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("ConfigCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("ConfigCmd.%s should have help tag", name)
		}
	}
}

func TestAnalyticsCmd_Exists(t *testing.T) {
	cmd := AnalyticsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Query", "Capabilities"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("AnalyticsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AnalyticsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestAppsCmd_Exists(t *testing.T) {
	cmd := AppsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"List", "Get"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("AppsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AppsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestGamesCmd_Exists(t *testing.T) {
	cmd := GamesCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Achievements", "Scores", "Events", "Players", "Capabilities"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("GamesCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("GamesCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestPurchasesCmd_Exists(t *testing.T) {
	cmd := PurchasesCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Products", "Subscriptions", "Verify", "Voided", "Capabilities"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PurchasesCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PurchasesCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestMonetizationCmd_Exists(t *testing.T) {
	cmd := MonetizationCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Products", "Subscriptions", "BasePlans", "Offers", "Capabilities"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("MonetizationCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("MonetizationCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestPermissionsCmd_Exists(t *testing.T) {
	cmd := PermissionsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Users", "Grants", "List"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PermissionsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PermissionsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestRecoveryCmd_Exists(t *testing.T) {
	cmd := RecoveryCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"List", "Create", "Deploy", "Cancel"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("RecoveryCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("RecoveryCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestIntegrityCmd_Exists(t *testing.T) {
	cmd := IntegrityCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	if _, ok := typeOfCmd.FieldByName("Decode"); !ok {
		t.Error("IntegrityCmd missing Decode subcommand")
	}
}

func TestCustomAppCmd_Exists(t *testing.T) {
	// CustomAppCmd is a stub with no subcommands yet
	_ = CustomAppCmd{}
}

func TestGroupingCmd_Exists(t *testing.T) {
	// GroupingCmd is a stub with no subcommands yet
	_ = GroupingCmd{}
}

func TestMigrateCmd_Exists(t *testing.T) {
	// MigrateCmd is a stub with no subcommands yet
	_ = MigrateCmd{}
}

// ============================================================================
// Test Implemented Commands Return Proper Errors (not stubs)
// ============================================================================

func TestPublishCommands_Implemented(t *testing.T) {
	globals := &Globals{}

	// Commands that return success with hardcoded info (no auth/package needed)
	t.Run("PublishCapabilitiesCmd", func(t *testing.T) {
		cmd := &PublishCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("PublishCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	t.Run("PublishAssetsSpecCmd", func(t *testing.T) {
		cmd := &PublishAssetsSpecCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("PublishAssetsSpecCmd.Run() should succeed, got: %v", err)
		}
	})

	// Commands that require package/auth — verify they don't return "not yet implemented"
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"PublishRolloutCmd", &PublishRolloutCmd{}},
		{"PublishPromoteCmd", &PublishPromoteCmd{}},
		{"PublishHaltCmd", &PublishHaltCmd{}},
		{"PublishRollbackCmd", &PublishRollbackCmd{}},
		{"PublishStatusCmd", &PublishStatusCmd{}},
		{"PublishListingUpdateCmd", &PublishListingUpdateCmd{}},
		{"PublishListingGetCmd", &PublishListingGetCmd{}},
		{"PublishListingDeleteCmd", &PublishListingDeleteCmd{}},
		{"PublishDetailsGetCmd", &PublishDetailsGetCmd{}},
		{"PublishDetailsUpdateCmd", &PublishDetailsUpdateCmd{}},
		{"PublishDetailsPatchCmd", &PublishDetailsPatchCmd{}},
		{"PublishImagesUploadCmd", &PublishImagesUploadCmd{Type: "icon", File: "icon.png"}},
		{"PublishImagesListCmd", &PublishImagesListCmd{Type: "icon"}},
		{"PublishImagesDeleteCmd", &PublishImagesDeleteCmd{Type: "icon", ID: "123"}},
		{"PublishImagesDeleteAllCmd", &PublishImagesDeleteAllCmd{Type: "icon"}},
		{"PublishAssetsUploadCmd", &PublishAssetsUploadCmd{Dir: "assets"}},
		{"PublishDeobfuscationUploadCmd", &PublishDeobfuscationUploadCmd{File: "mapping.txt", Type: "proguard"}},
		{"PublishTestersAddCmd", &PublishTestersAddCmd{}},
		{"PublishTestersRemoveCmd", &PublishTestersRemoveCmd{}},
		{"PublishTestersListCmd", &PublishTestersListCmd{}},
		{"PublishTestersGetCmd", &PublishTestersGetCmd{}},
		{"PublishBuildsListCmd", &PublishBuildsListCmd{}},
		{"PublishBuildsGetCmd", &PublishBuildsGetCmd{VersionCode: 1}},
		{"PublishBuildsExpireCmd", &PublishBuildsExpireCmd{VersionCode: 1}},
		{"PublishBuildsExpireAllCmd", &PublishBuildsExpireAllCmd{}},
		{"PublishBetaGroupsListCmd", &PublishBetaGroupsListCmd{}},
		{"PublishBetaGroupsGetCmd", &PublishBetaGroupsGetCmd{Group: "beta"}},
		{"PublishBetaGroupsCreateCmd", &PublishBetaGroupsCreateCmd{Group: "beta"}},
		{"PublishBetaGroupsUpdateCmd", &PublishBetaGroupsUpdateCmd{Group: "beta"}},
		{"PublishBetaGroupsDeleteCmd", &PublishBetaGroupsDeleteCmd{Group: "beta"}},
		{"PublishBetaGroupsAddTestersCmd", &PublishBetaGroupsAddTestersCmd{Group: "beta"}},
		{"PublishBetaGroupsRemoveTestersCmd", &PublishBetaGroupsRemoveTestersCmd{Group: "beta"}},
		{"PublishInternalShareUploadCmd", &PublishInternalShareUploadCmd{File: "test.aab"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				return // success is fine too
			}

			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestReviewsCommands_RequireAuth(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"ReviewsListCmd", &ReviewsListCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				t.Errorf("%s.Run() should return error without auth, got nil", tc.name)
				return
			}

			if !strings.Contains(err.Error(), "authentication not configured") {
				t.Errorf("%s.Run() error should contain 'authentication not configured', got: %v", tc.name, err)
			}
		})
	}
}

func TestReviewsCommands_RequireReviewID(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"ReviewsGetCmd", &ReviewsGetCmd{}},
		{"ReviewsReplyCmd", &ReviewsReplyCmd{}},
		{"ReviewsResponseGetCmd", &ReviewsResponseGetCmd{}},
		{"ReviewsResponseDeleteCmd", &ReviewsResponseDeleteCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				t.Errorf("%s.Run() should return error without review ID, got nil", tc.name)
				return
			}

			if !strings.Contains(err.Error(), "review ID is required") {
				t.Errorf("%s.Run() error should contain 'review ID is required', got: %v", tc.name, err)
			}
		})
	}
}

func TestVitalsCommands_Implemented(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	t.Run("VitalsCapabilitiesCmd", func(t *testing.T) {
		cmd := &VitalsCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("VitalsCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	t.Run("VitalsQueryCmd", func(t *testing.T) {
		cmd := &VitalsQueryCmd{}
		err := cmd.Run(globals)
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("VitalsQueryCmd.Run() should be implemented, but still returns 'not yet implemented'")
		}
	})
}

func TestVitalsCommands_RequireAuth(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"VitalsCrashesCmd", &VitalsCrashesCmd{}},
		{"VitalsAnrsCmd", &VitalsAnrsCmd{}},
		{"VitalsErrorsIssuesCmd", &VitalsErrorsIssuesCmd{}},
		{"VitalsErrorsReportsCmd", &VitalsErrorsReportsCmd{}},
		{"VitalsErrorsCountsGetCmd", &VitalsErrorsCountsGetCmd{}},
		{"VitalsErrorsCountsQueryCmd", &VitalsErrorsCountsQueryCmd{}},
		{"VitalsMetricsExcessiveWakeupsCmd", &VitalsMetricsExcessiveWakeupsCmd{}},
		{"VitalsMetricsSlowRenderingCmd", &VitalsMetricsSlowRenderingCmd{}},
		{"VitalsMetricsSlowStartCmd", &VitalsMetricsSlowStartCmd{}},
		{"VitalsMetricsStuckWakelocksCmd", &VitalsMetricsStuckWakelocksCmd{}},
		{"VitalsAnomaliesListCmd", &VitalsAnomaliesListCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				t.Errorf("%s.Run() should return error without auth, got nil", tc.name)
				return
			}

			if !strings.Contains(err.Error(), "authentication not configured") {
				t.Errorf("%s.Run() error should contain 'authentication not configured', got: %v", tc.name, err)
			}
		})
	}
}

func TestAnalyticsCommands_Implemented(t *testing.T) {
	t.Run("AnalyticsCapabilitiesCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &AnalyticsCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("AnalyticsCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	t.Run("AnalyticsQueryCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &AnalyticsQueryCmd{}
		err := cmd.Run(globals)
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("AnalyticsQueryCmd.Run() should be implemented, but still returns 'not yet implemented'")
		}
	})
}

func TestAppsCommands_Implemented(t *testing.T) {
	t.Run("AppsListCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &AppsListCmd{}
		err := cmd.Run(globals)
		// AppsListCmd returns informational output (no list-apps API), should succeed
		if err != nil {
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("AppsListCmd.Run() should be implemented, but still returns 'not yet implemented'")
			}
		}
	})

	t.Run("AppsGetCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &AppsGetCmd{Package: "com.example.app"}
		err := cmd.Run(globals)
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("AppsGetCmd.Run() should be implemented, but still returns 'not yet implemented'")
		}
	})
}

func TestGamesCommands_Implemented(t *testing.T) {
	t.Run("GamesCapabilitiesCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &GamesCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("GamesCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	// Commands that require auth — verify they don't return "not yet implemented"
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"GamesAchievementsResetCmd", &GamesAchievementsResetCmd{AchievementID: "test"}},
		{"GamesScoresResetCmd", &GamesScoresResetCmd{LeaderboardID: "test"}},
		{"GamesEventsResetCmd", &GamesEventsResetCmd{EventID: "test"}},
		{"GamesPlayersHideCmd", &GamesPlayersHideCmd{PlayerID: "player1", ApplicationID: "app1"}},
		{"GamesPlayersUnhideCmd", &GamesPlayersUnhideCmd{PlayerID: "player1", ApplicationID: "app1"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{}
			err := tc.cmd.Run(globals)
			if err == nil {
				return
			}
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestPurchasesCommands_Implemented(t *testing.T) {
	t.Run("PurchasesCapabilitiesCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &PurchasesCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("PurchasesCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"PurchasesProductsAcknowledgeCmd", &PurchasesProductsAcknowledgeCmd{ProductID: "p1", Token: "t1"}},
		{"PurchasesProductsConsumeCmd", &PurchasesProductsConsumeCmd{ProductID: "p1", Token: "t1"}},
		{"PurchasesSubscriptionsAcknowledgeCmd", &PurchasesSubscriptionsAcknowledgeCmd{SubscriptionID: "s1", Token: "t1"}},
		{"PurchasesSubscriptionsCancelCmd", &PurchasesSubscriptionsCancelCmd{SubscriptionID: "s1", Token: "t1"}},
		{"PurchasesSubscriptionsDeferCmd", &PurchasesSubscriptionsDeferCmd{SubscriptionID: "s1", Token: "t1", ExpectedExpiry: "2024-01-01", DesiredExpiry: "2024-02-01"}},
		{"PurchasesSubscriptionsRefundCmd", &PurchasesSubscriptionsRefundCmd{SubscriptionID: "s1", Token: "t1"}},
		{"PurchasesSubscriptionsRevokeCmd", &PurchasesSubscriptionsRevokeCmd{Token: "t1"}},
		{"PurchasesVerifyCmd", &PurchasesVerifyCmd{Token: "t1"}},
		{"PurchasesVoidedListCmd", &PurchasesVoidedListCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{}
			err := tc.cmd.Run(globals)
			if err == nil {
				return
			}
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestMonetizationCommands_Implemented(t *testing.T) {
	t.Run("MonetizationCapabilitiesCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &MonetizationCapabilitiesCmd{}
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("MonetizationCapabilitiesCmd.Run() should succeed, got: %v", err)
		}
	})

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"MonetizationProductsListCmd", &MonetizationProductsListCmd{}},
		{"MonetizationProductsGetCmd", &MonetizationProductsGetCmd{ProductID: "p1"}},
		{"MonetizationProductsCreateCmd", &MonetizationProductsCreateCmd{ProductID: "p1"}},
		{"MonetizationProductsUpdateCmd", &MonetizationProductsUpdateCmd{ProductID: "p1"}},
		{"MonetizationProductsDeleteCmd", &MonetizationProductsDeleteCmd{ProductID: "p1"}},
		{"MonetizationSubscriptionsListCmd", &MonetizationSubscriptionsListCmd{}},
		{"MonetizationSubscriptionsGetCmd", &MonetizationSubscriptionsGetCmd{SubscriptionID: "s1"}},
		{"MonetizationSubscriptionsCreateCmd", &MonetizationSubscriptionsCreateCmd{SubscriptionID: "s1", File: "sub.json"}},
		{"MonetizationSubscriptionsUpdateCmd", &MonetizationSubscriptionsUpdateCmd{SubscriptionID: "s1", File: "sub.json"}},
		{"MonetizationSubscriptionsPatchCmd", &MonetizationSubscriptionsPatchCmd{SubscriptionID: "s1", File: "sub.json"}},
		{"MonetizationSubscriptionsDeleteCmd", &MonetizationSubscriptionsDeleteCmd{SubscriptionID: "s1", Confirm: true}},
		{"MonetizationSubscriptionsArchiveCmd", &MonetizationSubscriptionsArchiveCmd{SubscriptionID: "s1"}},
		{"MonetizationSubscriptionsBatchGetCmd", &MonetizationSubscriptionsBatchGetCmd{IDs: []string{"s1"}}},
		{"MonetizationSubscriptionsBatchUpdateCmd", &MonetizationSubscriptionsBatchUpdateCmd{File: "batch.json"}},
		{"MonetizationOneTimeProductsListCmd", &MonetizationOneTimeProductsListCmd{}},
		{"MonetizationOneTimeProductsGetCmd", &MonetizationOneTimeProductsGetCmd{ProductID: "p1"}},
		{"MonetizationOneTimeProductsCreateCmd", &MonetizationOneTimeProductsCreateCmd{ProductID: "p1", File: "onetime.json"}},
		{"MonetizationOneTimeProductsUpdateCmd", &MonetizationOneTimeProductsUpdateCmd{ProductID: "p1", File: "onetime.json"}},
		{"MonetizationOneTimeProductsDeleteCmd", &MonetizationOneTimeProductsDeleteCmd{ProductID: "p1"}},
		{"MonetizationOneTimeProductsBatchGetCmd", &MonetizationOneTimeProductsBatchGetCmd{IDs: []string{"p1"}}},
		{"MonetizationOneTimeProductsBatchUpdateCmd", &MonetizationOneTimeProductsBatchUpdateCmd{File: "batch.json"}},
		{"MonetizationBasePlansActivateCmd", &MonetizationBasePlansActivateCmd{SubscriptionID: "s1", BasePlanID: "bp1"}},
		{"MonetizationBasePlansDeactivateCmd", &MonetizationBasePlansDeactivateCmd{SubscriptionID: "s1", BasePlanID: "bp1"}},
		{"MonetizationBasePlansDeleteCmd", &MonetizationBasePlansDeleteCmd{SubscriptionID: "s1", BasePlanID: "bp1", Confirm: true}},
		{"MonetizationBasePlansMigratePricesCmd", &MonetizationBasePlansMigratePricesCmd{SubscriptionID: "s1", BasePlanID: "bp1", RegionCode: "US", PriceMicros: 990000}},
		{"MonetizationBasePlansBatchMigrateCmd", &MonetizationBasePlansBatchMigrateCmd{SubscriptionID: "s1", File: "batch.json"}},
		{"MonetizationBasePlansBatchUpdateStatesCmd", &MonetizationBasePlansBatchUpdateStatesCmd{SubscriptionID: "s1", File: "batch.json"}},
		{"MonetizationOffersCreateCmd", &MonetizationOffersCreateCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferID: "o1", File: "offer.json"}},
		{"MonetizationOffersGetCmd", &MonetizationOffersGetCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferID: "o1"}},
		{"MonetizationOffersListCmd", &MonetizationOffersListCmd{SubscriptionID: "s1", BasePlanID: "bp1"}},
		{"MonetizationOffersDeleteCmd", &MonetizationOffersDeleteCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferID: "o1", Confirm: true}},
		{"MonetizationOffersActivateCmd", &MonetizationOffersActivateCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferID: "o1"}},
		{"MonetizationOffersDeactivateCmd", &MonetizationOffersDeactivateCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferID: "o1"}},
		{"MonetizationOffersBatchGetCmd", &MonetizationOffersBatchGetCmd{SubscriptionID: "s1", BasePlanID: "bp1", OfferIDs: []string{"o1"}}},
		{"MonetizationOffersBatchUpdateCmd", &MonetizationOffersBatchUpdateCmd{SubscriptionID: "s1", BasePlanID: "bp1", File: "batch.json"}},
		{"MonetizationOffersBatchUpdateStatesCmd", &MonetizationOffersBatchUpdateStatesCmd{SubscriptionID: "s1", BasePlanID: "bp1", File: "batch.json"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{}
			err := tc.cmd.Run(globals)
			if err == nil {
				return
			}
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestPermissionsCommands_Implemented(t *testing.T) {
	t.Run("PermissionsListCmd", func(t *testing.T) {
		globals := &Globals{}
		cmd := &PermissionsListCmd{}
		err := cmd.Run(globals)
		// PermissionsListCmd returns hardcoded permissions list, should succeed
		if err != nil {
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("PermissionsListCmd.Run() should be implemented, but still returns 'not yet implemented'")
			}
		}
	})

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"PermissionsUsersAddCmd", &PermissionsUsersAddCmd{Email: "test@example.com", Role: "admin"}},
		{"PermissionsUsersRemoveCmd", &PermissionsUsersRemoveCmd{Email: "test@example.com"}},
		{"PermissionsUsersListCmd", &PermissionsUsersListCmd{}},
		{"PermissionsGrantsAddCmd", &PermissionsGrantsAddCmd{Email: "test@example.com", Grant: "read"}},
		{"PermissionsGrantsRemoveCmd", &PermissionsGrantsRemoveCmd{Email: "test@example.com", Grant: "read"}},
		{"PermissionsGrantsListCmd", &PermissionsGrantsListCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{}
			err := tc.cmd.Run(globals)
			if err == nil {
				return
			}
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestRecoveryCommands_Implemented(t *testing.T) {
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"RecoveryListCmd", &RecoveryListCmd{}},
		{"RecoveryCreateCmd", &RecoveryCreateCmd{Type: "rollback", Reason: "test"}},
		{"RecoveryDeployCmd", &RecoveryDeployCmd{ID: "r1"}},
		{"RecoveryCancelCmd", &RecoveryCancelCmd{ID: "r1"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{}
			err := tc.cmd.Run(globals)
			if err == nil {
				return
			}
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but still returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestIntegrityCommands_Implemented(t *testing.T) {
	globals := &Globals{}

	cmd := &IntegrityDecodeCmd{Token: "test-token"}
	err := cmd.Run(globals)
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("IntegrityDecodeCmd.Run() should be implemented, but still returns 'not yet implemented'")
	}
}

// ============================================================================
// Test Command Field Tags
// ============================================================================

func TestPublishUploadCmd_FieldTags(t *testing.T) {
	cmd := PublishUploadCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		arg       string
		help      string
		enum      string
	}{
		{"File", "", "File to upload (APK or AAB)", ""},
		{"Track", "", "Target track", "internal,alpha,beta,production"},
		{"EditID", "", "Explicit edit transaction ID", ""},
		{"NoAutoCommit", "", "Keep edit open for manual commit", ""},
		{"DryRun", "", "Show intended actions without executing", ""},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("PublishUploadCmd missing field: %s", tc.fieldName)
			continue
		}

		if tc.help != "" {
			helpTag := field.Tag.Get("help")
			if helpTag == "" {
				t.Errorf("PublishUploadCmd.%s missing help tag", tc.fieldName)
			}
		}

		if tc.enum != "" {
			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("PublishUploadCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}
		}
	}
}

func TestPublishReleaseCmd_FieldTags(t *testing.T) {
	cmd := PublishReleaseCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
	}{
		{"Track", "internal,alpha,beta,production"},
		{"Status", "draft,completed,halted,inProgress"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("PublishReleaseCmd missing field: %s", tc.fieldName)
			continue
		}

		enumTag := field.Tag.Get("enum")
		if enumTag != tc.enum {
			t.Errorf("PublishReleaseCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
		}
	}
}

func TestPermissionsUsersAddCmd_FieldTags(t *testing.T) {
	cmd := PermissionsUsersAddCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Role")
	if !ok {
		t.Fatal("PermissionsUsersAddCmd missing Role field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "admin,developer,viewer"
	if enumTag != expected {
		t.Errorf("PermissionsUsersAddCmd.Role enum tag = %q, want %q", enumTag, expected)
	}

	requiredTag := field.Tag.Get("required")
	if requiredTag != "" {
		t.Errorf("PermissionsUsersAddCmd.Role required tag = %q, want empty string", requiredTag)
	}
}

func TestRecoveryCreateCmd_FieldTags(t *testing.T) {
	cmd := RecoveryCreateCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Type")
	if !ok {
		t.Fatal("RecoveryCreateCmd missing Type field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "rollback,emergency_update,version_hold"
	if enumTag != expected {
		t.Errorf("RecoveryCreateCmd.Type enum tag = %q, want %q", enumTag, expected)
	}
}

func TestGlobals_FieldTags(t *testing.T) {
	globals := Globals{}
	typeOfGlobals := reflect.TypeOf(globals)

	tests := []struct {
		fieldName string
		enum      string
	}{
		{"Output", "json,table,markdown,csv,excel"},
		{"StoreTokens", "auto,never,secure"},
	}

	for _, tc := range tests {
		field, ok := typeOfGlobals.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("Globals missing field: %s", tc.fieldName)
			continue
		}

		enumTag := field.Tag.Get("enum")
		if enumTag != tc.enum {
			t.Errorf("Globals.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
		}
	}
}

func TestAnalyticsQueryCmd_FieldTags(t *testing.T) {
	cmd := AnalyticsQueryCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Format")
	if !ok {
		t.Fatal("AnalyticsQueryCmd missing Format field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "json,csv"
	if enumTag != expected {
		t.Errorf("AnalyticsQueryCmd.Format enum tag = %q, want %q", enumTag, expected)
	}
}

func TestConfigCompletionCmd_FieldTags(t *testing.T) {
	cmd := ConfigCompletionCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Shell")
	if !ok {
		t.Fatal("ConfigCompletionCmd missing Shell field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "bash,zsh,fish"
	if enumTag != expected {
		t.Errorf("ConfigCompletionCmd.Shell enum tag = %q, want %q", enumTag, expected)
	}

	argTag := field.Tag.Get("arg")
	if argTag != "" {
		t.Errorf("ConfigCompletionCmd.Shell arg tag should be empty, got: %s", argTag)
	}
}

// ============================================================================
// Test Nested Command Structures
// ============================================================================

func TestVitalsErrorsCmd_NestedStructure(t *testing.T) {
	cmd := VitalsErrorsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Issues", "Reports", "Counts"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("VitalsErrorsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("VitalsErrorsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestVitalsMetricsCmd_NestedStructure(t *testing.T) {
	cmd := VitalsMetricsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"ExcessiveWakeups", "SlowRendering", "SlowStart", "StuckWakelocks",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("VitalsMetricsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("VitalsMetricsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestGamesAchievementsCmd_NestedStructure(t *testing.T) {
	cmd := GamesAchievementsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	field, ok := typeOfCmd.FieldByName("Reset")
	if !ok {
		t.Fatal("GamesAchievementsCmd missing Reset subcommand")
	}

	cmdTag := field.Tag.Get("cmd")
	if cmdTag != "" {
		t.Errorf("GamesAchievementsCmd.Reset should have cmd:\"\" tag")
	}
}

func TestPurchasesProductsCmd_NestedStructure(t *testing.T) {
	cmd := PurchasesProductsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Acknowledge", "Consume"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PurchasesProductsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PurchasesProductsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestMonetizationProductsCmd_NestedStructure(t *testing.T) {
	cmd := MonetizationProductsCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"List", "Get", "Create", "Update", "Delete"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("MonetizationProductsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("MonetizationProductsCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

func TestPermissionsUsersCmd_NestedStructure(t *testing.T) {
	cmd := PermissionsUsersCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Add", "Remove", "List"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PermissionsUsersCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PermissionsUsersCmd.%s should have cmd:\"\" tag", name)
		}
	}
}

// ============================================================================
// Test KongCLI Top-Level Structure
// ============================================================================

func TestKongCLI_HasExpectedTopLevelCommands(t *testing.T) {
	cli := KongCLI{}
	v := reflect.ValueOf(cli)
	typeOfCLI := v.Type()

	expectedCommands := []string{
		"Auth", "Config", "Publish", "Reviews", "Vitals", "Analytics",
		"Purchases", "Monetization", "Permissions", "Recovery", "Apps",
		"Games", "Integrity", "Migrate", "CustomApp", "Grouping", "Version",
	}

	for _, name := range expectedCommands {
		field, ok := typeOfCLI.FieldByName(name)
		if !ok {
			t.Errorf("KongCLI missing top-level command: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("KongCLI.%s should have cmd:\"\" tag", name)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("KongCLI.%s should have help tag", name)
		}
	}
}

func TestKongCLI_GlobalsStructure(t *testing.T) {
	cli := KongCLI{}
	typeOfCLI := reflect.TypeOf(cli)

	globalsField, ok := typeOfCLI.FieldByName("Globals")
	if !ok {
		t.Fatal("KongCLI missing Globals field")
	}

	if globalsField.Type.String() != "cli.Globals" {
		t.Errorf("KongCLI.Globals type = %s, want cli.Globals", globalsField.Type.String())
	}
}

// ============================================================================
// Test Auth Commands
// ============================================================================

func TestAuthCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := AuthCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Status", "Login", "Logout"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("AuthCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AuthCmd.%s should have cmd:\"\" tag", name)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("AuthCmd.%s should have help tag", name)
		}
	}
}

// ============================================================================
// Test Version Command
// ============================================================================

func TestVersionCmd_Run(t *testing.T) {
	globals := &Globals{}
	cmd := &VersionCmd{}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("VersionCmd.Run() returned error: %v", err)
	}
}

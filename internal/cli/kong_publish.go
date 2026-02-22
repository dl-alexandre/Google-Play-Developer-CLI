package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
)

// PublishCmd contains publishing commands.
type PublishCmd struct {
	Upload        PublishUploadCmd        `cmd:"" help:"Upload APK or AAB"`
	Release       PublishReleaseCmd       `cmd:"" help:"Create or update a release"`
	Rollout       PublishRolloutCmd       `cmd:"" help:"Update rollout percentage"`
	Promote       PublishPromoteCmd       `cmd:"" help:"Promote a release between tracks"`
	Halt          PublishHaltCmd          `cmd:"" help:"Halt a production rollout"`
	Rollback      PublishRollbackCmd      `cmd:"" help:"Rollback to a previous version"`
	Status        PublishStatusCmd        `cmd:"" help:"Get track status"`
	Tracks        PublishTracksCmd        `cmd:"" help:"List all tracks"`
	Capabilities  PublishCapabilitiesCmd  `cmd:"" help:"List publishing capabilities"`
	Listing       PublishListingCmd       `cmd:"" help:"Manage store listing"`
	Details       PublishDetailsCmd       `cmd:"" help:"Manage app details"`
	Images        PublishImagesCmd        `cmd:"" help:"Manage store images"`
	Assets        PublishAssetsCmd        `cmd:"" help:"Manage store assets"`
	Deobfuscation PublishDeobfuscationCmd `cmd:"" help:"Manage deobfuscation files"`
	Testers       PublishTestersCmd       `cmd:"" help:"Manage testers"`
	Builds        PublishBuildsCmd        `cmd:"" help:"Manage uploaded builds"`
	BetaGroups    PublishBetaGroupsCmd    `cmd:"" help:"Beta group management (ASC compatibility)"`
	InternalShare PublishInternalShareCmd `cmd:"" help:"Upload artifacts for internal sharing"`
}

// PublishUploadCmd uploads APK or AAB.
type PublishUploadCmd struct {
	File               string `arg:"" help:"File to upload (APK or AAB)" type:"existingfile"`
	Track              string `help:"Target track" default:"internal" enum:"internal,alpha,beta,production"`
	EditID             string `help:"Explicit edit transaction ID"`
	ObbMain            string `help:"Main expansion file path"`
	ObbPatch           string `help:"Patch expansion file path"`
	ObbMainRefVersion  int64  `help:"Reference version code for main expansion file"`
	ObbPatchRefVersion int64  `help:"Reference version code for patch expansion file"`
	NoAutoCommit       bool   `help:"Keep edit open for manual commit"`
	DryRun             bool   `help:"Show intended actions without executing"`
}

// Run executes the upload command.
func (cmd *PublishUploadCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish upload not yet implemented")
}

// PublishReleaseCmd creates or updates a release.
type PublishReleaseCmd struct {
	Track               string   `help:"Release track" default:"internal" enum:"internal,alpha,beta,production"`
	Name                string   `help:"Release name"`
	Status              string   `help:"Release status" default:"draft" enum:"draft,completed,halted,inProgress"`
	VersionCodes        []string `help:"Version codes to include (repeatable)"`
	RetainVersionCodes  []string `help:"Version codes to retain (repeatable)"`
	InAppUpdatePriority int      `help:"In-app update priority (0-5)" default:"-1"`
	ReleaseNotesFile    string   `help:"JSON file with localized release notes" type:"existingfile"`
	EditID              string   `help:"Explicit edit transaction ID"`
	NoAutoCommit        bool     `help:"Keep edit open for manual commit"`
	DryRun              bool     `help:"Show intended actions without executing"`
	Wait                bool     `help:"Wait for release to complete"`
	WaitTimeout         string   `help:"Maximum time to wait" default:"30m"`
}

// Run executes the release command.
func (cmd *PublishReleaseCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish release not yet implemented")
}

// PublishRolloutCmd updates rollout percentage.
type PublishRolloutCmd struct {
	Track        string  `help:"Release track" default:"production" enum:"internal,alpha,beta,production"`
	Percentage   float64 `help:"Rollout percentage (0.01-100.00)"`
	EditID       string  `help:"Explicit edit transaction ID"`
	NoAutoCommit bool    `help:"Keep edit open for manual commit"`
	DryRun       bool    `help:"Show intended actions without executing"`
}

// Run executes the rollout command.
func (cmd *PublishRolloutCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollout not yet implemented")
}

// PublishPromoteCmd promotes a release between tracks.
type PublishPromoteCmd struct {
	FromTrack    string  `help:"Source track"`
	ToTrack      string  `help:"Destination track"`
	Percentage   float64 `help:"Rollout percentage for destination" default:"0"`
	EditID       string  `help:"Explicit edit transaction ID"`
	NoAutoCommit bool    `help:"Keep edit open for manual commit"`
	DryRun       bool    `help:"Show intended actions without executing"`
}

// Run executes the promote command.
func (cmd *PublishPromoteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish promote not yet implemented")
}

// PublishHaltCmd halts a production rollout.
type PublishHaltCmd struct {
	Track        string `help:"Release track" default:"production" enum:"internal,alpha,beta,production"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the halt command.
func (cmd *PublishHaltCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish halt not yet implemented")
}

// PublishRollbackCmd rolls back to a previous version.
type PublishRollbackCmd struct {
	Track        string `help:"Release track"`
	VersionCode  string `help:"Specific version code to rollback to"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the rollback command.
func (cmd *PublishRollbackCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollback not yet implemented")
}

// PublishStatusCmd gets track status.
type PublishStatusCmd struct {
	Track string `help:"Release track (leave empty for all tracks)"`
}

// Run executes the status command.
func (cmd *PublishStatusCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish status not yet implemented")
}

// PublishTracksCmd lists all tracks.
type PublishTracksCmd struct{}

// Run executes the tracks command.
func (cmd *PublishTracksCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish tracks not yet implemented")
}

// PublishCapabilitiesCmd lists publishing capabilities.
type PublishCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *PublishCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish capabilities not yet implemented")
}

// PublishListingCmd manages store listing.
type PublishListingCmd struct {
	Update PublishListingUpdateCmd `cmd:"" help:"Update store listing"`
	Get    PublishListingGetCmd    `cmd:"" help:"Get store listing"`
	Delete PublishListingDeleteCmd `cmd:"" help:"Delete store listing"`
}

// PublishListingUpdateCmd updates store listing.
type PublishListingUpdateCmd struct {
	Locale       string `help:"Locale code" default:"en-US"`
	Title        string `help:"App title"`
	ShortDesc    string `help:"Short description"`
	FullDesc     string `help:"Full description"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the listing update command.
func (cmd *PublishListingUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing update not yet implemented")
}

// PublishListingGetCmd gets store listing.
type PublishListingGetCmd struct {
	Locale string `help:"Locale code (leave empty for all)"`
}

// Run executes the listing get command.
func (cmd *PublishListingGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing get not yet implemented")
}

// PublishListingDeleteCmd deletes store listing.
type PublishListingDeleteCmd struct {
	Locale       string `help:"Locale code (required)"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
	All          bool   `help:"Delete all store listings"`
}

// Run executes the listing delete command.
func (cmd *PublishListingDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing delete not yet implemented")
}

// PublishDetailsCmd manages app details.
type PublishDetailsCmd struct {
	Get    PublishDetailsGetCmd    `cmd:"" help:"Get app details"`
	Update PublishDetailsUpdateCmd `cmd:"" help:"Update app details"`
	Patch  PublishDetailsPatchCmd  `cmd:"" help:"Patch app details"`
}

// PublishDetailsGetCmd gets app details.
type PublishDetailsGetCmd struct{}

// Run executes the details get command.
func (cmd *PublishDetailsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details get not yet implemented")
}

// PublishDetailsUpdateCmd updates app details.
type PublishDetailsUpdateCmd struct {
	ContactEmail    string `help:"Contact email"`
	ContactPhone    string `help:"Contact phone"`
	ContactWebsite  string `help:"Contact website"`
	DefaultLanguage string `help:"Default language"`
	EditID          string `help:"Explicit edit transaction ID"`
	NoAutoCommit    bool   `help:"Keep edit open for manual commit"`
	DryRun          bool   `help:"Show intended actions without executing"`
}

// Run executes the details update command.
func (cmd *PublishDetailsUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details update not yet implemented")
}

// PublishDetailsPatchCmd patches app details.
type PublishDetailsPatchCmd struct {
	ContactEmail    string `help:"Contact email"`
	ContactPhone    string `help:"Contact phone"`
	ContactWebsite  string `help:"Contact website"`
	DefaultLanguage string `help:"Default language"`
	UpdateMask      string `help:"Fields to update (comma-separated)"`
	EditID          string `help:"Explicit edit transaction ID"`
	NoAutoCommit    bool   `help:"Keep edit open for manual commit"`
	DryRun          bool   `help:"Show intended actions without executing"`
}

// Run executes the details patch command.
func (cmd *PublishDetailsPatchCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details patch not yet implemented")
}

// PublishImagesCmd manages store images.
type PublishImagesCmd struct {
	Upload    PublishImagesUploadCmd    `cmd:"" help:"Upload an image"`
	List      PublishImagesListCmd      `cmd:"" help:"List images"`
	Delete    PublishImagesDeleteCmd    `cmd:"" help:"Delete an image"`
	DeleteAll PublishImagesDeleteAllCmd `cmd:"" help:"Delete all images for type"`
}

// PublishImagesUploadCmd uploads an image.
type PublishImagesUploadCmd struct {
	Type         string `arg:"" help:"Image type (icon, featureGraphic, phoneScreenshots, etc.)"`
	File         string `arg:"" help:"Image file path" type:"existingfile"`
	Locale       string `help:"Locale code" default:"en-US"`
	SyncImages   bool   `help:"Skip upload if identical image already exists"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images upload command.
func (cmd *PublishImagesUploadCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images upload not yet implemented")
}

// PublishImagesListCmd lists images.
type PublishImagesListCmd struct {
	Type   string `arg:"" help:"Image type (icon, featureGraphic, phoneScreenshots, etc.)"`
	Locale string `help:"Locale code" default:"en-US"`
	EditID string `help:"Explicit edit transaction ID"`
}

// Run executes the images list command.
func (cmd *PublishImagesListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images list not yet implemented")
}

// PublishImagesDeleteCmd deletes an image.
type PublishImagesDeleteCmd struct {
	Type         string `arg:"" help:"Image type"`
	ID           string `arg:"" help:"Image ID to delete"`
	Locale       string `help:"Locale code" default:"en-US"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images delete command.
func (cmd *PublishImagesDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images delete not yet implemented")
}

// PublishImagesDeleteAllCmd deletes all images for type.
type PublishImagesDeleteAllCmd struct {
	Type         string `arg:"" help:"Image type"`
	Locale       string `help:"Locale code" default:"en-US"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images deleteall command.
func (cmd *PublishImagesDeleteAllCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images deleteall not yet implemented")
}

// PublishAssetsCmd manages store assets.
type PublishAssetsCmd struct {
	Upload PublishAssetsUploadCmd `cmd:"" help:"Upload assets from directory"`
	Spec   PublishAssetsSpecCmd   `cmd:"" help:"Output asset validation matrix"`
}

// PublishAssetsUploadCmd uploads assets from directory.
type PublishAssetsUploadCmd struct {
	Dir          string `arg:"" help:"Assets directory" default:"assets"`
	Category     string `help:"Category to replace (phone, tablet, tv, wear)"`
	ReplaceAll   bool   `help:"Replace all existing assets"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the assets upload command.
func (cmd *PublishAssetsUploadCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets upload not yet implemented")
}

// PublishAssetsSpecCmd outputs asset validation matrix.
type PublishAssetsSpecCmd struct{}

// Run executes the assets spec command.
func (cmd *PublishAssetsSpecCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets spec not yet implemented")
}

// PublishDeobfuscationCmd manages deobfuscation files.
type PublishDeobfuscationCmd struct {
	Upload PublishDeobfuscationUploadCmd `cmd:"" help:"Upload deobfuscation file"`
}

// PublishDeobfuscationUploadCmd uploads deobfuscation file.
type PublishDeobfuscationUploadCmd struct {
	File         string `arg:"" help:"File to upload" type:"existingfile"`
	Type         string `help:"Deobfuscation file type: proguard or nativeCode" required:"" enum:"proguard,nativeCode"`
	VersionCode  int64  `help:"Version code to associate"`
	EditID       string `help:"Explicit edit transaction ID"`
	ChunkSize    int64  `help:"Upload chunk size in bytes" default:"10485760"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the deobfuscation upload command.
func (cmd *PublishDeobfuscationUploadCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish deobfuscation upload not yet implemented")
}

// PublishTestersCmd manages testers.
type PublishTestersCmd struct {
	Add    PublishTestersAddCmd    `cmd:"" help:"Add tester groups"`
	Remove PublishTestersRemoveCmd `cmd:"" help:"Remove tester groups"`
	List   PublishTestersListCmd   `cmd:"" help:"List tester groups"`
	Get    PublishTestersGetCmd    `cmd:"" help:"Get tester groups for a track"`
}

// PublishTestersAddCmd adds tester groups.
type PublishTestersAddCmd struct {
	Track        string   `help:"Track to add testers to" default:"internal"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the testers add command.
func (cmd *PublishTestersAddCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers add not yet implemented")
}

// PublishTestersRemoveCmd removes tester groups.
type PublishTestersRemoveCmd struct {
	Track        string   `help:"Track to remove testers from" default:"internal"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the testers remove command.
func (cmd *PublishTestersRemoveCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers remove not yet implemented")
}

// PublishTestersListCmd lists tester groups.
type PublishTestersListCmd struct {
	Track string `help:"Track to list testers for (empty for all)"`
}

// Run executes the testers list command.
func (cmd *PublishTestersListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers list not yet implemented")
}

// PublishTestersGetCmd gets tester groups for a track.
type PublishTestersGetCmd struct {
	Track string `help:"Track to get testers for (required)"`
}

// Run executes the testers get command.
func (cmd *PublishTestersGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers get not yet implemented")
}

// PublishBuildsCmd manages uploaded builds.
type PublishBuildsCmd struct {
	List      PublishBuildsListCmd      `cmd:"" help:"List uploaded builds"`
	Get       PublishBuildsGetCmd       `cmd:"" help:"Get build details"`
	Expire    PublishBuildsExpireCmd    `cmd:"" help:"Expire a build from tracks"`
	ExpireAll PublishBuildsExpireAllCmd `cmd:"" help:"Expire all builds from tracks"`
}

// PublishBuildsListCmd lists uploaded builds.
type PublishBuildsListCmd struct {
	Type   string `help:"Build type (apk, bundle, all)" default:"all"`
	EditID string `help:"Explicit edit transaction ID"`
}

// Run executes the builds list command.
func (cmd *PublishBuildsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds list not yet implemented")
}

// PublishBuildsGetCmd gets build details.
type PublishBuildsGetCmd struct {
	VersionCode int64  `arg:"" help:"Version code to get"`
	Type        string `help:"Build type (apk, bundle, all)" default:"all"`
	EditID      string `help:"Explicit edit transaction ID"`
}

// Run executes the builds get command.
func (cmd *PublishBuildsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds get not yet implemented")
}

// PublishBuildsExpireCmd expires a build from tracks.
type PublishBuildsExpireCmd struct {
	VersionCode  int64  `arg:"" help:"Version code to expire"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the builds expire command.
func (cmd *PublishBuildsExpireCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire not yet implemented")
}

// PublishBuildsExpireAllCmd expires all builds from tracks.
type PublishBuildsExpireAllCmd struct {
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the builds expire-all command.
func (cmd *PublishBuildsExpireAllCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire-all not yet implemented")
}

// PublishBetaGroupsCmd manages beta groups (ASC compatibility).
type PublishBetaGroupsCmd struct {
	List          PublishBetaGroupsListCmd          `cmd:"" help:"List beta groups"`
	Get           PublishBetaGroupsGetCmd           `cmd:"" help:"Get beta group details"`
	Create        PublishBetaGroupsCreateCmd        `cmd:"" help:"Create beta group"`
	Update        PublishBetaGroupsUpdateCmd        `cmd:"" help:"Update beta group testers"`
	Delete        PublishBetaGroupsDeleteCmd        `cmd:"" help:"Delete beta group"`
	AddTesters    PublishBetaGroupsAddTestersCmd    `cmd:"" help:"Add tester Google Groups to a beta group"`
	RemoveTesters PublishBetaGroupsRemoveTestersCmd `cmd:"" help:"Remove tester Google Groups from a beta group"`
}

// PublishBetaGroupsListCmd lists beta groups.
type PublishBetaGroupsListCmd struct {
	Track string `help:"Track to list (internal, alpha, beta). Empty lists all supported tracks"`
}

// Run executes the beta-groups list command.
func (cmd *PublishBetaGroupsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups list not yet implemented")
}

// PublishBetaGroupsGetCmd gets beta group details.
type PublishBetaGroupsGetCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups get command.
func (cmd *PublishBetaGroupsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups get not yet implemented")
}

// PublishBetaGroupsCreateCmd creates a beta group.
type PublishBetaGroupsCreateCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups create command.
func (cmd *PublishBetaGroupsCreateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups create not yet implemented")
}

// PublishBetaGroupsUpdateCmd updates beta group testers.
type PublishBetaGroupsUpdateCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to add"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups update command.
func (cmd *PublishBetaGroupsUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups update not yet implemented")
}

// PublishBetaGroupsDeleteCmd deletes a beta group.
type PublishBetaGroupsDeleteCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups delete command.
func (cmd *PublishBetaGroupsDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups delete not yet implemented")
}

// PublishBetaGroupsAddTestersCmd adds tester Google Groups to a beta group.
type PublishBetaGroupsAddTestersCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to add"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups add-testers command.
func (cmd *PublishBetaGroupsAddTestersCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups add-testers not yet implemented")
}

// PublishBetaGroupsRemoveTestersCmd removes tester Google Groups from a beta group.
type PublishBetaGroupsRemoveTestersCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to remove"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups remove-testers command.
func (cmd *PublishBetaGroupsRemoveTestersCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups remove-testers not yet implemented")
}

// PublishInternalShareCmd uploads artifacts for internal sharing.
type PublishInternalShareCmd struct {
	Upload PublishInternalShareUploadCmd `cmd:"" help:"Upload artifact for internal sharing"`
}

// PublishInternalShareUploadCmd uploads artifact for internal sharing.
type PublishInternalShareUploadCmd struct {
	File   string `arg:"" help:"File to upload (APK or AAB)" type:"existingfile"`
	DryRun bool   `help:"Show intended actions without executing"`
}

// Run executes the internal-share upload command.
func (cmd *PublishInternalShareUploadCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish internal-share upload not yet implemented")
}

// outputFormat returns the output format string.
func outputFormat(format string) string {
	return format
}

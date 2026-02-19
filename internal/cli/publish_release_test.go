package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dl-alexandre/gpd/internal/output"
)

func TestPublishCapabilitiesWorkflowMappings(t *testing.T) {
	cli := New()

	buf := &bytes.Buffer{}
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)

	cli.rootCmd.SetArgs([]string{"publish", "capabilities"})
	if err := cli.rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	data, ok := envelope["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", envelope["data"])
	}

	workflowMappings, ok := data["workflowMappings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected workflowMappings object, got %T", data["workflowMappings"])
	}

	asc, ok := workflowMappings["asc"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected workflowMappings.asc object, got %T", workflowMappings["asc"])
	}

	submit, ok := asc["submit"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected workflowMappings.asc.submit object, got %T", asc["submit"])
	}

	if submit["create"] != "gpd publish release" {
		t.Errorf("expected submit.create mapping, got %v", submit["create"])
	}
	if submit["status"] != "gpd publish status" {
		t.Errorf("expected submit.status mapping, got %v", submit["status"])
	}
	if submit["cancel"] != "gpd publish halt" {
		t.Errorf("expected submit.cancel mapping, got %v", submit["cancel"])
	}
}

func TestPublishReleaseHelpTextMapping(t *testing.T) {
	cli := New()
	publishCmd := requireCommand(t, cli.rootCmd, "publish")
	releaseCmd := requireCommand(t, publishCmd, "release")

	if releaseCmd.Long == "" {
		t.Fatal("expected release command to have Long help text")
	}
	if !strings.Contains(releaseCmd.Long, "ASC") {
		t.Fatalf("expected release help text to mention ASC mapping, got: %s", releaseCmd.Long)
	}
}

//go:build unit
// +build unit

package playship

import "testing"

func TestResolveReleaseParams(t *testing.T) {
	st, frac, err := ResolveReleaseParams("draft", 0)
	if err != nil || st != "draft" || frac != 0 {
		t.Fatalf("draft: %s %v %v", st, frac, err)
	}
	st, frac, err = ResolveReleaseParams("completed", 10)
	if err != nil || st != "inProgress" || frac != 0.1 {
		t.Fatalf("staged: %s %v %v", st, frac, err)
	}
	if _, _, err := ResolveReleaseParams("nope", 0); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestBuildTrackRelease(t *testing.T) {
	tr := BuildTrackRelease("production", "completed", []int64{7}, 0)
	if tr.Track != "production" || tr.Releases[0].Status != "completed" {
		t.Fatalf("%+v", tr)
	}
	st := BuildTrackRelease("production", "inProgress", []int64{7}, 0.25)
	if st.Releases[0].UserFraction != 0.25 {
		t.Fatalf("fraction=%v", st.Releases[0].UserFraction)
	}
}

func TestBuildPlan(t *testing.T) {
	plan := BuildPlan("com.ex", "a.aab", "internal", "draft", 0, 0)
	if plan["releaseStatus"] != "draft" {
		t.Fatalf("%v", plan["releaseStatus"])
	}
	steps := plan["steps"].([]map[string]interface{})
	found := false
	for _, s := range steps {
		if s["action"] == "release" {
			found = true
			if s["assignsArtifactToTrack"] != true {
				t.Fatal("missing assign flag")
			}
		}
	}
	if !found {
		t.Fatal("no release step")
	}
}

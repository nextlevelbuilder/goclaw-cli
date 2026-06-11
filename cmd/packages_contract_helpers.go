package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

func packageSpecFromRuntime(name, runtime string) (string, error) {
	name = strings.TrimSpace(name)
	runtime = strings.ToLower(strings.TrimSpace(runtime))
	if name == "" {
		return "", fmt.Errorf("package name is required")
	}
	if runtime == "" || strings.Contains(name, ":") {
		return name, nil
	}
	switch runtime {
	case "python", "python3", "pip":
		return "pip:" + name, nil
	case "node", "nodejs", "npm":
		return "npm:" + name, nil
	default:
		return "", fmt.Errorf("unsupported runtime %q; use python, node, or an explicit package spec", runtime)
	}
}

func githubReleasesPath(repo string, limit int) (string, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "", fmt.Errorf("--repo is required")
	}
	q := url.Values{}
	q.Set("repo", repo)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	return "/v1/packages/github-releases?" + q.Encode(), nil
}

func addInstalledPackageRows(tbl *output.TableData, data json.RawMessage) {
	if list := unmarshalList(data); len(list) > 0 {
		for _, pkg := range list {
			tbl.AddRow(str(pkg, "runtime"), packageName(pkg), str(pkg, "version"), str(pkg, "status"))
		}
		return
	}
	payload := unmarshalMap(data)
	for _, source := range []string{"system", "pip", "npm", "github"} {
		for _, pkg := range listFromValue(payload[source]) {
			tbl.AddRow(source, packageName(pkg), str(pkg, "version"), packageDetail(pkg))
		}
	}
}

func packageName(pkg map[string]any) string {
	if name := str(pkg, "name"); name != "" {
		return name
	}
	return str(pkg, "binary")
}

func packageDetail(pkg map[string]any) string {
	parts := []string{}
	if repo := str(pkg, "repo"); repo != "" {
		parts = append(parts, "repo="+repo)
	}
	if tag := str(pkg, "tag"); tag != "" {
		parts = append(parts, "tag="+tag)
	}
	if binaries := strings.Join(stringListFromValue(pkg["binaries"]), ","); binaries != "" {
		parts = append(parts, "binaries="+binaries)
	}
	if status := str(pkg, "status"); status != "" {
		parts = append(parts, "status="+status)
	}
	return strings.Join(parts, " ")
}

func addRuntimeRows(tbl *output.TableData, data json.RawMessage) {
	if list := unmarshalList(data); len(list) > 0 {
		for _, rt := range list {
			tbl.AddRow(str(rt, "name"), str(rt, "available"), str(rt, "version"), "")
		}
		return
	}
	payload := unmarshalMap(data)
	ready := str(payload, "ready")
	for _, rt := range listFromValue(payload["runtimes"]) {
		tbl.AddRow(str(rt, "name"), str(rt, "available"), str(rt, "version"), ready)
	}
}

func addDenyGroupRows(tbl *output.TableData, groups []map[string]any) {
	for _, group := range groups {
		tbl.AddRow(str(group, "name"), str(group, "description"), str(group, "default"))
	}
}

func addGitHubReleaseRows(tbl *output.TableData, releases []map[string]any) {
	for _, rel := range releases {
		tbl.AddRow(str(rel, "tag"), str(rel, "name"), str(rel, "prerelease"), releaseAssets(rel))
	}
}

func releaseAssets(rel map[string]any) string {
	for _, key := range []string{"matching_assets", "assets"} {
		if assets := strings.Join(stringListFromValue(rel[key]), ","); assets != "" {
			return assets
		}
	}
	return str(rel, "asset_count")
}

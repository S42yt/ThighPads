package main

import (
	"time"
)

const (
	appVersion        = "1.0.7"
	releasesURL       = "https://api.github.com/repos/s42yt/thighpads/releases/latest"
	updateCheckPeriod = 7 * 24 * time.Hour
)

type GithubRelease struct {
	TagName    string  `json:"tag_name"`
	Assets     []Asset `json:"assets"`
	PreRelease bool    `json:"prerelease"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

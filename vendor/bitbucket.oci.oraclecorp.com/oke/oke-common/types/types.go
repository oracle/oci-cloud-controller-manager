//go:generate make protobuf

package types

import "fmt"

// BuildInfoV1 is the API equivalent of gRPC BuildInfo
type BuildInfoV1 struct {
	GitCommit   string `json:"gitCommit"`
	Version     string `json:"version"`
	Release     string `json:"release"`
	BuildHost   string `json:"buildHost"`
	BuildDate   string `json:"buildDate"`
	BuildBranch string `json:"buildBranch"`
}

func (b BuildInfoV1) String() string {
	return fmt.Sprintf("%s-%s-%s", b.Version, b.Release, b.GitCommit)
}

func (b BuildInfo) ToV1() BuildInfoV1 {
	return BuildInfoV1(b)
}

func (b BuildInfoV1) ToProto() BuildInfo {
	return BuildInfo(b)
}

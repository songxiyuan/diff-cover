package merge

import (
	"testing"
)

var (
	repository = "https://github.com/songxiyuan/diff-cover-test.git"
	commitId1  = "585552d55c881d0e544deabf044ce096c9a4f01b"
	commitId2  = "1e708b63b1f8275c78572facdd52199dda288966"
)

func TestDiffCoverMerge(t *testing.T) {
	type args struct {
		cc1     CommitCover
		cc2     CommitCover
		tempDir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DiffCoverMerge merge test",
			args: args{
				cc1: CommitCover{
					Repository:    repository,
					Branch:        "init",
					CommitId:      commitId1,
					CoverFilePath: "./test/cover_1_4.out",
				},
				cc2: CommitCover{
					Repository:    repository,
					Branch:        "init",
					CommitId:      commitId2,
					CoverFilePath: "./test/cover_2_3.out",
				},
				tempDir: "tmp/",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := DiffCoverMerge(tt.args.cc1, tt.args.cc2, tt.args.tempDir); (err != nil) != tt.wantErr {
				t.Errorf("DiffCoverMerge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

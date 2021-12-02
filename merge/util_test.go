package merge

import "testing"

func TestGetModule(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name       string
		args       args
		wantModule string
	}{
		{
			name:       "",
			args:       args{
				content:
					`module test/module 

go 1.13

require (
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go v1.1.7 // indirect
	github.com/yuin/gopher-lua v0.0.0-20190514113301-1cd887cd7036 // indirect
	go.intra.xiaojukeji.com/apollo/apollo-golang-sdk-v2 v2.7.7+incompatible
	go.intra.xiaojukeji.com/foundation/didi-standard-lib v0.0.0-20201123053335-e67edadc52bf // indirect
	go.intra.xiaojukeji.com/platform-ha/onekey-degrade_sdk_go v3.2.8+incompatible
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/ini.v1 v1.52.0 // indirect
)
`,
			},
			wantModule: "test/module",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotModule := GetModule(tt.args.content); gotModule != tt.wantModule {
				t.Errorf("GetModule() = %v, want %v", gotModule, tt.wantModule)
			}
		})
	}
}

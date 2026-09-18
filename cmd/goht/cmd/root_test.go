package cmd

import "testing"

func TestExecute(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
	}{
		"root command succeeds":         {},
		"unknown command returns error": {args: []string{"unknown"}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			t.Cleanup(func() { rootCmd.SetArgs(nil) })

			err := Execute()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}

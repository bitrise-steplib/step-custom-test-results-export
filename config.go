package main

// Config holds the step inputs
type Config struct {
	TestName       string `env:"test_name,required"`
	BasePath       string `env:"base_path,required"`
	Pattern        string `env:"pattern,required"`
	TestResultsDir string `env:"bitrise_test_result_dir,dir"`
	VerboseLog     bool   `env:"verbose_log,opt[no,yes]"`
}

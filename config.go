package main

// Config holds the step inputs
type Config struct {
	TestName       string `env:"test_name,required"`
	SearchPath     string `env:"search_path,required"`
	TestResultsDir string `env:"bitrise_test_result_dir,dir"`
	VerboseLog     bool   `env:"verbose_log,opt[no,yes]"`
}

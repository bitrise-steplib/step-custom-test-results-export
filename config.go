package main

// config holds the step inputs
type config struct {
	TestName       string `env:"test_name,required"`
	BasePath       string `env:"base_path,required"`
	SearchPattern  string `env:"search_pattern,required"`
	TestResultsDir string `env:"bitrise_test_result_dir,dir"`
	VerboseLog     bool   `env:"verbose_log,opt[no,yes]"`
}
